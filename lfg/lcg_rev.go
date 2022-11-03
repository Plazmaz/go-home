package lfg

import (
	"errors"
	"fmt"
	"log"
	"math"
)

func modInverse(a, b float64) (float64, float64, float64) {
	// Based on http://anh.cs.luc.edu/331/notes/xgcd.pdf
	oldX := float64(1)
	oldY := float64(0)
	x := float64(0)
	y := float64(1)
	quotient := float64(0)

	for b != 0 {
		quotient = math.Floor(a / b)
		x, oldX = oldX-quotient*x, x
		y, oldY = oldY-quotient*y, y
		a, b = b, math.Mod(a, b)
	}
	return a, oldX, oldY
}

// CalcLCGSeed takes a known multiplier a, modulus m, output x, and count n,
// then finds x(0) for sequence x(n) = a*x(n-1) % m. This means it finds the seed.
func CalcLCGSeed(multiplier, modulus, curVal int64, count int) int64 {
	// 2147483647 * 5567 + 48271 * -247665088
	// mod m is 0 * 5567 + 48271 * (-247665088 mod m)
	// So our step is (-247665088 mod m)
	_, _, inverseMultiplier := modInverse(float64(modulus), float64(multiplier))
	step := int64(inverseMultiplier) % modulus
	outVal := curVal

	for i := int64(0); i < int64(count); i++ {
		outVal = outVal * step % modulus
	}

	return outVal
}

// ReverseLCG takes a value before and after being transformed by a LCG
// and returns the seed of that LCG
func ReverseLCG(originalValue, transformedValue int64) (int64, error) {
	a := int64(48271)
	modulus := int64((1 << 31) - 1)
	// First, we get the LCG outputs used to modify the two values
	valOne, valTwo, err := crackRands(originalValue, transformedValue)
	if err != nil {
		return -1, err
	}

	log.Printf("%d == %d * %d %% %d ? %t\n", valTwo, valOne, a, modulus, (valOne*a)%modulus == valTwo)
	log.Printf("In: %d %b\n", valTwo, valTwo)
	// Then we calculate the seed using the Extended Euclidean algorithm.
	seed := CalcLCGSeed(a, modulus, int64(valTwo), 1022)
	log.Printf("Seed: %d (mod %d)\n", seed, modulus)
	return seed, nil
}

// crackRands takes a state value before and after transformation by a value
// and returns the values used to transform the seeds.
func crackRands(originalStateValue, stateValue int64) (int64, int64, error) {
	x2Candidate, x3Candidate, x2Output, x2LastBits := int64(0), int64(0), int64(0), int64(0)
	modulus := int64((1 << 31) - 1)
	// XOR for negatives is undefined
	stateValue = int64(uint64(stateValue) ^ uint64(originalStateValue))
	// We can get the last 20 bits for x3
	x3LastBits := stateValue & ((1 << 20) - 1)
	// Now we need to bruteforce the 12 bits we're missing
	// xorTarget := (stateValue >> 20) & ((1 << 12) - 1)
	// fmt.Printf("xorTarget = %012b\n", xorTarget)

	x2BitsMask := int64(((1 << 20) - 1) << 20)
	// We know 0 bits were either both 1s or both 0s,
	// and 1 bits were either 1, 0 or 0, 1
	// However it may just be more efficient to bruteforce it naively.
	for i := int64(1); i <= 4096; i++ {
		x3Candidate = x3LastBits | i<<20
		// This is the int where we'll store our x2 values
		x2Output = stateValue ^ x3Candidate
		// Returns the last 22 bits of x2
		x2LastBits = (x2Output & x2BitsMask) >> 20
		for j := int64(1); j <= 4096; j++ {
			x2Candidate = x2LastBits | j<<20
			if (x2Candidate*48271)%modulus == x3Candidate {
				fmt.Printf("Found! x2: (%d)%032b, x3: (%d)%032b\n", x2Candidate, x2Candidate, x3Candidate, x3Candidate)
				return x2Candidate, x3Candidate, nil
			}
		}
	}
	// Not found :(
	return -1, -1, errors.New("value space exhausted, invalid sequence")

}
