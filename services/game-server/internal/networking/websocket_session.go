package networking

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/gorilla/websocket"
)

var nextSessionID atomic.Uint64

type webSocketSession struct {
	mu                  sync.RWMutex
	conn                *websocket.Conn
	sessionID           string
	context             SessionContext
	rooms               *rooms.RoomManager
	outbound            chan []byte
	outboundOverflowOnce sync.Once
	resyncRequests      chan queuedResyncRequest
	identity            SessionIdentity
	authVerifier        TokenVerifier
	matchResultReporter rooms.MatchResultReporter
	// realtimeState is owned exclusively by the write loop and intentionally not guarded by mu.
	realtimeState               realtime.RealtimeSessionState
	debugShapeCatalogSentRoomID string
	firstPacketMatchID          string
	firstInputPacketLogged      bool
	firstRespawnPacketLogged    bool
	webrtcTransport             *WebRTCTransport
}

func newWebSocketSession(conn *websocket.Conn, roomManager *rooms.RoomManager, authVerifier TokenVerifier, reporter rooms.MatchResultReporter) *webSocketSession {
	sessionNumber := nextSessionID.Add(1)
	if reporter == nil {
		reporter = rooms.NoopMatchResultReporter{}
	}

	return &webSocketSession{
		conn:                conn,
		sessionID:           "session-" + strconv.FormatUint(sessionNumber, 10),
		rooms:               roomManager,
		outbound:            make(chan []byte, 16),
		resyncRequests:      make(chan queuedResyncRequest, 4),
		identity:            NewGuestSessionIdentity(),
		authVerifier:        authVerifier,
		matchResultReporter: reporter,
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

func (session *webSocketSession) hasReadyWebRTCGameplayTransport() bool {
	transport := session.webRTCTransportSnapshot()
	return transport != nil && transport.Ready()
}

func (session *webSocketSession) webRTCTransportSnapshot() *WebRTCTransport {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.webrtcTransport
}

func (session *webSocketSession) resetDebugShapeCatalogSent() {
	session.mu.Lock()
	session.debugShapeCatalogSentRoomID = ""
	session.mu.Unlock()
}

func (session *webSocketSession) debugShapeCatalogSentFor(roomID string) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.debugShapeCatalogSentRoomID == roomID
}

func (session *webSocketSession) markDebugShapeCatalogSent(context SessionContext) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.context != context {
		return false
	}
	session.debugShapeCatalogSentRoomID = context.RoomID
	return true
}

func (session *webSocketSession) ensureWebRTCTransport() *WebRTCTransport {
	if transport := session.webRTCTransportSnapshot(); transport != nil {
		return transport
	}

	peer := NewWebRTCTransport(WebRTCSignalHooks{
		OnLocalICECandidate: func(media string, index int, name string) {
			session.enqueueWebRTCICECandidate(media, index, name)
		},
		OnReady: func() {
			context := session.sessionContext()
			session.enqueueWebRTCReady()
			transport := session.webRTCTransportSnapshot()
			if transport != nil {
				if err := transport.SendSmoke("server-ready", "server smoke peer ready"); err != nil {
					logging.Network.Warn("websocket webrtc smoke send failed",
						logging.FieldError, err,
						logging.FieldRoomID, context.RoomID,
						logging.FieldPlayerID, context.GamePlayerID,
						"session_id", session.sessionID,
					)
				}
			}
		},
		OnPacketReceived: func(packet map[string]any) {
			session.handleWebRTCPacket(packet)
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

func (session *webSocketSession) clearWebRTCTransport() {
	session.mu.Lock()
	transport := session.webrtcTransport
	session.webrtcTransport = nil
	session.mu.Unlock()
	if transport == nil {
		return
	}
	context := session.sessionContext()
	if err := transport.Close(); err != nil {
		logging.Network.Warn("websocket webrtc smoke peer close failed",
			logging.FieldError, err,
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
	}
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
	channels := make([]map[string]any, 0, len(webRTCGameplayChannelSpecs()))
	for _, spec := range webRTCGameplayChannelSpecs() {
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
		logging.Network.Warn("websocket packet encode failed",
			logging.FieldError, err,
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
		return
	}
	session.enqueue(encoded)
}

func (session *webSocketSession) HandleWebRTCOffer(descriptionType string, sdp string) {
	peer := session.ensureWebRTCTransport()
	answer, err := peer.HandleOffer(descriptionType, sdp)
	if err != nil {
		context := session.sessionContext()
		logging.Network.Warn("websocket webrtc offer handling failed",
			logging.FieldError, err,
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
		session.clearWebRTCTransport()
		return
	}
	session.enqueueWebRTCAnswer(answer.DescriptionType, answer.SDP)
}

func (session *webSocketSession) HandleWebRTCIceCandidate(media string, index int, name string) {
	context := session.sessionContext()
	transport := session.webRTCTransportSnapshot()
	if transport == nil {
		logging.Network.Debug("websocket webrtc ice candidate ignored before offer",
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
		return
	}
	if err := transport.AddRemoteCandidate(media, index, name); err != nil {
		logging.Network.Warn("websocket webrtc ice candidate handling failed",
			logging.FieldError, err,
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
	}
}

func (session *webSocketSession) HandleWebRTCSmoke(smokeID string, origin string, message string) {
	context := session.sessionContext()
	logging.Network.Info("websocket webrtc smoke received",
		logging.FieldRoomID, context.RoomID,
		logging.FieldPlayerID, context.GamePlayerID,
		"session_id", session.sessionID,
		"smoke_id", smokeID,
		"origin", origin,
		"message", message,
	)
}

func (session *webSocketSession) handleWebRTCPacket(packet map[string]any) {
	context := session.sessionContext()
	packetType := fmt.Sprint(packet["type"])
	if packetType != "webrtc_smoke" {
		logging.Network.Debug("websocket webrtc packet ignored",
			"session_id", session.sessionID,
			"type", packetType,
		)
		return
	}

	smokeID := fmt.Sprint(packet["smoke_id"])
	origin := fmt.Sprint(packet["origin"])
	logging.Network.Info("websocket webrtc smoke packet received",
		logging.FieldRoomID, context.RoomID,
		logging.FieldPlayerID, context.GamePlayerID,
		"session_id", session.sessionID,
		"smoke_id", smokeID,
		"origin", origin,
	)
	transport := session.webRTCTransportSnapshot()
	if transport == nil {
		return
	}
	if err := transport.SendSmoke(smokeID, "server reply"); err != nil {
		logging.Network.Warn("websocket webrtc smoke reply failed",
			logging.FieldError, err,
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
	}
}

func (session *webSocketSession) HandleWebRTCFailed(errorCode string, message string) {
	context := session.sessionContext()
	logging.Network.Info("websocket webrtc failed received",
		logging.FieldRoomID, context.RoomID,
		logging.FieldPlayerID, context.GamePlayerID,
		"session_id", session.sessionID,
		"error_code", errorCode,
		"message", message,
	)
	session.clearWebRTCTransport()
}
