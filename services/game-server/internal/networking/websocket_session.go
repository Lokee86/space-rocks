package networking

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	toolingrouter "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var nextSessionID atomic.Uint64

type webSocketSession struct {
	mu                   sync.RWMutex
	conn                 *websocket.Conn
	sessionID            string
	connectionTraceID    string
	context              SessionContext
	rooms                *rooms.RoomManager
	outbound             chan []byte
	outboundOverflowOnce sync.Once
	resyncRequests       chan queuedResyncRequest
	identity             SessionIdentity
	authVerifier         TokenVerifier
	matchResultReporter  rooms.MatchResultReporter
	buildService         *playerbuild.Service
	buildState           sessionBuildState
	// realtimeState is owned exclusively by the write loop and intentionally not guarded by mu.
	realtimeState            realtime.RealtimeSessionState
	viewTargetPlayerID       string
	firstPacketMatchID       string
	firstInputPacketLogged   bool
	firstRespawnPacketLogged bool
	webrtcTransport          *WebRTCTransport
	webrtcBackpressureLogged map[string]bool
	packetObserver           packetObserver
	receiverObserver         receiverObserver
	toolingRouter            *toolingrouter.Router
	toolingCapabilities      toolingrouter.CapabilitySet
}

func newWebSocketSession(conn *websocket.Conn, roomManager *rooms.RoomManager, authVerifier TokenVerifier, reporter rooms.MatchResultReporter, buildServices ...*playerbuild.Service) *webSocketSession {
	var buildService *playerbuild.Service
	if len(buildServices) > 0 {
		buildService = buildServices[0]
	}
	return newWebSocketSessionWithToolingAndBuilds(conn, roomManager, authVerifier, reporter, nil, nil, buildService)
}

func newWebSocketSessionWithTooling(conn *websocket.Conn, roomManager *rooms.RoomManager, authVerifier TokenVerifier, reporter rooms.MatchResultReporter, measurementController toolingrouter.MeasurementController, telemetryProvider toolingrouter.TelemetryProvider) *webSocketSession {
	return newWebSocketSessionWithToolingAndBuilds(conn, roomManager, authVerifier, reporter, measurementController, telemetryProvider, nil)
}

func newWebSocketSessionWithToolingAndBuilds(conn *websocket.Conn, roomManager *rooms.RoomManager, authVerifier TokenVerifier, reporter rooms.MatchResultReporter, measurementController toolingrouter.MeasurementController, telemetryProvider toolingrouter.TelemetryProvider, buildService *playerbuild.Service) *webSocketSession {
	sessionNumber := nextSessionID.Add(1)
	if reporter == nil {
		reporter = rooms.NoopMatchResultReporter{}
	}

	return &webSocketSession{
		conn:                     conn,
		sessionID:                "session-" + strconv.FormatUint(sessionNumber, 10),
		connectionTraceID:        uuid.NewString(),
		rooms:                    roomManager,
		outbound:                 make(chan []byte, 16),
		resyncRequests:           make(chan queuedResyncRequest, 4),
		identity:                 NewGuestSessionIdentity(),
		authVerifier:             authVerifier,
		matchResultReporter:      reporter,
		buildService:             buildService,
		webrtcBackpressureLogged: make(map[string]bool),
		packetObserver:           packetObserverFor(measurementController),
		receiverObserver:         receiverObserverFor(measurementController),
		toolingRouter:            toolingrouter.NewRouter(measurementController, telemetryProvider),
		toolingCapabilities:      toolingrouter.NewTemporaryCapabilitySet(),
	}
}

func (session *webSocketSession) SessionIdentity() SessionIdentity {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.identity
}

func (session *webSocketSession) SetAuthenticatedAccountIdentity(userID int64, accountID string, displayName string) {
	session.mu.Lock()
	session.identity = NewAuthenticatedAccountIdentity(userID, accountID, displayName)
	session.mu.Unlock()
}

func (session *webSocketSession) SetViewTargetPlayerID(playerID string) {
	session.mu.Lock()
	session.viewTargetPlayerID = playerID
	session.mu.Unlock()
}

func (session *webSocketSession) viewTargetPlayerIDSnapshot() string {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.viewTargetPlayerID
}

func (session *webSocketSession) clearViewTargetPlayerID() {
	session.SetViewTargetPlayerID("")
}

func (session *webSocketSession) shouldLogFirstInputPacket(matchID string) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	session.resetFirstPacketLoggingLocked(matchID)
	if session.firstInputPacketLogged {
		return false
	}
	session.firstInputPacketLogged = true
	return true
}

func (session *webSocketSession) shouldLogFirstRespawnPacket(matchID string) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	session.resetFirstPacketLoggingLocked(matchID)
	if session.firstRespawnPacketLogged {
		return false
	}
	session.firstRespawnPacketLogged = true
	return true
}

func (session *webSocketSession) resetFirstPacketLoggingLocked(matchID string) {
	if session.firstPacketMatchID == matchID {
		return
	}
	session.firstPacketMatchID = matchID
	session.firstInputPacketLogged = false
	session.firstRespawnPacketLogged = false
}

