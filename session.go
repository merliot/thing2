package thing2

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

//go:embed template/sessions.tmpl
var sessionsTemplate string

type session struct {
	sessionId  string
	conn       *websocket.Conn
	LastUpdate time.Time
	LastView   map[string]string // key: device Id; value: last view
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex
var sessionCount int32
var sessionCountMax = int32(1000)

func init() {
	go gcSessions()
}

func _newSession(sessionId string, conn *websocket.Conn) *session {
	return &session{
		sessionId:  sessionId,
		conn:       conn,
		LastUpdate: time.Now(),
		LastView:   make(map[string]string),
	}
}

func newSession() (string, bool) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if sessionCount >= sessionCountMax {
		return "", false
	}

	sessionId := uuid.New().String()
	sessions[sessionId] = _newSession(sessionId, nil)
	sessionCount += 1

	return sessionId, true
}

func (s session) Age() string {
	return time.Since(s.LastUpdate).String()
}

func sessionConn(sessionId string, conn *websocket.Conn) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		session.conn = conn
		session.LastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, conn)
		sessionCount += 1
	}
}

func sessionUpdate(sessionId string) bool {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		session.LastUpdate = time.Now()
		return true
	}

	// Session expired
	return false
}

func _sessionDeviceSave(sessionId, deviceId, view string) {

	println("------- SAVE DEVICE", deviceId, view)
	if session, ok := sessions[sessionId]; ok {
		session.LastUpdate = time.Now()
		session.LastView[deviceId] = view
	}
}

func sessionDeviceSave(sessionId, deviceId, view string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[sessionId]; ok {
		for deviceId, _ := range session.LastView {
			delete(session.LastView, deviceId)
		}
	}

	_sessionDeviceSave(sessionId, deviceId, view)
}

func (s session) renderPkt(pkt *Packet) {
	fmt.Println("XXXXXXXXXX s.renderPkt", pkt)
	view, ok := s.LastView[pkt.Dst]
	if ok {
		var buf bytes.Buffer
		if err := deviceRenderPkt(&buf, s.sessionId, pkt.Dst, view, pkt); err != nil {
			fmt.Println("Error rendering pkt:", err)
			return
		}
		websocket.Message.Send(s.conn, string(buf.Bytes()))
	}
}

func (s session) render(deviceId string) {
	view, ok := s.LastView[deviceId]
	if ok {
		var buf bytes.Buffer
		if err := deviceRender(&buf, s.sessionId, deviceId, view); err != nil {
			fmt.Println("Error rendering device:", err)
			return
		}
		websocket.Message.Send(s.conn, string(buf.Bytes()))
	}
}

func sessionDeviceRender(sessionId, deviceId string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[sessionId]; ok {
		if session.conn == nil {
			return
		}
		session.render(deviceId)
	}
}

func sessionsRoute(pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, session := range sessions {
		if session.conn == nil {
			continue
		}
		fmt.Println("=== sessionsRoute", pkt)
		session.renderPkt(pkt)
	}
}

func (s session) _renderUpdate(deviceId, template string, pageVars pageVars) {
	var buf bytes.Buffer
	if err := _deviceRenderUpdate(&buf, deviceId, template, pageVars); err != nil {
		fmt.Println("Error rendering template:", err)
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionsRouteUpdate(deviceId, template string, pageVars pageVars) {
	println("sessionsRouteUpdate")

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, session := range sessions {
		if session.conn == nil {
			continue
		}
		session._renderUpdate(deviceId, template, pageVars)
	}
}

func sessionsShow(w http.ResponseWriter, r *http.Request) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	templateShow(w, sessionsTemplate, sessions)
}

func gcSessions() {
	minute := 1 * time.Minute
	ticker := time.NewTicker(minute)
	defer ticker.Stop()
	for range ticker.C {
		sessionsMu.Lock()
		for sessionId, session := range sessions {
			if time.Since(session.LastUpdate) > minute {
				delete(sessions, sessionId)
				sessionCount -= 1
			}
		}
		sessionsMu.Unlock()
	}
}
