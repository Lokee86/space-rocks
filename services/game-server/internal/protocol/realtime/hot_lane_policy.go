package realtime

type HotLaneOffloadPolicy struct {
	AsteroidHotLaneEntityBudget int
	BulletHotLaneEntityBudget   int
	TargetEncodedPacketBytes    int
	HardEncodedPacketBytes      int
	MTUSafePacketBytes          int
}

const (
	DefaultAsteroidHotLaneEntityBudget = 48
	DefaultBulletHotLaneEntityBudget   = 48
	DefaultTargetEncodedPacketBytes    = 800
	DefaultHardEncodedPacketBytes      = 1200
	DefaultMTUSafePacketBytes          = 1500
)

func DefaultHotLaneOffloadPolicy() HotLaneOffloadPolicy {
	return HotLaneOffloadPolicy{
		AsteroidHotLaneEntityBudget: DefaultAsteroidHotLaneEntityBudget,
		BulletHotLaneEntityBudget:   DefaultBulletHotLaneEntityBudget,
		TargetEncodedPacketBytes:    DefaultTargetEncodedPacketBytes,
		HardEncodedPacketBytes:      DefaultHardEncodedPacketBytes,
		MTUSafePacketBytes:          DefaultMTUSafePacketBytes,
	}
}

func (p HotLaneOffloadPolicy) WorldHotEntityBudget() int {
	if p.AsteroidHotLaneEntityBudget < p.BulletHotLaneEntityBudget {
		return p.AsteroidHotLaneEntityBudget
	}
	return p.BulletHotLaneEntityBudget
}

func (p HotLaneOffloadPolicy) AsteroidFullOffloadThreshold() int {
	return p.AsteroidHotLaneEntityBudget * 2
}

func (p HotLaneOffloadPolicy) BulletFullOffloadThreshold() int {
	return p.BulletHotLaneEntityBudget * 2
}

func (p HotLaneOffloadPolicy) AsteroidNeedsChunkingThreshold() int {
	return p.AsteroidHotLaneEntityBudget * 3
}

func (p HotLaneOffloadPolicy) BulletNeedsChunkingThreshold() int {
	return p.BulletHotLaneEntityBudget * 3
}
