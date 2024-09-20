//go:build !tinygo

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

type lastView struct {
	view         string
	level        int
	showChildren bool
}

type session struct {
	sessionId  string
	conn       *websocket.Conn
	LastUpdate time.Time
	LastViews  map[string]lastView // key: device Id
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
		LastViews:  make(map[string]lastView),
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

	if s, ok := sessions[sessionId]; ok {
		s.conn = conn
		s.LastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, conn)
		sessionCount += 1
	}
}

func sessionUpdate(sessionId string) bool {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if s, ok := sessions[sessionId]; ok {
		s.LastUpdate = time.Now()
		return true
	}

	// Session expired
	return false
}

func _sessionSave(sessionId, deviceId, view string, level int, showChildren bool) {

	if s, ok := sessions[sessionId]; ok {
		s.LastUpdate = time.Now()
		lastView := s.LastViews[deviceId]
		lastView.view = view
		lastView.level = level
		lastView.showChildren = showChildren
		s.LastViews[deviceId] = lastView
	}
}

func _sessionLastView(sessionId, deviceId string) (lastView lastView, err error) {
	s, ok := sessions[sessionId]
	if !ok {
		err = fmt.Errorf("Invalid session %s", sessionId)
		return
	}
	lastView, ok = s.LastViews[deviceId]
	if !ok {
		err = fmt.Errorf("Session %s: invalid device Id %s", sessionId, deviceId)
	}
	return
}

func (s session) renderPkt(pkt *Packet) {
	var buf bytes.Buffer
	if err := deviceRenderPkt(&buf, &s, pkt); err != nil {
		fmt.Println("\nError rendering pkt:", err, "\n")
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionsRoute(pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, s := range sessions {
		if s.conn != nil {
			//fmt.Println("=== sessionsRoute", pkt)
			s.renderPkt(pkt)
		}
	}
}

func sessionRoute(sessionId string, pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		if s.conn != nil {
			//fmt.Println("=== sessionRoute", pkt)
			s.renderPkt(pkt)
		}
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
		for sessionId, s := range sessions {
			if time.Since(s.LastUpdate) > minute {
				delete(sessions, sessionId)
				sessionCount -= 1
			}
		}
		sessionsMu.Unlock()
	}
}
