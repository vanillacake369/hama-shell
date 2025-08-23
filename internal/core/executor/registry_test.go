package executor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionRegistry_registerSession_Success(t *testing.T) {
	// GIVEN: session registry and valid session
	registry := NewSessionRegistry()
	session := &SessionGroup{
		Key:       "test-key",
		SessionID: 12345,
		PGID:      12345,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
	}
	
	// WHEN: registering session
	err := registry.registerSession("test-key", session)
	
	// THEN: should register successfully
	assert.NoError(t, err)
	
	// AND: session should be retrievable
	retrieved, exists := registry.getSession("test-key")
	assert.True(t, exists)
	assert.Equal(t, session, retrieved)
}

func TestSessionRegistry_registerSession_EmptyKey(t *testing.T) {
	// GIVEN: session registry and empty key
	registry := NewSessionRegistry()
	session := &SessionGroup{Key: "test"}
	
	// WHEN: registering with empty key
	err := registry.registerSession("", session)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session key cannot be empty")
}

func TestSessionRegistry_registerSession_NilSession(t *testing.T) {
	// GIVEN: session registry and nil session
	registry := NewSessionRegistry()
	
	// WHEN: registering nil session
	err := registry.registerSession("test-key", nil)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session cannot be nil")
}

func TestSessionRegistry_registerSession_DuplicateKey(t *testing.T) {
	// GIVEN: session registry with existing session
	registry := NewSessionRegistry()
	session1 := &SessionGroup{Key: "test-key"}
	session2 := &SessionGroup{Key: "test-key"}
	
	// WHEN: registering first session then duplicate
	err1 := registry.registerSession("test-key", session1)
	err2 := registry.registerSession("test-key", session2)
	
	// THEN: first should succeed, second should fail
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "session with key 'test-key' already exists")
}

func TestSessionRegistry_getSession_Success(t *testing.T) {
	// GIVEN: session registry with registered session
	registry := NewSessionRegistry()
	session := &SessionGroup{
		Key:       "test-key",
		SessionID: 12345,
		PGID:      12345,
	}
	registry.registerSession("test-key", session)
	
	// WHEN: getting session
	retrieved, exists := registry.getSession("test-key")
	
	// THEN: should retrieve successfully
	assert.True(t, exists)
	assert.Equal(t, session, retrieved)
	assert.Equal(t, "test-key", retrieved.Key)
	assert.Equal(t, 12345, retrieved.SessionID)
}

func TestSessionRegistry_getSession_NotFound(t *testing.T) {
	// GIVEN: empty session registry
	registry := NewSessionRegistry()
	
	// WHEN: getting non-existent session
	retrieved, exists := registry.getSession("non-existent")
	
	// THEN: should not find session
	assert.False(t, exists)
	assert.Nil(t, retrieved)
}

func TestSessionRegistry_getSession_EmptyKey(t *testing.T) {
	// GIVEN: session registry
	registry := NewSessionRegistry()
	
	// WHEN: getting session with empty key
	retrieved, exists := registry.getSession("")
	
	// THEN: should not find session
	assert.False(t, exists)
	assert.Nil(t, retrieved)
}

func TestSessionRegistry_unregisterSession_Success(t *testing.T) {
	// GIVEN: session registry with registered session
	registry := NewSessionRegistry()
	session := &SessionGroup{Key: "test-key", SessionID: 12345}
	registry.registerSession("test-key", session)
	
	// WHEN: unregistering session
	removed, err := registry.unregisterSession("test-key")
	
	// THEN: should unregister successfully
	assert.NoError(t, err)
	assert.Equal(t, session, removed)
	
	// AND: session should no longer be retrievable
	_, exists := registry.getSession("test-key")
	assert.False(t, exists)
}

func TestSessionRegistry_unregisterSession_NotFound(t *testing.T) {
	// GIVEN: empty session registry
	registry := NewSessionRegistry()
	
	// WHEN: unregistering non-existent session
	removed, err := registry.unregisterSession("non-existent")
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Nil(t, removed)
	assert.Contains(t, err.Error(), "session with key 'non-existent' not found")
}

func TestSessionRegistry_unregisterSession_EmptyKey(t *testing.T) {
	// GIVEN: session registry
	registry := NewSessionRegistry()
	
	// WHEN: unregistering with empty key
	removed, err := registry.unregisterSession("")
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Nil(t, removed)
	assert.Contains(t, err.Error(), "session key cannot be empty")
}

