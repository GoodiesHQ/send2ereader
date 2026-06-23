package store

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// File represents an uploaded file
type File struct {
	Name     string
	Data     []byte
	Kepubify bool
}

// Session represnts a file transfer session
type Session struct {
	Code      string
	CreatedAt time.Time
	mu        sync.Mutex
	file      *File
}

// SetFile safely sets the file for the session
func (s *Session) SetFile(f *File) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return fmt.Errorf("file already set for session %s", s.Code)
	}

	s.file = f
	return nil
}

// GetFile safely retrieves the file for the session
func (s *Session) GetFile() (File, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return File{}, false
	}

	return *s.file, true
}

// Store manages all active sessions
type Store struct {
	mu        sync.RWMutex
	TTL       time.Duration
	sessions  map[string]*Session
	generator *codeGenerator
}

// New creates a new Store instance
func New(ttl time.Duration) *Store {
	return &Store{
		TTL:       ttl,
		sessions:  make(map[string]*Session),
		generator: nil,
	}
}

// initGenerator initializes the code generator for the given code size
func (s *Store) initGenerator(size int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the generator is already initialized with the correct code size
	if s.generator != nil && s.generator.size == size {
		return nil
	}

	// Initialize a new code generator for the specified code size
	generator, err := newCodeGenerator(size)
	if err != nil {
		return err
	}

	s.generator = generator
	return nil
}

// CreateCode generates a new session with a unique permuted code
func (s *Store) CreateCode(size int) (*Session, error) {
	if err := s.initGenerator(size); err != nil {
		return nil, err
	}

	// Generation should be deterministic and sequential
	// The tries just ensures we don't get stuck in an infinite loop if the code space is exhausted.
	const tries = 256
	for i := 0; i < tries; i++ {
		code := s.generator.Next()

		s.mu.Lock()
		if _, exists := s.sessions[code]; !exists {
			sess := &Session{
				Code:      code,
				CreatedAt: time.Now(),
			}
			s.sessions[code] = sess
			s.mu.Unlock()
			return sess, nil
		}
		s.mu.Unlock()
	}

	return nil, fmt.Errorf("failed to generate unique code after %d attempts", tries)
}

// Get retrieves a session by code, returning false if it doesn't exist or is expired
func (s *Store) Get(code string) (*Session, bool) {
	// Normalize the code to uppercase and trim whitespace
	code = strings.ToUpper(strings.TrimSpace(code))
	s.mu.RLock()
	sess, ok := s.sessions[code]
	s.mu.RUnlock()

	// If the session doesn't exist, return false.
	if !ok {
		return nil, false
	}

	// If the session exists but is expired, delete it, then return false.
	if time.Since(sess.CreatedAt) > s.TTL {
		s.Delete(code)
		return nil, false
	}

	// Return the session and true if it exists and is not expired.
	return sess, true
}

// Delete removes a session by code
func (s *Store) Delete(code string) {
	s.mu.Lock()
	delete(s.sessions, code)
	s.mu.Unlock()
}

// Cleanup removes all expired sessions from the store
func (s *Store) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, sess := range s.sessions {
		if time.Since(sess.CreatedAt) > s.TTL {
			delete(s.sessions, code)
		}
	}
}
