package realtime

import "sort"

// SendPlanSummary is the scheduler-facing summary payload for later metrics.
type SendPlanSummary = ScheduleSummary

type SendPlan struct {
	Included []ScheduleRecord
	Deferred []ScheduleRecord
	Summary  SendPlanSummary
}

func SelectSendPlan(records []ScheduleRecord) SendPlan {
	ordered := append([]ScheduleRecord(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return scheduleRank(ordered[i]) < scheduleRank(ordered[j])
	})

	plan := SendPlan{
		Included: make([]ScheduleRecord, 0, len(ordered)),
		Deferred: make([]ScheduleRecord, 0, len(ordered)),
	}

	budget := TargetBytes
	for _, record := range ordered {
		if record.DeliveryClass == DeliveryClassDebugOnly && budget < WarningBytes {
			continue
		}
		budget = appendPlannedRecord(&plan, record, budget)
	}

	plan.Summary = summarizeSendPlan(plan.Included, plan.Deferred)
	return plan
}

func isRealHotLaneChunkRecord(record ScheduleRecord) bool {
	if record.ChunkCount <= 1 {
		return false
	}
	return record.Lane == LaneBullets || record.Lane == LaneAsteroids
}

func appendPlannedRecord(plan *SendPlan, record ScheduleRecord, budget int) int {
	estimated := record.EstimatedBytes
	if estimated <= 0 {
		estimated = EstimatePacketBytes(record.PacketFamily, 1, 0)
	}

	if isRequiredRecord(record) || isRealHotLaneChunkRecord(record) {
		plan.Included = append(plan.Included, record)
		return budget - estimated
	}

	if budget-estimated >= 0 {
		plan.Included = append(plan.Included, record)
		return budget - estimated
	}

	plan.Deferred = append(plan.Deferred, record)
	return budget
}

func summarizeSendPlan(included []ScheduleRecord, deferred []ScheduleRecord) SendPlanSummary {
	summary := SendPlanSummary{
		IncludedCount: len(included),
		DeferredCount: len(deferred),
	}

	for _, record := range included {
		summary.CreateCount += record.CreateCount
		summary.UpdateCount += record.UpdateCount
		summary.DeleteCount += record.DeleteCount
		summary.RequiredCount += record.RequiredCount
		summary.SupersededCount += record.SupersededCount
	}

	return summary
}

func scheduleRank(record ScheduleRecord) int {
	if isRequiredRecord(record) {
		return laneBootstrapRank(record)
	}
	if record.DeliveryClass == DeliveryClassEventOnce || record.PacketFamily == PacketFamilyEventBatch {
		return 4
	}
	if record.Priority == PriorityHigh {
		return 5
	}
	if record.Priority == PriorityMedium {
		return 6
	}
	if record.Lane == LaneSession || record.Priority == PriorityLow {
		return 7
	}
	return 8
}

func hotPacketCadenceAllows(mode HotLaneMode, sequence int) bool {
	if sequence <= 0 {
		sequence = 1
	}
	switch mode {
	case HotLaneModeFullOwned30Hz:
		return sequence%2 == 0
	case HotLaneModeFullOwned20Hz:
		return sequence%3 == 0
	default:
		return true
	}
}

func laneBootstrapRank(record ScheduleRecord) int {
	switch record.Lane {
	case LaneWorld:
		return 0
	case LaneOverlay:
		return 1
	case LaneSession:
		return 2
	default:
		return 3
	}
}
func isRequiredRecord(record ScheduleRecord) bool {
	if record.DeliveryClass == DeliveryClassRequired {
		return true
	}
	if record.Priority == PriorityCritical {
		return true
	}
	if record.Lane == LaneControl {
		return true
	}
	if record.PacketFamily == PacketFamilyWorldFull || record.PacketFamily == PacketFamilyOverlayFull || record.PacketFamily == PacketFamilySessionFull {
		return true
	}
	if record.PacketFamily == PacketFamilyResyncRequest || record.PacketFamily == PacketFamilyResyncRequired {
		return true
	}
	return false
}

func BuildDeferredStore(records []ScheduleRecord) *DeferredStore {
	store := NewDeferredStore()
	for _, record := range records {
		store.Add(record)
	}
	return store
}
