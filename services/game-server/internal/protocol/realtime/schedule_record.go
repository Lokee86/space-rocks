package realtime

// ScheduleRecord captures a single realtime send-plan decision for later metrics.
type ScheduleRecord struct {
	Lane            Lane
	PacketFamily    string
	RecordKind      string
	EntityFamily    string
	EntityID        string
	Priority        Priority
	DeliveryClass   DeliveryClass
	SupersessionKey string
	EstimatedBytes  int
	AgeTicks        int
	ChunkIndex      int
	ChunkCount      int
	IsFinalChunk    bool
	PayloadRef      any

	IncludedCount   int
	DeferredCount   int
	SupersededCount int
	RequiredCount   int
	CreateCount     int
	UpdateCount     int
	DeleteCount     int

	// PriorityBandCounts can be populated when the scheduler has simple band buckets.
	PriorityBandCounts map[string]int
}

// ScheduleSummary exposes the send-plan counts without requiring packet emission.
type ScheduleSummary struct {
	IncludedCount   int
	DeferredCount   int
	SupersededCount int
	RequiredCount   int
	CreateCount     int
	UpdateCount     int
	DeleteCount     int

	PriorityBandCounts map[string]int
}

// Summary returns a copy of the record counts for downstream metrics.
func (r ScheduleRecord) Summary() ScheduleSummary {
	summary := ScheduleSummary{
		IncludedCount:   r.IncludedCount,
		DeferredCount:   r.DeferredCount,
		SupersededCount: r.SupersededCount,
		RequiredCount:   r.RequiredCount,
		CreateCount:     r.CreateCount,
		UpdateCount:     r.UpdateCount,
		DeleteCount:     r.DeleteCount,
	}

	if len(r.PriorityBandCounts) > 0 {
		summary.PriorityBandCounts = make(map[string]int, len(r.PriorityBandCounts))
		for band, count := range r.PriorityBandCounts {
			summary.PriorityBandCounts[band] = count
		}
	}

	return summary
}
