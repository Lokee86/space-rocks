package realtime

import "testing"

func findMatchingScheduleRecords(records []ScheduleRecord, entityID, recordKind string, deliveryClass DeliveryClass, priority ...Priority) []ScheduleRecord {
	matches := make([]ScheduleRecord, 0, len(records))
	for _, record := range records {
		if record.EntityID != entityID || record.RecordKind != recordKind || record.DeliveryClass != deliveryClass {
			continue
		}
		if len(priority) > 0 && record.Priority != priority[0] {
			continue
		}
		matches = append(matches, record)
	}
	return matches
}

func assertMatchingScheduleRecordsChunked(t *testing.T, matches []ScheduleRecord) {
	t.Helper()
	if len(matches) <= 1 {
		t.Fatalf("matches = %#v, want chunked records", matches)
	}

	chunkCount := matches[0].ChunkCount
	if chunkCount <= 1 {
		t.Fatalf("matches = %#v, want chunked records", matches)
	}

	saw := make(map[int]bool, chunkCount)
	var finalCount int
	for _, record := range matches {
		if record.ChunkCount != chunkCount {
			t.Fatalf("matches = %#v, want consistent chunk count", matches)
		}
		if record.ChunkIndex < 0 || record.ChunkIndex >= chunkCount {
			t.Fatalf("matches = %#v, want chunk indexes in range", matches)
		}
		saw[record.ChunkIndex] = true
		if record.IsFinalChunk {
			finalCount++
		}
	}

	if len(saw) != chunkCount {
		t.Fatalf("matches = %#v, want chunk indexes to cover %d chunks", matches, chunkCount)
	}
	if finalCount != 1 {
		t.Fatalf("matches = %#v, want exactly one final chunk", matches)
	}
}

func TestSelectSendPlanPrefersCriticalBeforeLow(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneSession, PacketFamily: PacketFamilySessionDelta, RecordKind: "update", Priority: PriorityLow, EstimatedBytes: 100, EntityID: "low"},
		{Lane: LaneControl, PacketFamily: PacketFamilyResyncRequired, RecordKind: "required", Priority: PriorityCritical, DeliveryClass: DeliveryClassRequired, EstimatedBytes: 100, EntityID: "critical"},
	}

	plan := SelectSendPlan(records)
	if len(plan.Included) < 2 {
		t.Fatalf("included records = %#v, want both records", plan.Included)
	}
	if plan.Included[0].EntityID != "critical" {
		t.Fatalf("first included record = %q, want critical", plan.Included[0].EntityID)
	}
}

func TestSelectSendPlanNeverDropsRequiredDelete(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneWorld, PacketFamily: PacketFamilyWorldDelta, RecordKind: "delete", DeliveryClass: DeliveryClassRequired, Priority: PriorityCritical, EstimatedBytes: HardCapBytes + 200, EntityID: "delete-1"},
	}

	plan := SelectSendPlan(records)
	matches := findMatchingScheduleRecords(plan.Included, "delete-1", "delete", DeliveryClassRequired, PriorityCritical)
	if len(matches) == 0 {
		t.Fatalf("included records = %#v, want required delete included", plan.Included)
	}
	if len(matches) == 1 {
		if matches[0].RecordKind != "delete" || matches[0].DeliveryClass != DeliveryClassRequired || matches[0].Priority != PriorityCritical {
			t.Fatalf("included record = %#v, want required delete included", matches[0])
		}
		return
	}
	assertMatchingScheduleRecordsChunked(t, matches)
}

func TestSelectSendPlanDefersRequiredOverflowInsteadOfDiscarding(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneWorld, PacketFamily: PacketFamilyWorldFull, RecordKind: "create", DeliveryClass: DeliveryClassRequired, Priority: PriorityCritical, EstimatedBytes: HardCapBytes + 500, EntityID: "oversized"},
	}

	plan := SelectSendPlan(records)
	for _, record := range plan.Deferred {
		if record.EntityID == "oversized" {
			return
		}
	}

	matches := findMatchingScheduleRecords(plan.Included, "oversized", "create", DeliveryClassRequired, PriorityCritical)
	if len(matches) == 0 {
		t.Fatalf("included records = %#v deferred records = %#v, want oversized required record retained", plan.Included, plan.Deferred)
	}
	if len(matches) == 1 {
		if matches[0].ChunkCount <= 1 {
			t.Fatalf("included record = %#v, want chunked/staged oversized required record retained", matches[0])
		}
		return
	}
	assertMatchingScheduleRecordsChunked(t, matches)
}

func TestSelectSendPlanDropsDebugUnderPressure(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneSession, PacketFamily: PacketFamilySessionDelta, RecordKind: "update", Priority: PriorityLow, EstimatedBytes: TargetBytes, EntityID: "pressure"},
		{Lane: LaneSession, PacketFamily: PacketFamilySessionDelta, RecordKind: "update", Priority: PriorityDebug, DeliveryClass: DeliveryClassDebugOnly, EstimatedBytes: 10, EntityID: "debug"},
	}

	plan := SelectSendPlan(records)
	for _, record := range append(plan.Included, plan.Deferred...) {
		if record.EntityID == "debug" {
			t.Fatalf("debug record survived pressure: %#v", record)
		}
	}
}

func TestSelectSendPlanPrefersLocalOverlayBeforeFarWorld(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneWorld, PacketFamily: PacketFamilyWorldDelta, RecordKind: "update", Priority: PriorityHigh, EstimatedBytes: 50, EntityID: "world-far"},
		{Lane: LaneOverlay, PacketFamily: PacketFamilyOverlayFull, RecordKind: "update", EntityFamily: "self", EstimatedBytes: 50, EntityID: "overlay-self"},
	}

	plan := SelectSendPlan(records)
	if len(plan.Included) < 2 {
		t.Fatalf("included records = %#v, want both records", plan.Included)
	}
	if plan.Included[0].EntityID != "overlay-self" {
		t.Fatalf("first included record = %q, want overlay-self", plan.Included[0].EntityID)
	}
}

