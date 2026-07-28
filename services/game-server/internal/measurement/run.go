package measurement

import (
	"sort"
	"sync"
	"time"
)

const (
	defaultSampleInterval = time.Second
	defaultSampleCapacity = 3600
	maxPacketKeys         = 128
	maxIdentifierLength   = 64
)

type RunOption func(*runConfig)

type runConfig struct {
	clock          func() time.Time
	sampleInterval time.Duration
	sampleCapacity int
	processSampler *ProcessSampler
	roomCount      func() int
}

func WithClock(clock func() time.Time) RunOption {
	return func(config *runConfig) {
		if clock != nil {
			config.clock = clock
		}
	}
}

func WithSampleInterval(interval time.Duration) RunOption {
	return func(config *runConfig) {
		if interval > 0 {
			config.sampleInterval = interval
		}
	}
}

func WithSampleCapacity(capacity int) RunOption {
	return func(config *runConfig) {
		if capacity > 0 {
			config.sampleCapacity = capacity
		}
	}
}

func WithProcessSampler(sampler *ProcessSampler) RunOption {
	return func(config *runConfig) {
		if sampler != nil {
			config.processSampler = sampler
		}
	}
}

func WithRoomCountProvider(provider func() int) RunOption {
	return func(config *runConfig) { config.roomCount = provider }
}

type Run struct {
	mu                         sync.Mutex
	context                    RunContext
	clock                      func() time.Time
	start                      time.Time
	sampleInterval             time.Duration
	nextSampleElapsed          time.Duration
	processSampler             *ProcessSampler
	roomCount                  func() int
	sampleCapacity             int
	active                     bool
	finalReport                *ServerReport
	ticks                      tickAccumulator
	window                     tickAccumulator
	lastEntities               EntityCounts
	hasEntities                bool
	lastSampleElapsed          time.Duration
	packets                    map[packetKey]*PacketSummary
	droppedPacketObservations  uint64
	receiverTicks              uint64
	receiverSkippedSendTicks   uint64
	receiverCandidateBuild     tickAccumulator
	receiverCandidateBuildPeak ReceiverCandidateBuildPeak
	receiverSnapshotCapture    tickAccumulator
	receiverPendingEventCopy   tickAccumulator
	receiverInterestFilter     tickAccumulator
	receiverLaneCandidates     tickAccumulator
	receiverLaneStateAdvance   tickAccumulator
	receiverWorldHotLifecycle  tickAccumulator
	receiverPlayerLocator      tickAccumulator
	receiverOverlayCandidates  tickAccumulator
	receiverSessionCandidates  tickAccumulator
	receiverEventCandidates    tickAccumulator
	receiverCandidateFinalize  tickAccumulator
	receiverChunkPlanning      tickAccumulator
	receiverScheduling         tickAccumulator
	receiverEncoding           tickAccumulator
	receiverOutbound           tickAccumulator
	receiverLanes              map[string]*ReceiverLaneSummary
	samples                    sampleRing
}

type packetKey struct {
	lane   string
	family string
}

func NewRun(context RunContext, options ...RunOption) *Run {
	config := runConfig{
		clock:          time.Now,
		sampleInterval: defaultSampleInterval,
		sampleCapacity: defaultSampleCapacity,
		processSampler: defaultProcessSampler,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}

	now := config.clock()
	if context.StartTime.IsZero() {
		context.StartTime = now
	}
	return &Run{
		context:           context,
		clock:             config.clock,
		start:             context.StartTime,
		sampleInterval:    config.sampleInterval,
		nextSampleElapsed: config.sampleInterval,
		processSampler:    config.processSampler,
		roomCount:         config.roomCount,
		sampleCapacity:    config.sampleCapacity,
		active:            true,
		packets:           make(map[packetKey]*PacketSummary),
		receiverLanes:     make(map[string]*ReceiverLaneSummary),
		samples:           newSampleRing(config.sampleCapacity),
	}
}

func (run *Run) ObserveSimulation(duration time.Duration, entities EntityCounts) {
	if duration < 0 {
		duration = 0
	}
	run.mu.Lock()
	defer run.mu.Unlock()
	if !run.active {
		return
	}

	run.ticks.add(duration)
	run.window.add(duration)
	run.lastEntities = entities
	run.hasEntities = true
	now := run.clock()
	elapsed := now.Sub(run.start)
	if elapsed < run.nextSampleElapsed {
		return
	}

	run.sampleLocked(now, elapsed)
	run.nextSampleElapsed = elapsed + run.sampleInterval
}

