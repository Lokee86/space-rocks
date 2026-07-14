package rng

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"time"
)

type Source struct {
	seed int64
	rng  *rand.Rand
}

func New(seed int64) *Source {
	return &Source{
		seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
	}
}

func NewProduction() *Source {
	var seed int64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &seed); err != nil {
		seed = time.Now().UnixNano()
	}

	return New(seed)
}

func (source *Source) Seed() int64 {
	return source.seed
}

func (source *Source) Intn(limit int) int {
	return source.rng.Intn(limit)
}

func (source *Source) Float64() float64 {
	return source.rng.Float64()
}