func (session *webSocketSession) shouldLogWebRTCBackpressure(lane string) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.webrtcBackpressureLogged == nil {
		session.webrtcBackpressureLogged = make(map[string]bool)
	}
	if session.webrtcBackpressureLogged[lane] {
		return false
	}
	session.webrtcBackpressureLogged[lane] = true
	return true
}

func (session *webSocketSession) hasReadyWebRTCTransport() bool {
	transport := session.webRTCTransportSnapshot()
	return transport != nil && transport.Ready()
}

func (session *webSocketSession) webRTCTransportSnapshot() *WebRTCTransport {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.webrtcTransport
}

func (session *webSocketSession) ensureWebRTCTransport() *WebRTCTransport {
	if transport := session.webRTCTransportSnapshot(); transport != nil {
		return transport
	}

	var peer *WebRTCTransport
	peer = NewWebRTCTransport(WebRTCSignalHooks{
		OnLocalICECandidate: func(media string, index int, name string) {
			if !session.isCurrentWebRTCTransport(peer) {
				return
			}
			session.enqueueWebRTCICECandidate(media, index, name)
		},
		OnReady: func() {
			if !session.isCurrentWebRTCTransport(peer) {
				return
			}
			session.enqueueWebRTCReady()
			_ = peer.SendSmoke("server-ready", "server smoke peer ready")
		},
		OnPacketReceived: func(packet map[string]any, lane string) {
			if !session.isCurrentWebRTCTransport(peer) {
				return
			}
			if lane == "tooling" {
				session.handleToolingPacket(packet)
				return
			}
			session.handleWebRTCPacket(packet)
		},
		OnChannelClosed: func(lane string) {
			if !session.isCurrentWebRTCTransport(peer) {
				return
			}
			context := session.sessionContext()
			bufferedBytes, _ := peer.BufferedAmountForLane(lane)
			logging.Emit(observability.Request{
				Event: observability.EventNamePacketRouteFailed,
				Context: observability.Context{
					TraceID:   session.connectionTraceID,
					SessionID: session.sessionID,
					RoomID:    context.RoomID,
					PlayerID:  context.GamePlayerID,
				},
				Fields: observability.Fields{
					"error_code":     "webrtc_channel_closed",
					"failure_mode":   "webrtc_channel_closed",
					"lane":           lane,
					"transport":      "webrtc",
					"buffered_bytes": bufferedBytes,
				},
			})
			if lane == "tooling" {
				session.closeTooling()
			}
			session.clearWebRTCTransportIfCurrent(peer)
		},
	})
	session.mu.Lock()
	if session.webrtcTransport != nil {
		transport := session.webrtcTransport
		session.mu.Unlock()
		_ = peer.Close()
		return transport
	}
	session.webrtcTransport = peer
	session.mu.Unlock()
	return peer
}

func (session *webSocketSession) isCurrentWebRTCTransport(expected *WebRTCTransport) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return expected != nil && session.webrtcTransport == expected
}

func (session *webSocketSession) clearWebRTCTransport() {
	session.mu.Lock()
	transport := session.webrtcTransport
	session.webrtcTransport = nil
	session.mu.Unlock()
	if transport == nil {
		return
	}
	_ = transport.Close()
}

func (session *webSocketSession) clearWebRTCTransportIfCurrent(expected *WebRTCTransport) bool {
	if expected == nil {
		return false
	}
	session.mu.Lock()
	if session.webrtcTransport != expected {
		session.mu.Unlock()
		return false
	}
	session.webrtcTransport = nil
	session.mu.Unlock()
	_ = expected.Close()
	return true
}

func (session *webSocketSession) enqueueWebRTCAnswer(descriptionType string, sdp string) {
	session.enqueuePacket(map[string]any{
		"type":             "webrtc_answer",
		"description_type": descriptionType,
		"sdp":              sdp,
	})
}

func (session *webSocketSession) enqueueWebRTCICECandidate(media string, index int, name string) {
	session.enqueuePacket(map[string]any{
		"type":  "webrtc_ice_candidate",
		"media": media,
		"index": index,
		"name":  name,
	})
}

func (session *webSocketSession) enqueueWebRTCReady() {
	channels := make([]map[string]any, 0, len(webRTCChannelSpecs()))
	for _, spec := range webRTCChannelSpecs() {
		channels = append(channels, map[string]any{
			"lane":          spec.Lane,
			"channel_label": spec.Label,
			"channel_id":    spec.ID,
		})
	}
	session.enqueuePacket(map[string]any{
		"type":     "webrtc_ready",
		"channels": channels,
	})
}

func (session *webSocketSession) enqueueWebRTCFailed(errorCode string, message string) {
	session.enqueuePacket(map[string]any{
		"type":       "webrtc_failed",
		"error_code": errorCode,
		"message":    message,
	})
}

