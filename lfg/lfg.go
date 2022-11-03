package lfg

import (
	"fmt"

	"../rng"
)

func indexOf(val int64, check []int64) int {
	for idx, b := range check {
		if b == val {
			return idx
		}
	}
	return -1
}

// getFeedIndex offsets the current tap index to get the feed value.
// The feed value is `j` values in the past. For golang's default source, `j = 273`.
// curIndex is the currently modified index, a sum of S[getFeedIndex(curIndex - 1)] and S[curIndex - 1]
func getFeedIndex(curIndex int) int {
	var idx = (rng.RngLen - rng.RngTap) - curIndex - 1
	if idx < 0 {
		idx += rng.RngLen
	}
	return idx
}

// StepState insert newVal into the state
func StepState(curIdx int, state *[rng.RngLen]int64, newVal int64) bool {
	// This simply offsets
	var feedIdx = getFeedIndex(curIdx)
	if state[feedIdx] != 0 {
		fmt.Println("Found old feed value! Current Index: ", curIdx,
			" Old Index: ", feedIdx)
		var tapVal = newVal - state[feedIdx]
		var tapIdx = indexOf(tapVal, state[:])
		if tapIdx > 0 {
			fmt.Println("Found known tap index, state finished!")
			return true
		}
	}
	// We've updated our feed now regardless of if we know the value
	state[feedIdx] = newVal
	return false
}

// StepState16 reconstructs partial state if only 32 bit integers are used
func StepState16(curIdx int, state *[rng.RngLen]int64, newVal int16) bool {
	return StepState(curIdx, state, int64(newVal)<<48)
}

// StepState32 reconstructs partial state if only 32 bit integers are used
func StepState32(curIdx int, state *[rng.RngLen]int64, newVal int32) bool {
	return StepState(curIdx, state, int64(newVal)<<32)
}

// RollForwards runs the LFG until the expected value is produced.
func RollForwards(src *rng.RngSource, knownState [rng.RngLen]int64,
	expectedIdx int, expected int64) ([rng.RngLen]int64, int) {
	var curIdx = expectedIdx
	var success = false
	for i := 0; i <= 5000; i++ {
		var val = int64(src.Uint64())
		if val == expected {
			fmt.Printf("Found matching state after %d iterations. Breaking.\n", i)
			success = true
			break
			// If our lower 32 bits have been cleared, try a partial match on expected.
		} else if val&0000000011111111 == 0 && val == ((expected>>32)<<32) {
			fmt.Printf("Found 32 bit matching state at %d iterations. Breaking.\n", i)
			success = true
			break
		}

	}
	if !success {
		fmt.Printf("Exhausted search after %d iterations.\n", 5000)
	}
	return knownState, curIdx
}

// RollBackwards runs the LFG backwards by the given steps amount.
func RollBackwards(src *rng.RngSource, steps int64) {
	const rngLen = 607
	// Tap = cur output
	var curTapIdx = src.Tap
	var curFeedIdx = src.Feed
	if curTapIdx > rngLen {
		curTapIdx -= rngLen
	}
	if curFeedIdx > rngLen {
		curFeedIdx -= rngLen
	}

	var lastFeed = int64(-1)
	for i := int64(0); i < steps; i++ {
		curTap := src.Vec[curTapIdx]
		curFeed := src.Vec[curFeedIdx]
		lastFeed = int64(curFeed - curTap)
		src.Vec[curFeedIdx] = lastFeed
		// The tap and feed decrease when generating forward, then loop
		// from 0 to rngLen. This does the opposite.
		curTapIdx++
		curFeedIdx++
		if curTapIdx >= rngLen {
			curTapIdx -= rngLen
		}
		if curFeedIdx >= rngLen {
			curFeedIdx -= rngLen
		}

	}
	src.Tap = curTapIdx
	src.Feed = curFeedIdx

}

// ApplyState takes the RngSource provied, and sets its state to the state value provided.
// curIdx is used to set the tap/feed, and should be the latest feed index.
func ApplyState(src *rng.RngSource, state [rng.RngLen]int64, curIdx int) {
	if curIdx > rng.RngLen {
		curIdx -= rng.RngLen
	}
	src.Vec = state
	src.Tap = curIdx
	src.Feed = getFeedIndex(curIdx - 1)
}

// UpdateState takes an existing source value,
// then updates the tap/feed and inserts newVal into the state.
func UpdateState(src *rng.RngSource, newVal int64) {
	var curIdx = src.Tap
	var feedIdx = getFeedIndex(curIdx - 1)
	if src.Vec[feedIdx]+src.Vec[curIdx] != newVal {
		fmt.Printf("Invalid state detected (%d != %d), indicates we may have missed data...\n",
			src.Vec[feedIdx]+src.Vec[curIdx], newVal)
	} else {
		fmt.Println("We're already there!")
	}

	// Roll forwards until we reproduce the intended output.
	RollForwards(src, src.Vec, curIdx, newVal)
}
