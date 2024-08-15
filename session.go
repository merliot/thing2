package thing2

import (
	"bytes"
	_ "embed"
	"net/http"
	"net/url"
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
	LastURL    map[string]*url.URL // key: device Id; value: last URL
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex

func init() {
	go gcSessions()
}

func _newSession(sessionId string, conn *websocket.Conn) *session {
	return &session{
		sessionId:  sessionId,
		conn:       conn,
		LastUpdate: time.Now(),
		LastURL:    make(map[string]*url.URL),
	}
}

func newSession() string {
	// TODO check and limit number of sessions

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	sessionId := uuid.New().String()
	sessions[sessionId] = _newSession(sessionId, nil)
	return sessionId
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
	}
}

func sessionUpdate(sessionId string) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		session.LastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, nil)
	}
}

func _sessionDeviceSave(sessionId, deviceId string, url *url.URL) {

	println("------- SAVE DEVICE", deviceId, url.String())
	if session, ok := sessions[sessionId]; ok {
		session.LastUpdate = time.Now()
		session.LastURL[deviceId] = url
	}
}

func sessionDeviceSave(sessionId, deviceId string, url *url.URL) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[sessionId]; ok {
		for deviceId, _ := range session.LastURL {
			delete(session.LastURL, deviceId)
		}
	}

	_sessionDeviceSave(sessionId, deviceId, url)
}

func (s session) _render(deviceId string, url *url.URL) {
	var buf bytes.Buffer
	if err := _deviceRender(&buf, s.sessionId, deviceId, url); err != nil {
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func (s session) render(deviceId string, url *url.URL) {
	var buf bytes.Buffer
	if err := deviceRender(&buf, s.sessionId, deviceId, url); err != nil {
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionDeviceRender(sessionId, deviceId string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[sessionId]; ok {
		if session.conn == nil {
			return
		}
		if url, ok := session.LastURL[deviceId]; ok {
			session.render(deviceId, url)
		}
	}
}

func sessionsRoute(deviceId string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, session := range sessions {
		if session.conn == nil {
			continue
		}
		if url, ok := session.LastURL[deviceId]; ok {
			session._render(deviceId, url)
		}

	}
}

func sessionsShow(w http.ResponseWriter, r *http.Request) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	templateShow(w, sessionsTemplate, sessions)
}

func gcSessions() {
	dur := 1 * time.Minute
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for range ticker.C {
		sessionsMu.Lock()
		for sessionId, session := range sessions {
			if time.Since(session.LastUpdate) > dur {
				delete(sessions, sessionId)
			}
		}
		sessionsMu.Unlock()
	}
}
