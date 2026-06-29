package realtime

type DeferredStore struct {
	records map[string]ScheduleRecord
}

type DeferredAck struct {
	Record ScheduleRecord
	Sent   bool
}

func NewDeferredStore() *DeferredStore {
	return &DeferredStore{records: make(map[string]ScheduleRecord)}
}

func (store *DeferredStore) Add(record ScheduleRecord) {
	if store.records == nil {
		store.records = make(map[string]ScheduleRecord)
	}

	if record.DeliveryClass == DeliveryClassDebugOnly {
		return
	}

	if record.SupersessionKey != "" {
		if existing, ok := store.records[record.SupersessionKey]; ok && existing.AgeTicks <= record.AgeTicks {
			store.records[record.SupersessionKey] = record
			return
		}
	}

	key := storeKey(record)
	store.records[key] = record
}

func (store *DeferredStore) Stage(record ScheduleRecord) {
	store.Add(record)
}

func (store *DeferredStore) AddOrReplace(record ScheduleRecord) {
	store.Add(record)
}

func (store *DeferredStore) Supersede(record ScheduleRecord) bool {
	if record.SupersessionKey == "" {
		return false
	}
	if existing, ok := store.records[record.SupersessionKey]; ok && existing.AgeTicks <= record.AgeTicks {
		store.records[record.SupersessionKey] = record
		return true
	}
	return false
}

func (store *DeferredStore) Acknowledge(record ScheduleRecord) DeferredAck {
	store.MarkSent(record)
	return DeferredAck{Record: record, Sent: true}
}

func (store *DeferredStore) Age() []ScheduleRecord {
	aged := make([]ScheduleRecord, 0, len(store.records))
	for key, record := range store.records {
		record.AgeTicks++
		store.records[key] = record
		aged = append(aged, record)
	}
	return aged
}

func (store *DeferredStore) Pending() []ScheduleRecord {
	pending := make([]ScheduleRecord, 0, len(store.records))
	for _, record := range store.records {
		pending = append(pending, record)
	}
	return pending
}

func (store *DeferredStore) MarkSent(record ScheduleRecord) {
	delete(store.records, storeKey(record))
}

func storeKey(record ScheduleRecord) string {
	if record.SupersessionKey != "" {
		return record.SupersessionKey
	}
	if record.EntityID != "" {
		return record.EntityID
	}
	return record.PacketFamily + ":" + record.RecordKind
}
