package measurement

import "time"

const ReportVersion = 6

type RunContext struct {
	RunID          string    `json:"run_id"`
	SessionID      string    `json:"session_id"`
	RoomID         string    `json:"room_id"`
	MatchID        string    `json:"match_id"`
	SimulationSeed int64     `json:"simulation_seed"`
	BuildVersion   string    `json:"build_version"`
	StartTime      time.Time `json:"start_time"`
}

type StopReason string

const (
	StopReasonComplete     StopReason = "complete"
	StopReasonPartial      StopReason = "partial"
	StopReasonDisconnected StopReason = "disconnected"
	StopReasonShutdown     StopReason = "shutdown"
)

type EntityCounts struct {
	Players               int `json:"players"`
	PlayerSessions        int `json:"player_sessions"`
	Enemies               int `json:"enemies"`
	Asteroids             int `json:"asteroids"`
	Projectiles           int `json:"projectiles"`
	Pickups               int `json:"pickups"`
	RadialEffects         int `json:"radial_effects"`
	AsteroidsSpawnedTotal int `json:"asteroids_spawned_total"`
}

type ProcessSample struct {
	HeapAllocatedBytes   int64   `json:"heap_allocated_bytes"`
	HeapInUseBytes       int64   `json:"heap_in_use_bytes"`
	SystemBytes          int64   `json:"system_bytes"`
	ResidentSetBytes     int64   `json:"resident_set_bytes"`
	PeakResidentSetBytes int64   `json:"peak_resident_set_bytes"`
	CPUUserTimeNanos     int64   `json:"cpu_user_time_nanos"`
	CPUSystemTimeNanos   int64   `json:"cpu_system_time_nanos"`
	CPUUtilizationCores  float64 `json:"cpu_utilization_cores"`
	Goroutines           int     `json:"goroutines"`
	GCCycles             uint32  `json:"gc_cycles"`
	GCPauseTotalNanos    uint64  `json:"gc_pause_total_nanos"`
	GCPauseWindowNanos   uint64  `json:"gc_pause_window_nanos"`
	GCLastPauseNanos     uint64  `json:"gc_last_pause_nanos"`
}

type TickSummary struct {
	Count   uint64        `json:"count"`
	Minimum time.Duration `json:"minimum"`
	Maximum time.Duration `json:"maximum"`
	Total   time.Duration `json:"total"`
	Average time.Duration `json:"average"`
}

type PacketObservation struct {
	Lane         string `json:"lane"`
	PacketFamily string `json:"packet_family"`
	EncodedBytes int    `json:"encoded_bytes"`
}

type PacketSummary struct {
	Lane                string `json:"lane"`
	PacketFamily        string `json:"packet_family"`
	PacketCount         uint64 `json:"packet_count"`
	EncodedBytesTotal   uint64 `json:"encoded_bytes_total"`
	MaximumEncodedBytes int    `json:"maximum_encoded_bytes"`
}

type ReceiverLaneObservation struct {
	Lane          string
	BufferedBytes uint64
	Skipped       bool
}

type ReceiverLaneCandidateBuildObservation struct {
	StateAdvanceDuration      time.Duration `json:"state_advance_duration"`
	WorldHotLifecycleDuration time.Duration `json:"world_hot_lifecycle_duration"`
	PlayerLocatorDuration     time.Duration `json:"player_locator_duration"`
	OverlayDuration           time.Duration `json:"overlay_duration"`
	SessionDuration           time.Duration `json:"session_duration"`
	EventDuration             time.Duration `json:"event_duration"`
	CandidateFinalizeDuration time.Duration `json:"candidate_finalize_duration"`
}

type ReceiverCandidateBuildObservation struct {
	SnapshotCaptureDuration  time.Duration                         `json:"snapshot_capture_duration"`
	PendingEventCopyDuration time.Duration                         `json:"pending_event_copy_duration"`
	InterestFilterDuration   time.Duration                         `json:"interest_filter_duration"`
	LaneCandidatesDuration   time.Duration                         `json:"lane_candidates_duration"`
	LaneCandidatePhases      ReceiverLaneCandidateBuildObservation `json:"lane_candidate_phases"`
	ChunkPlanningDuration    time.Duration                         `json:"chunk_planning_duration"`
	SchedulingDuration       time.Duration                         `json:"scheduling_duration"`
}

