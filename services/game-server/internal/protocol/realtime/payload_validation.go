package realtime

import (
	"fmt"
	"reflect"
)

var realtimeLanePayloadFamilyMatrix = map[Lane]map[RealtimeLaneCandidateKind]map[string]struct{}{
	LaneWorld:   {RealtimeLaneCandidateKindFull: {PacketFamilyWorldFull: {}}, RealtimeLaneCandidateKindDelta: {PacketFamilyWorldDelta: {}}},
	LaneOverlay: {RealtimeLaneCandidateKindFull: {PacketFamilyOverlayFull: {}}, RealtimeLaneCandidateKindDelta: {PacketFamilyOverlayDelta: {}}},
	LaneSession: {RealtimeLaneCandidateKindFull: {PacketFamilySessionFull: {}}, RealtimeLaneCandidateKindDelta: {PacketFamilySessionDelta: {}}},
	LaneShips: {
		RealtimeLaneCandidateKindDelta: {
			PacketFamilyShipDelta:     {},
			PacketFamilyPlayerLocator: {},
		},
	},
	LaneShipsLifecycle: {
		RealtimeLaneCandidateKindDelta: {PacketFamilyShipsLifecycle: {}},
	},
	LaneAsteroids: {
		RealtimeLaneCandidateKindDelta: {PacketFamilyAsteroidDelta: {}},
	},
	LaneAsteroidsLifecycle: {
		RealtimeLaneCandidateKindDelta: {PacketFamilyAsteroidsLifecycle: {}},
	},
	LaneBullets: {
		RealtimeLaneCandidateKindDelta: {PacketFamilyBulletDelta: {}},
	},
	LaneBulletsLifecycle: {
		RealtimeLaneCandidateKindDelta: {PacketFamilyBulletsLifecycle: {}},
	},
	LaneEvent: {RealtimeLaneCandidateKindEventBatch: {PacketFamilyEventBatch: {}}},
}

var realtimeLanePayloadTypes = map[reflect.Type]struct{}{
	reflect.TypeOf(WorldFullPacket{}):         {},
	reflect.TypeOf(WorldWireFullPacket{}):     {},
	reflect.TypeOf(WorldDeltaPacket{}):        {},
	reflect.TypeOf(WorldWireDeltaPacket{}):    {},
	reflect.TypeOf(OverlayFullPacket{}):       {},
	reflect.TypeOf(OverlayWireFullPacket{}):   {},
	reflect.TypeOf(OverlayLaneDelta{}):        {},
	reflect.TypeOf(OverlayWireLaneDelta{}):    {},
	reflect.TypeOf(SessionFullPacket{}):       {},
	reflect.TypeOf(SessionWireFullPacket{}):   {},
	reflect.TypeOf(SessionLaneDelta{}):        {},
	reflect.TypeOf(SessionWireLaneDelta{}):    {},
	reflect.TypeOf(ShipWireDeltaPacket{}):     {},
	reflect.TypeOf(PlayerLocatorPacket{}):     {},
	reflect.TypeOf(AsteroidWireDeltaPacket{}): {},
	reflect.TypeOf(BulletWireDeltaPacket{}):   {},
	reflect.TypeOf(EventBatchPacket{}):        {},
}

func ValidateRealtimeLanePayload(payload RealtimeLanePayload) error {
	if payload == nil || isTypedNil(payload) {
		return fmt.Errorf("realtime lane payload is nil")
	}
	lane := payload.Lane()
	kind := payload.CandidateKind()
	family := payload.PacketFamily()
	if lane == "" {
		return fmt.Errorf("realtime lane payload has empty lane")
	}
	if kind == "" {
		return fmt.Errorf("realtime lane payload has empty candidate kind")
	}
	if family == "" {
		return fmt.Errorf("realtime lane payload has empty packet family")
	}
	metadata, ok := payload.LaneMetadata()
	if !ok {
		return fmt.Errorf("realtime lane payload is missing metadata")
	}
	if metadata.Lane == "" {
		return fmt.Errorf("realtime lane payload metadata has empty lane")
	}
	if metadata.Lane != lane {
		return fmt.Errorf("realtime lane payload lane %q does not match metadata lane %q", lane, metadata.Lane)
	}

	kinds, ok := realtimeLanePayloadFamilyMatrix[lane]
	if !ok {
		return fmt.Errorf("unsupported realtime lane %q", lane)
	}
	families, ok := kinds[kind]
	if !ok {
		return fmt.Errorf("unsupported realtime lane candidate kind %q for lane %q", kind, lane)
	}
	if _, ok := families[family]; !ok {
		return fmt.Errorf("unsupported packet family %q for lane %q and kind %q", family, lane, kind)
	}
	if _, ok := realtimeLanePayloadTypes[reflect.TypeOf(payload)]; !ok {
		return fmt.Errorf("unsupported realtime lane payload type %T", payload)
	}
	return nil
}

func isTypedNil(value any) bool {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
