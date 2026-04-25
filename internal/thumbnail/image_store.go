package thumbnail

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// DefaultMaxImages is the default maximum number of images the store will hold.
const DefaultMaxImages = 100

// ErrStoreFull is returned when the store has reached its maximum capacity.
var ErrStoreFull = errors.New("image store is full")

// StoredImage holds a generated image along with its metadata in the store.
type StoredImage struct {
	ID        string
	Provider  string
	Style     string
	Data      []byte
	CreatedAt time.Time
}

// GeneratedImageStore provides thread-safe, TTL-based temporary storage
// for generated thumbnail images before user selection.
type GeneratedImageStore struct {
	mu       sync.RWMutex
	images   map[string]StoredImage
	ttl      time.Duration
	maxItems int
	nowFunc  func() time.Time // for testing
}

// NewGeneratedImageStore creates a new store with the given TTL.
// Images older than ttl are considered expired and cleaned up.
// The store holds at most DefaultMaxImages items.
func NewGeneratedImageStore(ttl time.Duration) *GeneratedImageStore {
	return &GeneratedImageStore{
		images:   make(map[string]StoredImage),
		ttl:      ttl,
		maxItems: DefaultMaxImages,
		nowFunc:  time.Now,
	}
}

// Add stores a generated image and returns its unique ID.
// Returns ErrStoreFull if the store has reached its maximum capacity.
func (s *GeneratedImageStore) Add(img GeneratedImage) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}

	// Copy the byte slice to prevent caller mutation.
	dataCopy := make([]byte, len(img.Data))
	copy(dataCopy, img.Data)

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.images) >= s.maxItems {
		return "", ErrStoreFull
	}

	s.images[id] = StoredImage{
		ID:        id,
		Provider:  img.Provider,
		Style:     img.Style,
		Data:      dataCopy,
		CreatedAt: s.nowFunc(),
	}

	return id, nil
}

// Get retrieves an image by ID. Returns the image and true if found and not
// expired, or a zero value and false otherwise.
// The returned Data slice is a copy and safe to modify.
func (s *GeneratedImageStore) Get(id string) (StoredImage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	img, ok := s.images[id]
	if !ok {
		return StoredImage{}, false
	}

	if s.nowFunc().Sub(img.CreatedAt) > s.ttl {
		return StoredImage{}, false
	}

	// Return a copy of Data to prevent callers from mutating stored bytes.
	dataCopy := make([]byte, len(img.Data))
	copy(dataCopy, img.Data)
	img.Data = dataCopy

	return img, true
}

// Claim atomically retrieves and removes an image from the store.
// It returns the image and true if found and not expired, or a zero value and false otherwise.
// This prevents TOCTOU race conditions where two concurrent requests could both
// claim the same image via separate Get+Remove calls.
func (s *GeneratedImageStore) Claim(id string) (StoredImage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	img, ok := s.images[id]
	if !ok {
		return StoredImage{}, false
	}

	if s.nowFunc().Sub(img.CreatedAt) > s.ttl {
		return StoredImage{}, false
	}

	delete(s.images, id)

	// Return a copy of Data to prevent callers from mutating stored bytes.
	dataCopy := make([]byte, len(img.Data))
	copy(dataCopy, img.Data)
	img.Data = dataCopy

	return img, true
}

// Remove deletes an image from the store. Returns true if it existed.
func (s *GeneratedImageStore) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.images[id]
	if ok {
		delete(s.images, id)
	}
	return ok
}

// Cleanup removes all expired images from the store.
func (s *GeneratedImageStore) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFunc()
	removed := 0
	for id, img := range s.images {
		if now.Sub(img.CreatedAt) > s.ttl {
			delete(s.images, id)
			removed++
		}
	}
	return removed
}

// Len returns the number of images currently stored (including expired ones
// not yet cleaned up).
func (s *GeneratedImageStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.images)
}

// generateID creates a random hex ID (16 bytes = 32 hex chars).
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
