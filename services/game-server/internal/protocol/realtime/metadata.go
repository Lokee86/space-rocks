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



func (metadata Metadata) WithChunk(chunkIndex int, chunkCount int) Metadata {
	metadata.ChunkIndex = chunkIndex
	metadata.ChunkCount = chunkCount
	metadata.IsFinalChunk = chunkIndex == chunkCount-1
	return metadata
}
