package networking

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

var nextSessionID atomic.Uint64

type webSocketSession struct {
	conn                        *websocket.Conn
	sessionID                   string
	currentRoomID               string
	currentGamePlayerID         string
	room                        *rooms.Room
	rooms                       *rooms.RoomManager
	outbound                    chan []byte
	identity                    SessionIdentity
	authVerifier                TokenVerifier
	matchResultReporter         rooms.MatchResultReporter
	realtimeState               realtime.RealtimeSessionState
	debugShapeCatalogSentRoomID string
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
		identity:            NewGuestSessionIdentity(),
		authVerifier:        authVerifier,
		matchResultReporter: reporter,
	}
}

func (session *webSocketSession) SessionIdentity() SessionIdentity {
	return session.identity
}

func (session *webSocketSession) SetAuthenticatedAccountIdentity(userID int64, accountID string, displayName string) {
	session.identity = NewAuthenticatedAccountIdentity(userID, accountID, displayName)
}

func (session *webSocketSession) hasReadyWebRTCGameplayTransport() bool {
	return session.webrtcTransport != nil && session.webrtcTransport.Ready()
}

func (session *webSocketSession) resetDebugShapeCatalogSent() {
	session.debugShapeCatalogSentRoomID = ""
}

func (session *webSocketSession) ensureWebRTCTransport() *WebRTCTransport {
	if session.webrtcTransport != nil {
		return session.webrtcTransport
	}

	peer := NewWebRTCTransport(WebRTCSignalHooks{
		OnLocalICECandidate: func(media string, index int, name string) {
			session.enqueueWebRTCICECandidate(media, index, name)
		},
		OnReady: func() {
			session.enqueueWebRTCReady()
			if session.webrtcTransport != nil {
				if err := session.webrtcTransport.SendSmoke("server-ready", "server smoke peer ready"); err != nil {
					logging.Network.Warn("websocket webrtc smoke send failed",
						logging.FieldError, err,
						logging.FieldRoomID, session.currentRoomID,
						logging.FieldPlayerID, session.currentGamePlayerID,
						"session_id", session.sessionID,
					)
				}
			}
		},
		OnPacketReceived: func(packet map[string]any) {
			session.handleWebRTCPacket(packet)
		},
	})
	session.webrtcTransport = peer
	return peer
}

func (session *webSocketSession) clearWebRTCTransport() {
	if session.webrtcTransport == nil {
		return
	}
	if err := session.webrtcTransport.Close(); err != nil {
		logging.Network.Warn("websocket webrtc smoke peer close failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
	}
	session.webrtcTransport = nil
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
		logging.Network.Warn("websocket packet encode failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
		return
	}
	session.outbound <- encoded
}

func (session *webSocketSession) HandleWebRTCOffer(descriptionType string, sdp string) {
	peer := session.ensureWebRTCTransport()
	answer, err := peer.HandleOffer(descriptionType, sdp)
	if err != nil {
		logging.Network.Warn("websocket webrtc offer handling failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
		session.clearWebRTCTransport()
		return
	}
	session.enqueueWebRTCAnswer(answer.DescriptionType, answer.SDP)
}

func (session *webSocketSession) HandleWebRTCIceCandidate(media string, index int, name string) {
	if session.webrtcTransport == nil {
		logging.Network.Debug("websocket webrtc ice candidate ignored before offer",
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
		return
	}
	if err := session.webrtcTransport.AddRemoteCandidate(media, index, name); err != nil {
		logging.Network.Warn("websocket webrtc ice candidate handling failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
	}
}

func (session *webSocketSession) HandleWebRTCSmoke(smokeID string, origin string, message string) {
	logging.Network.Info("websocket webrtc smoke received",
		"session_id", session.sessionID,
		"smoke_id", smokeID,
		"origin", origin,
		"message", message,
	)
}

func (session *webSocketSession) handleWebRTCPacket(packet map[string]any) {
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
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		"session_id", session.sessionID,
		"smoke_id", smokeID,
		"origin", origin,
	)
	if session.webrtcTransport == nil {
		return
	}
	if err := session.webrtcTransport.SendSmoke(smokeID, "server reply"); err != nil {
		logging.Network.Warn("websocket webrtc smoke reply failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
	}
}

func (session *webSocketSession) HandleWebRTCFailed(errorCode string, message string) {
	logging.Network.Info("websocket webrtc failed received",
		"session_id", session.sessionID,
		"error_code", errorCode,
		"message", message,
	)
	session.clearWebRTCTransport()
}
