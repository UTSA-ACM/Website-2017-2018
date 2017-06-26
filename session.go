package main

import (
	"math/rand"
	"sync"
	"time"
)

var (
	manager SessionManager
)

// session stores a username and the expiration time
type session struct {
	username string
	expires  time.Time
}

// SessionManager  Stored string is a session struct, and the key string is the session id
type SessionManager struct {
	lock sync.RWMutex
	m    map[string]session
}

func init() {
	manager.m = make(map[string]session)
}

func (sm *SessionManager) Write(token, username string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	sm.m[token] = session{username, time.Now().Add(time.Hour * 72)}
}

func (sm *SessionManager) Read(token string) string {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	sess, ok := sm.m[token]

	if !ok {
		return ""
	}

	if sess.expires.Sub(time.Now()) < 1 {
		delete(sm.m, token)
		return ""
	}

	return sess.username
}

func newSession(username string) string {
	token := newToken()
	manager.Write(token, username)

	return token
}

func checkSession(token string) string {
	return manager.Read(token)
}

func newToken() string {
	b := make([]byte, 40)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