func TestSessionRegistry_listSessions(t *testing.T) {
	// GIVEN: session registry with multiple sessions
	registry := NewSessionRegistry()
	session1 := &SessionGroup{Key: "session-1"}
	session2 := &SessionGroup{Key: "session-2"}
	session3 := &SessionGroup{Key: "session-3"}
	
	registry.registerSession("session-1", session1)
	registry.registerSession("session-2", session2)
	registry.registerSession("session-3", session3)
	
	// WHEN: listing sessions
	keys := registry.listSessions()
	
	// THEN: should return all session keys
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "session-1")
	assert.Contains(t, keys, "session-2")
	assert.Contains(t, keys, "session-3")
}

func TestSessionRegistry_listSessions_Empty(t *testing.T) {
	// GIVEN: empty session registry
	registry := NewSessionRegistry()
	
	// WHEN: listing sessions
	keys := registry.listSessions()
	
	// THEN: should return empty list
	assert.Empty(t, keys)
}

func TestSessionRegistry_getAllSessions(t *testing.T) {
	// GIVEN: session registry with multiple sessions
	registry := NewSessionRegistry()
	session1 := &SessionGroup{Key: "session-1", SessionID: 111}
	session2 := &SessionGroup{Key: "session-2", SessionID: 222}
	
	registry.registerSession("session-1", session1)
	registry.registerSession("session-2", session2)
	
	// WHEN: getting all sessions
	sessions := registry.getAllSessions()
	
	// THEN: should return all sessions with correct mapping
	assert.Len(t, sessions, 2)
	assert.Equal(t, session1, sessions["session-1"])
	assert.Equal(t, session2, sessions["session-2"])
	assert.Equal(t, 111, sessions["session-1"].SessionID)
	assert.Equal(t, 222, sessions["session-2"].SessionID)
}

func TestSessionRegistry_sessionCount(t *testing.T) {
	// GIVEN: session registry
	registry := NewSessionRegistry()
	
	// WHEN: checking count initially
	initialCount := registry.sessionCount()
	
	// THEN: should be zero
	assert.Equal(t, 0, initialCount)
	
	// WHEN: adding sessions
	session1 := &SessionGroup{Key: "session-1"}
	session2 := &SessionGroup{Key: "session-2"}
	
	registry.registerSession("session-1", session1)
	registry.registerSession("session-2", session2)
	
	count := registry.sessionCount()
	
	// THEN: count should increase
	assert.Equal(t, 2, count)
	
	// WHEN: removing a session
	registry.unregisterSession("session-1")
	finalCount := registry.sessionCount()
	
	// THEN: count should decrease
	assert.Equal(t, 1, finalCount)
}

func TestSessionRegistry_hasSession(t *testing.T) {
	// GIVEN: session registry with one session
	registry := NewSessionRegistry()
	session := &SessionGroup{Key: "test-key"}
	registry.registerSession("test-key", session)
	
	// WHEN: checking if session exists
	exists1 := registry.hasSession("test-key")
	exists2 := registry.hasSession("non-existent")
	
	// THEN: should correctly identify existence
	assert.True(t, exists1)
	assert.False(t, exists2)
}

func TestSessionRegistry_clearAllSessions(t *testing.T) {
	// GIVEN: session registry with multiple sessions
	registry := NewSessionRegistry()
	session1 := &SessionGroup{Key: "session-1"}
	session2 := &SessionGroup{Key: "session-2"}
	session3 := &SessionGroup{Key: "session-3"}
	
	registry.registerSession("session-1", session1)
	registry.registerSession("session-2", session2)
	registry.registerSession("session-3", session3)
	
	// Verify sessions exist
	assert.Equal(t, 3, registry.sessionCount())
	
	// WHEN: clearing all sessions
	registry.clearAllSessions()
	
	// THEN: all sessions should be removed
	assert.Equal(t, 0, registry.sessionCount())
	assert.Empty(t, registry.listSessions())
	assert.False(t, registry.hasSession("session-1"))
	assert.False(t, registry.hasSession("session-2"))
	assert.False(t, registry.hasSession("session-3"))
}

func TestSessionRegistry_ConcurrentAccess(t *testing.T) {
	// GIVEN: session registry
	registry := NewSessionRegistry()
	
	// WHEN: concurrent operations
	numGoroutines := 10
	done := make(chan bool, numGoroutines*2)
	
	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			session := &SessionGroup{
				Key:       fmt.Sprintf("session-%d", id),
				SessionID: id,
			}
			err := registry.registerSession(fmt.Sprintf("session-%d", id), session)
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	
	// Concurrent retrieval
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Wait a bit for registration
			time.Sleep(10 * time.Millisecond)
			session, exists := registry.getSession(fmt.Sprintf("session-%d", id))
			if exists {
				assert.Equal(t, id, session.SessionID)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}
	
	// THEN: all operations should complete successfully
	assert.Equal(t, numGoroutines, registry.sessionCount())
}