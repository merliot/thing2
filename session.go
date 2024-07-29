package thing2

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type session struct {
	id         string
	conn       *websocket.Conn
	lastUpdate time.Time
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex

func init() {
	go gcSessions()
}

func newSession() string {
	// TODO check and limit size of sessions
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	id := uuid.New().String()
	sessions[id] = &session{
		id:         id,
		lastUpdate: time.Now(),
	}
	println("newSession", id)
	return id
}

func sessionConn(id string, conn *websocket.Conn) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if s, ok := sessions[id]; ok {
		println("sessionConn", id, conn)
		s.conn = conn
		s.lastUpdate = time.Now()
	} else {
		sessions[id] = &session{
			id:         id,
			conn:       conn,
			lastUpdate: time.Now(),
		}
	}
}

func sessionUpdate(id string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if session, ok := sessions[id]; ok {
		println("sessionUpdate", id)
		session.lastUpdate = time.Now()
	}
}

func gcSessions() {
	dur := 1 * time.Minute
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for range ticker.C {
		println("gcSessions ticked")
		sessionsMu.Lock()
		for id, session := range sessions {
			println("gcSessions considering", id)
			if time.Since(session.lastUpdate) > dur {
				println("gcSessions", id)
				delete(sessions, id)
			}
		}
		sessionsMu.Unlock()
	}
}