type ReceiverTickObservation struct {
	CandidateBuildDuration time.Duration
	CandidateBuildPhases   ReceiverCandidateBuildObservation
	EncodingDuration       time.Duration
	OutboundDuration       time.Duration
	SkippedSend            bool
	Lanes                  []ReceiverLaneObservation
}

type ReceiverLaneSummary struct {
	Lane                 string `json:"lane"`
	SampleCount          uint64 `json:"sample_count"`
	CurrentBufferedBytes uint64 `json:"current_buffered_bytes"`
	PeakBufferedBytes    uint64 `json:"peak_buffered_bytes"`
	SkippedSendTicks     uint64 `json:"skipped_send_ticks"`
}

type ReceiverLaneCandidateBuildSummary struct {
	StateAdvanceTime      TickSummary `json:"state_advance_time"`
	WorldHotLifecycleTime TickSummary `json:"world_hot_lifecycle_time"`
	PlayerLocatorTime     TickSummary `json:"player_locator_time"`
	OverlayTime           TickSummary `json:"overlay_time"`
	SessionTime           TickSummary `json:"session_time"`
	EventTime             TickSummary `json:"event_time"`
	CandidateFinalizeTime TickSummary `json:"candidate_finalize_time"`
}

type ReceiverCandidateBuildSummary struct {
	SnapshotCaptureTime  TickSummary                       `json:"snapshot_capture_time"`
	PendingEventCopyTime TickSummary                       `json:"pending_event_copy_time"`
	InterestFilterTime   TickSummary                       `json:"interest_filter_time"`
	LaneCandidatesTime   TickSummary                       `json:"lane_candidates_time"`
	LaneCandidatePhases  ReceiverLaneCandidateBuildSummary `json:"lane_candidate_phases"`
	ChunkPlanningTime    TickSummary                       `json:"chunk_planning_time"`
	SchedulingTime       TickSummary                       `json:"scheduling_time"`
}

type ReceiverCandidateBuildPeak struct {
	Total  time.Duration                     `json:"total"`
	Phases ReceiverCandidateBuildObservation `json:"phases"`
}

type ReceiverSummary struct {
	TickCount            uint64                        `json:"tick_count"`
	SkippedSendTicks     uint64                        `json:"skipped_send_ticks"`
	CandidateBuildTime   TickSummary                   `json:"candidate_build_time"`
	CandidateBuildPhases ReceiverCandidateBuildSummary `json:"candidate_build_phases"`
	CandidateBuildPeak   ReceiverCandidateBuildPeak    `json:"candidate_build_peak"`
	EncodingTime         TickSummary                   `json:"encoding_time"`
	OutboundTime         TickSummary                   `json:"outbound_time"`
	Lanes                []ReceiverLaneSummary         `json:"lanes"`
}

type PeriodicSample struct {
	Elapsed    time.Duration `json:"elapsed"`
	Entities   EntityCounts  `json:"entities"`
	Process    ProcessSample `json:"process"`
	RoomCount  int           `json:"room_count"`
	TickWindow TickSummary   `json:"tick_window"`
}

type ServerReport struct {
	Version                   int              `json:"version"`
	Context                   RunContext       `json:"context"`
	StopReason                StopReason       `json:"stop_reason,omitempty"`
	Complete                  bool             `json:"complete"`
	StartedAt                 time.Time        `json:"started_at"`
	EndedAt                   time.Time        `json:"ended_at"`
	Duration                  time.Duration    `json:"duration"`
	Ticks                     TickSummary      `json:"ticks"`
	Samples                   []PeriodicSample `json:"samples"`
	OverwrittenSampleCount    uint64           `json:"overwritten_sample_count"`
	Packets                   []PacketSummary  `json:"packets"`
	DroppedPacketObservations uint64           `json:"dropped_packet_observations"`
	Receiver                  ReceiverSummary  `json:"receiver"`
}

type SimulationObserver interface {
	ObserveSimulation(time.Duration, EntityCounts)
}

type SimulationObserverFunc func(time.Duration, EntityCounts)

func (observer SimulationObserverFunc) ObserveSimulation(duration time.Duration, counts EntityCounts) {
	observer(duration, counts)
}
