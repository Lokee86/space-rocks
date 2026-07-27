package measurement

import "time"

const ReportVersion = 2

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
	Players              int `json:"players"`
	PlayerSessions       int `json:"player_sessions"`
	Enemies              int `json:"enemies"`
	Asteroids            int `json:"asteroids"`
	Projectiles          int `json:"projectiles"`
	Pickups              int `json:"pickups"`
	RadialEffects        int `json:"radial_effects"`
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
	Lane              string `json:"lane"`
	PacketFamily     string `json:"packet_family"`
	PacketCount       uint64 `json:"packet_count"`
	EncodedBytesTotal uint64 `json:"encoded_bytes_total"`
	MaximumEncodedBytes int `json:"maximum_encoded_bytes"`
}

type PeriodicSample struct {
	Elapsed      time.Duration `json:"elapsed"`
	Entities     EntityCounts  `json:"entities"`
	Process      ProcessSample `json:"process"`
	RoomCount    int           `json:"room_count"`
	TickWindow   TickSummary   `json:"tick_window"`
}

type ServerReport struct {
	Version                   int             `json:"version"`
	Context                   RunContext      `json:"context"`
	StopReason                StopReason      `json:"stop_reason,omitempty"`
	Complete                  bool            `json:"complete"`
	StartedAt                 time.Time       `json:"started_at"`
	EndedAt                   time.Time       `json:"ended_at"`
	Duration                  time.Duration   `json:"duration"`
	Ticks                     TickSummary     `json:"ticks"`
	Samples                   []PeriodicSample `json:"samples"`
	OverwrittenSampleCount    uint64          `json:"overwritten_sample_count"`
	Packets                   []PacketSummary `json:"packets"`
	DroppedPacketObservations uint64          `json:"dropped_packet_observations"`
}

type SimulationObserver interface {
	ObserveSimulation(time.Duration, EntityCounts)
}

type SimulationObserverFunc func(time.Duration, EntityCounts)

func (observer SimulationObserverFunc) ObserveSimulation(duration time.Duration, counts EntityCounts) {
	observer(duration, counts)
}
