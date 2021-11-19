package sample1

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             sync.Map
}

type price struct {
	value float64
	got   time.Time
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		prices:             sync.Map{},
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	p, ok := c.prices.Load(itemCode)
	if ok && time.Since(p.(price).got) < c.maxAge {
		return p.(price).value, nil
	}
	value, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	p = price{value, time.Now()}
	c.prices.Store(itemCode, p)
	return p.(price).value, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	// req represents a price request
	type req struct {
		pos      int // position of the itemCode in itemCodes
		itemCode string
	}
	// res represents a price result
	type res struct {
		pos   int // position of the itemCode in itemCodes
		price float64
	}

	// Separate channels for price requests, results and the error

	reqChan := make(chan req, len(itemCodes)/runtime.NumCPU())
	resChan := make(chan res, len(itemCodes)/runtime.NumCPU())

	// Use a wait group to coordinate workers running

	var wg errgroup.Group
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Go(func() error {
			for req := range reqChan {
				price, err := c.GetPriceFor(req.itemCode)
				if err != nil {
					return err
				}
				resChan <- res{req.pos, price}
			}
			return nil
		})
	}

	// Send each item code to workers using the reqChan, then close it.

	for i, itemCode := range itemCodes {
		reqChan <- req{i, itemCode}
	}
	close(reqChan)

	// Use another goroutine to aggregate the results inside a single slice

	resultsChan := make(chan []float64)
	go func() {
		defer close(resultsChan)

		results := make([]float64, len(itemCodes))
		for res := range resChan {
			results[res.pos] = res.price
		}
		resultsChan <- results
	}()

	err := wg.Wait()
	close(resChan)
	if err != nil {
		return []float64{}, err
	}
	return <-resultsChan, nil
}
