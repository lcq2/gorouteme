package session

import (
	"container/list"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Session struct {
	sid        string
	accessTime time.Time
	value      map[interface{}]interface{}
	lock       sync.RWMutex
}

type SessionManager struct {
	sessions    map[string]*list.Element
	list        list.List
	cookieName  string
	maxLifeTime int64
	lock        sync.RWMutex
}

func (session *Session) touch() {
	session.lock.Lock()
	defer session.lock.Unlock()
	session.accessTime = time.Now()
}

func (session *Session) Set(key, value interface{}) {
	session.lock.Lock()
	defer session.lock.Unlock()
	session.value[key] = value
}

func (session *Session) Get(key interface{}) interface{} {
	session.lock.RLock()
	defer session.lock.RUnlock()
	if v, ok := session.value[key]; ok {
		return v
	}
	return nil
}

func (session *Session) Delete(key interface{}) {
	session.lock.Lock()
	defer session.lock.Unlock()
	delete(session.value, key)
}

func (session *Session) Flush() {
	session.lock.Lock()
	defer session.lock.Unlock()
	session.value = make(map[interface{}]interface{})
}

func (session *Session) UserLoggedIn() bool {
	session.lock.RLock()
	defer session.lock.RUnlock()
	if v, ok := session.value["loggedIn"]; ok {
		return v.(bool)
	}
	return false
}

func (session *Session) SessionId() string {
	return session.sid
}

var sessionManager *SessionManager

func InitSessionManager(cookieName string, maxLifeTime int64) {
	sessionManager = &SessionManager{cookieName: cookieName, maxLifeTime: maxLifeTime, sessions: make(map[string]*list.Element)}
	go sessionManager.ExpireSessions()
}

func Manager() *SessionManager {
	return sessionManager
}

func (manager *SessionManager) sessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *SessionManager) ExpireSessions() {
	manager.lock.RLock()
	for {
		element := manager.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*Session).accessTime.Unix() + manager.maxLifeTime) < time.Now().Unix() {
			manager.lock.RUnlock()
			manager.lock.Lock()
			manager.list.Remove(element)
			delete(manager.sessions, element.Value.(*Session).sid)
			manager.lock.Unlock()
			manager.lock.RLock()
		} else {
			break
		}
	}
	manager.lock.RUnlock()
	time.AfterFunc(time.Duration(manager.maxLifeTime)*time.Second, func() { manager.ExpireSessions() })
}

func (manager *SessionManager) sidFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return "", errors.New("Invalid session")
	}
	return url.QueryUnescape(cookie.Value)
}

func (manager *SessionManager) SessionNew(w http.ResponseWriter, r *http.Request) *Session {
	sid := manager.sessionID()
	session := &Session{sid: sid, accessTime: time.Now(), value: make(map[interface{}]interface{})}
	manager.lock.Lock()
	manager.sessions[sid] = manager.list.PushFront(session)
	manager.lock.Unlock()

	cookie := &http.Cookie{
		Name:     manager.cookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
	return session
}

func (manager *SessionManager) SessionGet(r *http.Request) *Session {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil || sid == "" {
		return nil
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()
	if el, ok := manager.sessions[sid]; ok {
		el.Value.(*Session).accessTime = time.Now()
		manager.list.MoveToFront(el)
		return el.Value.(*Session)
	}
	return nil
}

func (manager *SessionManager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil || sid == "" {
		return
	}

	manager.lock.Lock()
	if el, ok := manager.sessions[sid]; ok {
		delete(manager.sessions, sid)
		manager.list.Remove(el)
	}
	manager.lock.Unlock()
	cookie = &http.Cookie{
		Name:     manager.cookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