func TestSelectSendPlanOrderingIsDeterministic(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneWorld, PacketFamily: PacketFamilyWorldDelta, RecordKind: "update", EntityID: "b", Priority: PriorityMedium, EstimatedBytes: 1},
		{Lane: LaneWorld, PacketFamily: PacketFamilyWorldDelta, RecordKind: "update", EntityID: "a", Priority: PriorityMedium, EstimatedBytes: 1},
	}

	first := SelectSendPlan(records)
	second := SelectSendPlan(records)
	if len(first.Included) != len(second.Included) {
		t.Fatalf("included lengths differ: %d vs %d", len(first.Included), len(second.Included))
	}
	for i := range first.Included {
		if first.Included[i].EntityID != second.Included[i].EntityID {
			t.Fatalf("included ordering changed: %#v vs %#v", first.Included, second.Included)
		}
	}
}

func TestDeferredStoreSupersedesOlderMovementByKey(t *testing.T) {
	store := NewDeferredStore()
	store.Add(ScheduleRecord{SupersessionKey: "ship-1", EntityID: "old", AgeTicks: 1, DeliveryClass: DeliveryClassHotSupersedable})
	if !store.Supersede(ScheduleRecord{SupersessionKey: "ship-1", EntityID: "new", AgeTicks: 2, DeliveryClass: DeliveryClassHotSupersedable}) {
		t.Fatalf("expected supersession to replace older record")
	}
	pending := store.Pending()
	if len(pending) != 1 || pending[0].EntityID != "new" {
		t.Fatalf("pending = %#v, want replaced record", pending)
	}
}

func TestDeferredStoreAgesDeferrableRecords(t *testing.T) {
	store := NewDeferredStore()
	store.Add(ScheduleRecord{EntityID: "defer-1", DeliveryClass: DeliveryClassDeferrable, AgeTicks: 0})
	aged := store.Age()
	if len(aged) != 1 || aged[0].AgeTicks != 1 {
		t.Fatalf("aged = %#v, want age tick incremented", aged)
	}
}

func TestDeferredStoreStagesChunkedFullSnapshot(t *testing.T) {
	store := NewDeferredStore()
	store.Stage(ScheduleRecord{EntityID: "snapshot", PacketFamily: PacketFamilyWorldFull, DeliveryClass: DeliveryClassRequired, ChunkCount: 3, ChunkIndex: 0, IsFinalChunk: false})
	pending := store.Pending()
	if len(pending) != 1 || pending[0].ChunkCount != 3 {
		t.Fatalf("pending = %#v, want staged chunked snapshot", pending)
	}
}

func TestDeferredStoreEventOnceRemainsPendingUntilAck(t *testing.T) {
	store := NewDeferredStore()
	record := ScheduleRecord{EntityID: "event-batch", PacketFamily: PacketFamilyEventBatch, DeliveryClass: DeliveryClassEventOnce}
	store.Add(record)
	if len(store.Pending()) != 1 {
		t.Fatalf("pending = %#v, want event batch pending", store.Pending())
	}
	store.Acknowledge(record)
	if len(store.Pending()) != 0 {
		t.Fatalf("pending = %#v, want acked event batch removed", store.Pending())
	}
}

func TestSendPlanExposesCounts(t *testing.T) {
	records := []ScheduleRecord{
		{EntityID: "create", RecordKind: "create", CreateCount: 1, EstimatedBytes: 1},
		{EntityID: "delete", RecordKind: "delete", DeleteCount: 1, EstimatedBytes: 1},
	}
	plan := SelectSendPlan(records)
	if plan.Summary.CreateCount != 1 || plan.Summary.DeleteCount != 1 {
		t.Fatalf("summary = %#v, want counts exposed", plan.Summary)
	}
}


func TestSelectSendPlanDefersLowerPriorityUnderBudgetPressure(t *testing.T) {
	records := []ScheduleRecord{
		{Lane: LaneWorld, PacketFamily: PacketFamilyWorldDelta, RecordKind: "update", Priority: PriorityHigh, EstimatedBytes: TargetBytes, EntityID: "high"},
		{Lane: LaneSession, PacketFamily: PacketFamilySessionDelta, RecordKind: "update", Priority: PriorityLow, EstimatedBytes: 10, EntityID: "low"},
	}

	plan := SelectSendPlan(records)
	if len(plan.Deferred) != 1 || plan.Deferred[0].EntityID != "low" {
		t.Fatalf("deferred = %#v, want lower priority deferred", plan.Deferred)
	}
}

func TestDeferredStoreAgesAndSupersedesHotRecords(t *testing.T) {
	store := NewDeferredStore()
	store.Add(ScheduleRecord{SupersessionKey: "ship-1", EntityID: "old", DeliveryClass: DeliveryClassHotSupersedable, AgeTicks: 0})
	aged := store.Age()
	if len(aged) != 1 || aged[0].AgeTicks != 1 {
		t.Fatalf("aged = %#v, want age incremented", aged)
	}
	if !store.Supersede(ScheduleRecord{SupersessionKey: "ship-1", EntityID: "new", DeliveryClass: DeliveryClassHotSupersedable, AgeTicks: 2}) {
		t.Fatal("expected hot supersedable record to be replaced")
	}
	pending := store.Pending()
	if len(pending) != 1 || pending[0].EntityID != "new" {
		t.Fatalf("pending = %#v, want superseded record replaced", pending)
	}
}
