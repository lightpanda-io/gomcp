package main

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lightpanda-io/gomcp/mcp"
)

var InvalidSessionId = errors.New("invalid session id")

type SessionId uuid.UUID

func (id SessionId) String() string {
	return uuid.UUID(id).String()
}

func (id *SessionId) Set(v string) error {
	if len(v) == 0 {
		return InvalidSessionId
	}

	vv, err := uuid.Parse(v)
	if err != nil {
		return InvalidSessionId
	}

	*id = SessionId(vv)
	return nil
}

type Sessions struct {
	sync.Mutex
	s map[SessionId]*Session
}

func NewSessions() *Sessions {
	return &Sessions{
		s: make(map[SessionId]*Session),
	}
}

func (ss *Sessions) Add(s *Session) {
	ss.Lock()
	ss.s[s.id] = s
	ss.Unlock()
}

func (ss *Sessions) Get(id SessionId) (*Session, bool) {
	ss.Lock()
	s, ok := ss.s[id]
	ss.Unlock()
	return s, ok
}

func (ss *Sessions) Remove(id SessionId) {
	ss.Lock()
	delete(ss.s, id)
	ss.Unlock()
}

type Session struct {
	sync.Mutex
	id        SessionId
	creq      chan mcp.Request
	createdAt time.Time
}

func NewSession() *Session {
	return &Session{
		id:        SessionId(uuid.New()),
		creq:      make(chan mcp.Request),
		createdAt: time.Now(),
	}
}

func (s *Session) Close() {
	close(s.creq)
}

func (s *Session) Requests() chan mcp.Request {
	return s.creq
}