func (run *Run) ObservePacket(observation PacketObservation) {
	run.mu.Lock()
	defer run.mu.Unlock()
	if !run.active || observation.EncodedBytes < 0 || !validIdentifier(observation.Lane) || !validIdentifier(observation.PacketFamily) {
		if run.active {
			run.droppedPacketObservations++
		}
		return
	}

	key := packetKey{lane: observation.Lane, family: observation.PacketFamily}
	summary, ok := run.packets[key]
	if !ok {
		if len(run.packets) >= maxPacketKeys {
			run.droppedPacketObservations++
			return
		}
		summary = &PacketSummary{Lane: observation.Lane, PacketFamily: observation.PacketFamily}
		run.packets[key] = summary
	}
	summary.PacketCount++
	summary.EncodedBytesTotal += uint64(observation.EncodedBytes)
	if observation.EncodedBytes > summary.MaximumEncodedBytes {
		summary.MaximumEncodedBytes = observation.EncodedBytes
	}
}

func (run *Run) ObservePacketWrite(lane string, packetFamily string, encodedBytes int) {
	run.ObservePacket(PacketObservation{Lane: lane, PacketFamily: packetFamily, EncodedBytes: encodedBytes})
}

func (run *Run) ObserveReceiverTick(observation ReceiverTickObservation) {
	run.mu.Lock()
	defer run.mu.Unlock()
	if !run.active {
		return
	}

	run.receiverTicks++
	if observation.SkippedSend {
		run.receiverSkippedSendTicks++
	}
	candidateBuildDuration := maxDuration(observation.CandidateBuildDuration)
	candidateBuildPhases := ReceiverCandidateBuildObservation{
		SnapshotCaptureDuration:  maxDuration(observation.CandidateBuildPhases.SnapshotCaptureDuration),
		PendingEventCopyDuration: maxDuration(observation.CandidateBuildPhases.PendingEventCopyDuration),
		InterestFilterDuration:   maxDuration(observation.CandidateBuildPhases.InterestFilterDuration),
		LaneCandidatesDuration:   maxDuration(observation.CandidateBuildPhases.LaneCandidatesDuration),
		LaneCandidatePhases: ReceiverLaneCandidateBuildObservation{
			StateAdvanceDuration:      maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.StateAdvanceDuration),
			WorldHotLifecycleDuration: maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.WorldHotLifecycleDuration),
			PlayerLocatorDuration:     maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.PlayerLocatorDuration),
			OverlayDuration:           maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.OverlayDuration),
			SessionDuration:           maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.SessionDuration),
			EventDuration:             maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.EventDuration),
			CandidateFinalizeDuration: maxDuration(observation.CandidateBuildPhases.LaneCandidatePhases.CandidateFinalizeDuration),
		},
		ChunkPlanningDuration: maxDuration(observation.CandidateBuildPhases.ChunkPlanningDuration),
		SchedulingDuration:    maxDuration(observation.CandidateBuildPhases.SchedulingDuration),
	}
	run.receiverCandidateBuild.add(candidateBuildDuration)
	if candidateBuildDuration > run.receiverCandidateBuildPeak.Total {
		run.receiverCandidateBuildPeak = ReceiverCandidateBuildPeak{
			Total:  candidateBuildDuration,
			Phases: candidateBuildPhases,
		}
	}
	run.receiverSnapshotCapture.add(candidateBuildPhases.SnapshotCaptureDuration)
	run.receiverPendingEventCopy.add(candidateBuildPhases.PendingEventCopyDuration)
	run.receiverInterestFilter.add(candidateBuildPhases.InterestFilterDuration)
	run.receiverLaneCandidates.add(candidateBuildPhases.LaneCandidatesDuration)
	run.receiverLaneStateAdvance.add(candidateBuildPhases.LaneCandidatePhases.StateAdvanceDuration)
	run.receiverWorldHotLifecycle.add(candidateBuildPhases.LaneCandidatePhases.WorldHotLifecycleDuration)
	run.receiverPlayerLocator.add(candidateBuildPhases.LaneCandidatePhases.PlayerLocatorDuration)
	run.receiverOverlayCandidates.add(candidateBuildPhases.LaneCandidatePhases.OverlayDuration)
	run.receiverSessionCandidates.add(candidateBuildPhases.LaneCandidatePhases.SessionDuration)
	run.receiverEventCandidates.add(candidateBuildPhases.LaneCandidatePhases.EventDuration)
	run.receiverCandidateFinalize.add(candidateBuildPhases.LaneCandidatePhases.CandidateFinalizeDuration)
	run.receiverChunkPlanning.add(candidateBuildPhases.ChunkPlanningDuration)
	run.receiverScheduling.add(candidateBuildPhases.SchedulingDuration)
	run.receiverEncoding.add(maxDuration(observation.EncodingDuration))
	run.receiverOutbound.add(maxDuration(observation.OutboundDuration))
	for _, laneObservation := range observation.Lanes {
		if !validIdentifier(laneObservation.Lane) {
			continue
		}
		summary := run.receiverLanes[laneObservation.Lane]
		if summary == nil {
			summary = &ReceiverLaneSummary{Lane: laneObservation.Lane}
			run.receiverLanes[laneObservation.Lane] = summary
		}
		summary.SampleCount++
		summary.CurrentBufferedBytes = laneObservation.BufferedBytes
		if laneObservation.BufferedBytes > summary.PeakBufferedBytes {
			summary.PeakBufferedBytes = laneObservation.BufferedBytes
		}
		if laneObservation.Skipped {
			summary.SkippedSendTicks++
		}
	}
}

