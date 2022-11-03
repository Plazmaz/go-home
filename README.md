# Go Home

Code for cracking the seed of Go's default `random.Source`. This is very messy, and likely won't be maintained much, but I will accept PRs and feedback.

# Layout

| Folder               | Description                                                                                                                     |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------|
|notes_and_research    | "scratch pad" notes and thought dumps, full of dead ends and learning experiences. Probably not very legible.                                                                                                                       |
|rng                   | Contains a modified version of go's "rng" class for convenience methods and simplicity (⚠ not used to "cheat" and get state, only used so that I can put the recovered state into something with a nice API and the same defaults⚠)|
|lfg                   | Utilities for state recovery of the lagged Fibonacci generator and reversing/cracking the seed of the multiplicative linear congruential generator                                                                                  | 

Writeup here:  
<insert link>
