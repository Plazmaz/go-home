package main

import (
	"fmt"
	"log"

	"./lfg"
	"./rng"
)

func main() {
	var src = rng.RngSource{}
	src.Seed(1234)

	fmt.Println("Testing initial state application...")
	var state [rng.RngLen]int64
	var curIdx = 0
	for curIdx := 0; curIdx < rng.RngLen; curIdx++ {
		if lfg.StepState(curIdx, &state, int64(src.Uint64())) {
			break
		}
	}

	var newRand = rng.RngSource{}
	lfg.ApplyState(&newRand, state, curIdx)
	stepsToTake := 608
	for i := 0; i < stepsToTake; i++ {
		newRand.Uint64()
	}

	seed, err := RecoverSeed(newRand, int64(stepsToTake)+rng.RngLen)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(int32(seed))
}

type LCGCheckResult struct {
	seed int64
	err  error
}

func RecoverSeed(randWithState rng.RngSource, steps int64) (int64, error) {

	// Since we know the original value (before transformation), we can use that and the modified value
	// (after seed) to first derive the value used to modify state, then rolling that all the way back
	// to the real seed.
	lfg.RollBackwards(&randWithState, steps)
	seed, err := lfg.ReverseLCG(rng.RngCooked[randWithState.Feed-1],
		randWithState.Vec[randWithState.Feed-1])
	if err != nil {
		return -1, fmt.Errorf("unable to find seed after %d steps: %w", steps, err)
	}
	log.Println("Rolled back tap value: ", randWithState.Tap, randWithState.Vec[randWithState.Tap])
	log.Println("Rolled back feed value: ", randWithState.Feed, randWithState.Vec[randWithState.Feed])
	return seed, nil
}
