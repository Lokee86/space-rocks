package measurement

type sampleRing struct {
	items        []PeriodicSample
	next         int
	count        int
	overwritten  uint64
}

func newSampleRing(capacity int) sampleRing {
	if capacity < 1 {
		capacity = 1
	}
	return sampleRing{items: make([]PeriodicSample, capacity)}
}

func (ring *sampleRing) add(sample PeriodicSample) {
	if ring.count == len(ring.items) {
		ring.overwritten++
	} else {
		ring.count++
	}
	ring.items[ring.next] = sample
	ring.next = (ring.next + 1) % len(ring.items)
}

func (ring sampleRing) list() []PeriodicSample {
	result := make([]PeriodicSample, ring.count)
	if ring.count == 0 {
		return result
	}
	start := 0
	if ring.count == len(ring.items) {
		start = ring.next
	}
	for i := range result {
		result[i] = ring.items[(start+i)%len(ring.items)]
	}
	return result
}
