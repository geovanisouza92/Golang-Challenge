# Golang-Challenge

Challenge test

We ask that you complete the following challenge to evaluate your development skills.

## The Challenge

Finish the implementation of the provided Transparent Cache package.

## Show your work

1. Create a **Private** repository and share it with the recruiter ( please dont make a pull request, clone the private repository and create a new private one on your profile)
2. Commit each step of your process so we can follow your thought process.
3. Give your interviewer access to the private repo

## What to build

Take a look at the current TransparentCache implementation.

You'll see some "TODO" items in the project for features that are still missing.

The solution can be implemented either in Golang or Java ( but you must be able to read code in Golang to realize the exercise ).

Also, you'll see that some of the provided tests are failing because of that.

The following is expected for solving the challenge:

* Design and implement the missing features in the cache
* Make the failing tests pass, trying to make none (or minimal) changes to them
* Add more tests if needed to show that your implementation really works

## Deliverables we expect:

* Your code in a private Github repo
* README file with the decisions taken and important notes

## Time Spent

You need to fully complete the challenge. We suggest not spending more than 3 days total. Please make commits as often as possible so we can see the time you spent and please do not make one commit. We will evaluate the code and time spent.

What we want to see is how well you handle yourself given the time you spend on the problem, how you think, and how you prioritize when time is insufficient to solve everything.

Please email your solution as soon as you have completed the challenge or the time is up.

## About my design decisions

First I felt the need for some kind of time marking when a price was cached. I preferred to keep the time close to the price instead having another data structure holding that information. So I created a `price` data structure with the price value and the time when that value was got.

Then I've made the minimal changes on the `GetPriceFor` method to keep track of the time, calculating the age of the information and placing new prices on internal cache.

When I was looking at the second _TODO_, I thought about the risk of making multiple _goroutines_ accessing the same data variables and instead implementing some complicated _mutex_ locking, I prefered to use the standard library (in this case, `sync.Map`) first, making sure that change would not break the passing tests.

After that, I directed my attention to the concurrent implementation of prices querying. I knew it would involve _goroutines_ and some king of synchronization, so I implemented first the spread of _goroutines_, one for each CPU on the computer because I thought it would be conservative enough. Maybe I could launch one per `itemCode`, but I thought it was over-engineering given that the requirements didn't specified anything about it.

I used a `sync.WaitGroup` to synchronize with the main _goroutine_ and channels to communicate back and forth. As I understood that the implementation should be a all-or-nothing operation, I used another _goroutine_ and channel to aggregate the results of the operation and used a `select` to either return the results or the first error found.

Besides it's not mentioned, I used two additional structs to hold the position of each `itemCode` in `itemCodes` parameter to ensure every result would be in the same order, as the synchronous version would do.