func (session *webSocketSession) enqueuePacket(packet map[string]any) {
	encoded, err := json.Marshal(packet)
	if err != nil {
		context := session.sessionContext()
		packetType, _ := packet["type"].(string)
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    session.connectionTraceID,
				SessionID:  session.sessionID,
				RoomID:     context.RoomID,
				PlayerID:   context.GamePlayerID,
				PacketType: packetType,
			},
			Fields: observability.Fields{
				"error_code":   "websocket_control_packet_encode_failed",
				"failure_mode": "websocket_control_packet_encode_failed",
			},
		})
		return
	}
	session.enqueue(encoded)
}

func (session *webSocketSession) HandleWebRTCOffer(descriptionType string, sdp string) {
	if session.webRTCTransportSnapshot() != nil {
		session.clearWebRTCTransport()
	}
	peer := session.ensureWebRTCTransport()
	answer, err := peer.HandleOffer(descriptionType, sdp)
	if err != nil {
		context := session.sessionContext()
		logging.Emit(observability.Request{
			Event: observability.EventNamePacketRouteFailed,
			Context: observability.Context{
				TraceID:    session.connectionTraceID,
				SessionID:  session.sessionID,
				RoomID:     context.RoomID,
				PlayerID:   context.GamePlayerID,
				PacketType: "webrtc_offer",
			},
			Fields: observability.Fields{
				"error_code":   "webrtc_offer_handling_failed",
				"failure_mode": "webrtc_offer_handling_failed",
				"transport":    "webrtc",
			},
		})
		session.clearWebRTCTransport()
		return
	}
	session.enqueueWebRTCAnswer(answer.DescriptionType, answer.SDP)
}

func (session *webSocketSession) HandleWebRTCIceCandidate(media string, index int, name string) {
	context := session.sessionContext()
	transport := session.webRTCTransportSnapshot()
	if transport == nil {
		return
	}
	if err := transport.AddRemoteCandidate(media, index, name); err != nil {
		logging.Emit(observability.Request{
			Event: observability.EventNamePacketRouteFailed,
			Context: observability.Context{
				TraceID:    session.connectionTraceID,
				SessionID:  session.sessionID,
				RoomID:     context.RoomID,
				PlayerID:   context.GamePlayerID,
				PacketType: "webrtc_ice_candidate",
			},
			Fields: observability.Fields{
				"error_code":   "webrtc_ice_candidate_handling_failed",
				"failure_mode": "webrtc_ice_candidate_handling_failed",
				"transport":    "webrtc",
			},
		})
	}
}

func (session *webSocketSession) HandleWebRTCSmoke(_ string, _ string, _ string) {}

func (session *webSocketSession) handleWebRTCPacket(packet map[string]any) {
	context := session.sessionContext()
	packetType := fmt.Sprint(packet["type"])
	if packetType != "webrtc_smoke" {
		return
	}

	smokeID := fmt.Sprint(packet["smoke_id"])
	transport := session.webRTCTransportSnapshot()
	if transport == nil {
		return
	}
	if err := transport.SendSmoke(smokeID, "server reply"); err != nil {
		logging.Emit(observability.Request{
			Event: observability.EventNameGameServerWriteFailed,
			Context: observability.Context{
				TraceID:    session.connectionTraceID,
				SessionID:  session.sessionID,
				RoomID:     context.RoomID,
				PlayerID:   context.GamePlayerID,
				PacketType: "webrtc_smoke",
			},
			Fields: observability.Fields{
				"error_code":   "webrtc_smoke_reply_failed",
				"failure_mode": "webrtc_smoke_reply_failed",
				"transport":    "webrtc",
			},
		})
	}
}

func (session *webSocketSession) HandleWebRTCFailed(_ string, _ string) {
	session.clearWebRTCTransport()
}

func (session *webSocketSession) handleToolingPacket(packet map[string]any) {
	router := session.toolingRouter
	transport := session.webRTCTransportSnapshot()
	if router == nil || transport == nil {
		return
	}
	router.Handle(session.toolingContext(), transport, packet)
}

func (session *webSocketSession) writeToolingProtocolMessage() {
	router := session.toolingRouter
	transport := session.webRTCTransportSnapshot()
	if router == nil || transport == nil {
		return
	}
	router.Tick(session.toolingContext(), transport)
}

func (session *webSocketSession) closeTooling() {
	if session.toolingRouter == nil {
		return
	}
	session.toolingRouter.Close(session.toolingContext())
}

func (session *webSocketSession) toolingContext() toolingrouter.Context {
	context := session.sessionContext()
	toolingContext := toolingrouter.Context{
		SessionID:    session.sessionID,
		RoomID:       context.RoomID,
		GamePlayerID: context.GamePlayerID,
		Room:         context.Room,
		Capabilities: session.toolingCapabilities,
	}
	if context.Room == nil {
		return toolingContext
	}
	gameplayContext := context.Room.GameplayContext()
	if gameplayContext.Game == nil {
		return toolingContext
	}
	control := game.NewControl(gameplayContext.Game)
	toolingContext.CommandController = devtools.NewController(devtools.Dependencies{Target: control})
	return toolingContext
}