func (run *Run) Snapshot() ServerReport {
	run.mu.Lock()
	defer run.mu.Unlock()
	if run.finalReport != nil {
		return cloneReport(*run.finalReport)
	}
	return run.reportLocked("")
}

func (run *Run) Finalize(reasons ...StopReason) ServerReport {
	run.mu.Lock()
	defer run.mu.Unlock()
	if run.finalReport != nil {
		return cloneReport(*run.finalReport)
	}
	reason := StopReasonComplete
	if len(reasons) > 0 && reasons[0] != "" {
		reason = reasons[0]
	}
	if run.hasEntities {
		now := run.clock()
		elapsed := now.Sub(run.start)
		if elapsed > run.lastSampleElapsed || len(run.samples.list()) == 0 {
			run.sampleLocked(now, elapsed)
		}
	}
	report := run.reportLocked(reason)
	run.active = false
	run.finalReport = &report
	return cloneReport(report)
}

func (run *Run) IsActive() bool {
	run.mu.Lock()
	defer run.mu.Unlock()
	return run.active
}

// Reset clears an active run's accumulated measurements while preserving its
// run identity and observer attachment.
func (run *Run) Reset() {
	run.mu.Lock()
	defer run.mu.Unlock()
	if !run.active {
		return
	}

	now := run.clock()
	run.start = now
	run.context.StartTime = now
	run.nextSampleElapsed = run.sampleInterval
	run.ticks = tickAccumulator{}
	run.window = tickAccumulator{}
	run.lastEntities = EntityCounts{}
	run.hasEntities = false
	run.lastSampleElapsed = 0
	run.packets = make(map[packetKey]*PacketSummary)
	run.droppedPacketObservations = 0
	run.receiverTicks = 0
	run.receiverSkippedSendTicks = 0
	run.receiverCandidateBuild = tickAccumulator{}
	run.receiverCandidateBuildPeak = ReceiverCandidateBuildPeak{}
	run.receiverSnapshotCapture = tickAccumulator{}
	run.receiverPendingEventCopy = tickAccumulator{}
	run.receiverInterestFilter = tickAccumulator{}
	run.receiverLaneCandidates = tickAccumulator{}
	run.receiverLaneStateAdvance = tickAccumulator{}
	run.receiverWorldHotLifecycle = tickAccumulator{}
	run.receiverPlayerLocator = tickAccumulator{}
	run.receiverOverlayCandidates = tickAccumulator{}
	run.receiverSessionCandidates = tickAccumulator{}
	run.receiverEventCandidates = tickAccumulator{}
	run.receiverCandidateFinalize = tickAccumulator{}
	run.receiverChunkPlanning = tickAccumulator{}
	run.receiverScheduling = tickAccumulator{}
	run.receiverEncoding = tickAccumulator{}
	run.receiverOutbound = tickAccumulator{}
	run.receiverLanes = make(map[string]*ReceiverLaneSummary)
	run.samples = newSampleRing(run.sampleCapacity)
}

