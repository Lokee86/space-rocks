package realtime

type Lane string

type SnapshotKind string

type Metadata struct {
	Lane           Lane
	Sequence       int
	BaselineID     string
	SnapshotID     string
	ServerSentMsec int
	SnapshotKind   SnapshotKind
	ChunkIndex     int
	ChunkCount     int
	IsFinalChunk   bool
}

type WorldFullPacket struct {
	Type      string
	Metadata  Metadata
	Ships     []WorldShipRecord
	Bullets   []WorldBulletRecord
	Asteroids []WorldAsteroidRecord
	Pickups   []WorldPickupRecord
}

type OverlayFullPacket struct {
	Type     string
	Metadata Metadata
	Receiver OverlayReceiverRecord
}

type SessionFullPacket struct {
	Type           string
	Metadata       Metadata
	Players        []SessionPlayerRecord
	PlayerLifecycle []SessionLifecycleRecord
	TotalAsteroids int
}

type EventBatchPacket struct {
	Type     string
	Metadata Metadata
	Batch    EventBatchRecord
}



func FullBaselineID(lane Lane, sequence int) string {
	return string(lane) + "-baseline-" + itoa(sequence)
}

func DeltaSnapshotID(lane Lane, sequence int) string {
	return string(lane) + "-snapshot-" + itoa(sequence)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}

	negative := value < 0
	if negative {
		value = -value
	}

	var digits [20]byte
	idx := len(digits)
	for value > 0 {
		idx--
		digits[idx] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		idx--
		digits[idx] = '-'
	}
	return string(digits[idx:])
}

func (metadata Metadata) WithChunk(chunkIndex int, chunkCount int) Metadata {
	metadata.ChunkIndex = chunkIndex
	metadata.ChunkCount = chunkCount
	metadata.IsFinalChunk = chunkIndex == chunkCount-1
	return metadata
}
