package store

/*
 * This file is intended to generate codes in the following manner:
 * 1) Every possible code should be generated exactly once until exhausted
 * 2) Codes should be non-sequential and not trivially predictable
 * 3) Codes should be roughly uniformly distributed across the code space (as random-looking as possible)
 *
 * A Feistel network is the best way to achieve this.
 */

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync/atomic"
)

// codeAlphabet excludes visually ambiguous characters: 0/O, 1/I
// Length is 32, so byte % len(codeAlphabet) is unbiased because 256 % 32 == 0
const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// codeGenerator is used to generate unique, non-sequential codes for sessions.
type codeGenerator struct {
	size      int    // length of the code in characters
	bits      uint   // number of bits needed to represent all possible codes (size * 5)
	split     uint   // number of bits in the right half of the Feistel network
	mask      uint64 // bitmask for the entire code space
	rightMask uint64 // bitmask for the right half
	leftMask  uint64 // bitmask for the left half
	roundKeys [3]uint64
	counter   uint64
}

// newCodeGenerator creates a new code generator for the specified code size, which determines how many unique codes can be generated before collisions occur.
func newCodeGenerator(size int) (*codeGenerator, error) {
	if size <= 0 {
		return nil, fmt.Errorf("code size must be a positive integer")
	}

	// Calculate the number of bits needed to represent all possible codes (size * 5)
	bits := uint(size * 5)
	if bits > 63 {
		return nil, fmt.Errorf("code size too large")
	}

	// Set the mask to cover the entire code space (2^bits - 1)
	mask := uint64(1<<bits) - 1

	// Determine how to split the bits for the Feistel network. We want to split the bits into two halves, but if the number of bits is odd, one half will be larger than the other.
	split := bits / 2
	if split == 0 {
		split = 1
	}

	leftBits := bits - split
	leftMask := uint64(1<<leftBits) - 1
	rightMask := uint64(1<<split) - 1

	var roundKeys [3]uint64
	for i := 0; i < len(roundKeys); i++ {
		key, err := randomUint64()
		if err != nil {
			return nil, err
		}
		roundKeys[i] = key
	}

	return &codeGenerator{
		size:      size,
		bits:      bits,
		split:     split,
		mask:      mask,
		leftMask:  leftMask,
		rightMask: rightMask,
		roundKeys: roundKeys,
	}, nil
}

// Next generates the next code by incrementing the counter, applying the Feistel permutation, and encoding it as a string.
func (g *codeGenerator) Next() string {
	idx := atomic.AddUint64(&g.counter, 1) - 1
	idx &= g.mask
	return g.encode(g.permute(idx))
}

// permute applies a Feistel network to the input value using the round keys.
func (g *codeGenerator) permute(value uint64) uint64 {
	left := value >> g.split
	right := value & g.rightMask

	for _, key := range g.roundKeys {
		left, right = right, left^g.roundFunction(right, key)
	}

	return ((left << g.split) | right) & g.mask
}

// roundFunction is a simple mixing function that combines the value and key to produce a pseudo-random output.
func (g *codeGenerator) roundFunction(value, key uint64) uint64 {
	v := value ^ key
	v *= 0x9e3779b97f4a7c15 // golden ratio constant
	v ^= v >> 32
	return (v >> (64 - (g.bits - g.split))) & g.leftMask
}

// converts the permuted uint64 value to a code string using the codeAlphabet
func (g *codeGenerator) encode(value uint64) string {
	code := make([]byte, g.size)
	for i := g.size - 1; i >= 0; i-- {
		code[i] = codeAlphabet[value&0x1f]
		value >>= 5
	}
	return string(code)
}

// randomUint64 generates a random uint64 using crypto/rand
func randomUint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