func (run *Run) reportLocked(reason StopReason) ServerReport {
	ended := run.clock()
	report := ServerReport{
		Version:                   ReportVersion,
		Context:                   run.context,
		StopReason:                reason,
		Complete:                  reason == StopReasonComplete,
		StartedAt:                 run.start,
		EndedAt:                   ended,
		Duration:                  ended.Sub(run.start),
		Ticks:                     run.ticks.snapshot(),
		Samples:                   run.samples.list(),
		OverwrittenSampleCount:    run.samples.overwritten,
		DroppedPacketObservations: run.droppedPacketObservations,
	}
	for _, summary := range run.packets {
		report.Packets = append(report.Packets, *summary)
	}
	report.Receiver = ReceiverSummary{
		TickCount:          run.receiverTicks,
		SkippedSendTicks:   run.receiverSkippedSendTicks,
		CandidateBuildTime: run.receiverCandidateBuild.snapshot(),
		CandidateBuildPhases: ReceiverCandidateBuildSummary{
			SnapshotCaptureTime:  run.receiverSnapshotCapture.snapshot(),
			PendingEventCopyTime: run.receiverPendingEventCopy.snapshot(),
			InterestFilterTime:   run.receiverInterestFilter.snapshot(),
			LaneCandidatesTime:   run.receiverLaneCandidates.snapshot(),
			LaneCandidatePhases: ReceiverLaneCandidateBuildSummary{
				StateAdvanceTime:      run.receiverLaneStateAdvance.snapshot(),
				WorldHotLifecycleTime: run.receiverWorldHotLifecycle.snapshot(),
				PlayerLocatorTime:     run.receiverPlayerLocator.snapshot(),
				OverlayTime:           run.receiverOverlayCandidates.snapshot(),
				SessionTime:           run.receiverSessionCandidates.snapshot(),
				EventTime:             run.receiverEventCandidates.snapshot(),
				CandidateFinalizeTime: run.receiverCandidateFinalize.snapshot(),
			},
			ChunkPlanningTime: run.receiverChunkPlanning.snapshot(),
			SchedulingTime:    run.receiverScheduling.snapshot(),
		},
		CandidateBuildPeak: run.receiverCandidateBuildPeak,
		EncodingTime:       run.receiverEncoding.snapshot(),
		OutboundTime:       run.receiverOutbound.snapshot(),
	}
	for _, summary := range run.receiverLanes {
		report.Receiver.Lanes = append(report.Receiver.Lanes, *summary)
	}
	sort.Slice(report.Receiver.Lanes, func(i, j int) bool {
		return report.Receiver.Lanes[i].Lane < report.Receiver.Lanes[j].Lane
	})
	sort.Slice(report.Packets, func(i, j int) bool {
		if report.Packets[i].Lane == report.Packets[j].Lane {
			return report.Packets[i].PacketFamily < report.Packets[j].PacketFamily
		}
		return report.Packets[i].Lane < report.Packets[j].Lane
	})
	if report.Samples == nil {
		report.Samples = []PeriodicSample{}
	}
	if report.Packets == nil {
		report.Packets = []PacketSummary{}
	}
	if report.Receiver.Lanes == nil {
		report.Receiver.Lanes = []ReceiverLaneSummary{}
	}
	return report
}

func (run *Run) sampleLocked(now time.Time, elapsed time.Duration) {
	process := run.processSampler.Sample(now)
	roomCount := 0
	if run.roomCount != nil {
		roomCount = run.roomCount()
	}
	run.samples.add(PeriodicSample{
		Elapsed:    elapsed,
		Entities:   run.lastEntities,
		Process:    process,
		RoomCount:  roomCount,
		TickWindow: run.window.snapshot(),
	})
	run.window = tickAccumulator{}
	run.lastSampleElapsed = elapsed
}

func validIdentifier(value string) bool {
	return value != "" && len(value) <= maxIdentifierLength
}

func maxDuration(duration time.Duration) time.Duration {
	if duration < 0 {
		return 0
	}
	return duration
}

type tickAccumulator struct {
	count uint64
	min   time.Duration
	max   time.Duration
	total time.Duration
}

func (accumulator *tickAccumulator) add(duration time.Duration) {
	if accumulator.count == 0 || duration < accumulator.min {
		accumulator.min = duration
	}
	if accumulator.count == 0 || duration > accumulator.max {
		accumulator.max = duration
	}
	accumulator.count++
	accumulator.total += duration
}

func (accumulator tickAccumulator) snapshot() TickSummary {
	average := time.Duration(0)
	if accumulator.count > 0 {
		average = accumulator.total / time.Duration(accumulator.count)
	}
	return TickSummary{Count: accumulator.count, Minimum: accumulator.min, Maximum: accumulator.max, Total: accumulator.total, Average: average}
}

func cloneReport(report ServerReport) ServerReport {
	report.Samples = append([]PeriodicSample(nil), report.Samples...)
	report.Packets = append([]PacketSummary(nil), report.Packets...)
	report.Receiver.Lanes = append([]ReceiverLaneSummary(nil), report.Receiver.Lanes...)
	return report
}
