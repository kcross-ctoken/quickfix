// Copyright (c) quickfixengine.org  All rights reserved.
//
// This file may be distributed under the terms of the quickfixengine.org
// license as defined by quickfixengine.org and appearing in the file
// LICENSE included in the packaging of this file.
//
// This file is provided AS IS with NO WARRANTY OF ANY KIND, INCLUDING
// THE WARRANTY OF DESIGN, MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE.
//
// See http://www.quickfixengine.org/LICENSE for licensing information.
//
// Contact ask@quickfixengine.org if any conditions of this licensing
// are not clear to you.

package quickfix

import (
	"errors"
	"sync"
)

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{sessions: make(map[SessionID]*session)}
}

type Registry struct {
	sessionsLock sync.RWMutex
	sessions     map[SessionID]*session
}

var errDuplicateSessionID = errors.New("Duplicate SessionID")
var errUnknownSession = errors.New("Unknown session")

// Messagable is a Message or something that can be converted to a Message.
type Messagable interface {
	ToMessage() *Message
}

// Send determines the session to send Messagable using header fields BeginString, TargetCompID, SenderCompID.
func (r *Registry) Send(m Messagable) (err error) {
	msg := m.ToMessage()
	var beginString FIXString
	if err := msg.Header.GetField(tagBeginString, &beginString); err != nil {
		return err
	}

	var targetCompID FIXString
	if err := msg.Header.GetField(tagTargetCompID, &targetCompID); err != nil {
		return err
	}

	var senderCompID FIXString
	if err := msg.Header.GetField(tagSenderCompID, &senderCompID); err != nil {
		return err
	}

	sessionID := SessionID{BeginString: string(beginString), TargetCompID: string(targetCompID), SenderCompID: string(senderCompID)}

	return r.SendToTarget(msg, sessionID)
}

// SendToTarget sends a message based on the sessionID. Convenient for use in FromApp since it provides a session ID for incoming messages.
func (r *Registry) SendToTarget(m Messagable, sessionID SessionID) error {
	msg := m.ToMessage()
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return errUnknownSession
	}

	return session.queueForSend(msg)
}

// ResetSession resets session's sequence numbers.
func (r *Registry) ResetSession(sessionID SessionID) error {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return errUnknownSession
	}
	session.log.OnEvent("Session reset")
	session.State.ShutdownNow(session)
	if err := session.dropAndReset(); err != nil {
		session.logError(err)
		return err
	}

	return nil
}

// UnregisterSession removes a session from the set of known sessions.
func (r *Registry) UnregisterSession(sessionID SessionID) error {
	r.sessionsLock.Lock()
	defer r.sessionsLock.Unlock()

	if _, ok := r.sessions[sessionID]; ok {
		delete(r.sessions, sessionID)
		return nil
	}

	return errUnknownSession
}

// SetNextTargetMsgSeqNum set the next expected target message sequence number for the session matching the session id.
func (r *Registry) SetNextTargetMsgSeqNum(sessionID SessionID, seqNum int) error {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return errUnknownSession
	}
	return session.store.SetNextTargetMsgSeqNum(seqNum)
}

// SetNextSenderMsgSeqNum sets the next outgoing message sequence number for the session matching the session id.
func (r *Registry) SetNextSenderMsgSeqNum(sessionID SessionID, seqNum int) error {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return errUnknownSession
	}
	return session.store.SetNextSenderMsgSeqNum(seqNum)
}

// GetExpectedSenderNum retrieves the expected sender sequence number for the session matching the session id.
func (r *Registry) GetExpectedSenderNum(sessionID SessionID) (int, error) {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return 0, errUnknownSession
	}
	return session.store.NextSenderMsgSeqNum(), nil
}

// GetExpectedTargetNum retrieves the next target sequence number for the session matching the session id.
func (r *Registry) GetExpectedTargetNum(sessionID SessionID) (int, error) {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return 0, errUnknownSession
	}
	return session.store.NextTargetMsgSeqNum(), nil
}

// GetMessageStore returns the MessageStore interface for session matching the session id.
func (r *Registry) GetMessageStore(sessionID SessionID) (MessageStore, error) {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return nil, errUnknownSession
	}
	return session.store, nil
}

// GetLog returns the Log interface for session matching the session id.
func (r *Registry) GetLog(sessionID SessionID) (Log, error) {
	session, ok := r.lookupSession(sessionID)
	if !ok {
		return nil, errUnknownSession
	}
	return session.log, nil
}

func (r *Registry) registerSession(s *session) error {
	r.sessionsLock.Lock()
	defer r.sessionsLock.Unlock()

	if _, ok := r.sessions[s.sessionID]; ok {
		return errDuplicateSessionID
	}

	r.sessions[s.sessionID] = s
	return nil
}

func (r *Registry) lookupSession(sessionID SessionID) (s *session, ok bool) {
	r.sessionsLock.RLock()
	defer r.sessionsLock.RUnlock()

	s, ok = r.sessions[sessionID]
	return
}
