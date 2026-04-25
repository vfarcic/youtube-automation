package thumbnail

import (
	"sync"
	"testing"
	"time"
)

func TestGeneratedImageStore_AddAndGet(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	img := GeneratedImage{
		Provider: "gemini",
		Style:    "with illustration",
		Data:     []byte("image-bytes"),
	}

	id, err := store.Add(img)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if id == "" {
		t.Fatal("Add() returned empty ID")
	}

	got, ok := store.Get(id)
	if !ok {
		t.Fatal("Get() returned false for existing image")
	}
	if got.ID != id {
		t.Errorf("ID = %q, want %q", got.ID, id)
	}
	if got.Provider != "gemini" {
		t.Errorf("Provider = %q, want %q", got.Provider, "gemini")
	}
	if got.Style != "with illustration" {
		t.Errorf("Style = %q, want %q", got.Style, "with illustration")
	}
	if string(got.Data) != "image-bytes" {
		t.Errorf("Data = %q, want %q", string(got.Data), "image-bytes")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestGeneratedImageStore_GetNonExistent(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	_, ok := store.Get("nonexistent-id")
	if ok {
		t.Error("Get() returned true for non-existent ID")
	}
}

func TestGeneratedImageStore_Remove(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	img := GeneratedImage{Provider: "test", Style: "test", Data: []byte("data")}
	id, _ := store.Add(img)

	tests := []struct {
		name    string
		id      string
		wantOK  bool
		wantLen int
	}{
		{
			name:    "remove existing",
			id:      id,
			wantOK:  true,
			wantLen: 0,
		},
		{
			name:    "remove already removed",
			id:      id,
			wantOK:  false,
			wantLen: 0,
		},
		{
			name:    "remove non-existent",
			id:      "no-such-id",
			wantOK:  false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := store.Remove(tt.id)
			if ok != tt.wantOK {
				t.Errorf("Remove() = %v, want %v", ok, tt.wantOK)
			}
			if store.Len() != tt.wantLen {
				t.Errorf("Len() = %d, want %d", store.Len(), tt.wantLen)
			}
		})
	}
}

func TestGeneratedImageStore_TTLExpiry(t *testing.T) {
	now := time.Now()
	store := NewGeneratedImageStore(5 * time.Minute)
	store.nowFunc = func() time.Time { return now }

	img := GeneratedImage{Provider: "test", Style: "test", Data: []byte("data")}
	id, _ := store.Add(img)

	// Image should be retrievable before TTL.
	if _, ok := store.Get(id); !ok {
		t.Fatal("image should be retrievable before TTL")
	}

	// Advance time past TTL.
	store.nowFunc = func() time.Time { return now.Add(6 * time.Minute) }

	// Image should be expired now.
	if _, ok := store.Get(id); ok {
		t.Error("Get() should return false for expired image")
	}

	// Image is still in the store until cleanup.
	if store.Len() != 1 {
		t.Errorf("Len() = %d, want 1 (expired but not cleaned up)", store.Len())
	}
}

func TestGeneratedImageStore_Cleanup(t *testing.T) {
	now := time.Now()
	store := NewGeneratedImageStore(5 * time.Minute)
	store.nowFunc = func() time.Time { return now }

	// Add 3 images.
	for range 3 {
		store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("d")})
	}
	if store.Len() != 3 {
		t.Fatalf("expected 3 images, got %d", store.Len())
	}

	// Advance past TTL and add 1 more.
	store.nowFunc = func() time.Time { return now.Add(6 * time.Minute) }
	store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("new")})

	// Cleanup should remove the 3 expired images.
	removed := store.Cleanup()
	if removed != 3 {
		t.Errorf("Cleanup() removed %d, want 3", removed)
	}
	if store.Len() != 1 {
		t.Errorf("Len() after cleanup = %d, want 1", store.Len())
	}
}

func TestGeneratedImageStore_CleanupNoExpired(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)
	store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("d")})

	removed := store.Cleanup()
	if removed != 0 {
		t.Errorf("Cleanup() removed %d, want 0", removed)
	}
}

func TestGeneratedImageStore_UniqueIDs(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)
	ids := make(map[string]bool)

	for range 50 {
		id, err := store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("d")})
		if err != nil {
			t.Fatalf("Add() error: %v", err)
		}
		if ids[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestGeneratedImageStore_ThreadSafety(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent writes.
	ids := make(chan string, goroutines)
	for range goroutines {
		wg.Go(func() {
			id, err := store.Add(GeneratedImage{
				Provider: "concurrent",
				Style:    "test",
				Data:     []byte("data"),
			})
			if err != nil {
				t.Errorf("concurrent Add() error: %v", err)
				return
			}
			ids <- id
		})
	}
	wg.Wait()
	close(ids)

	if store.Len() != goroutines {
		t.Errorf("Len() = %d, want %d after concurrent adds", store.Len(), goroutines)
	}

	// Collect IDs for concurrent reads and removes.
	var allIDs []string
	for id := range ids {
		allIDs = append(allIDs, id)
	}

	// Concurrent reads.
	for _, id := range allIDs {
		wg.Go(func() {
			store.Get(id)
		})
	}
	wg.Wait()

	// Concurrent removes.
	for _, id := range allIDs {
		wg.Go(func() {
			store.Remove(id)
		})
	}
	wg.Wait()

	if store.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after concurrent removes", store.Len())
	}
}

func TestGeneratedImageStore_ConcurrentReadWrite(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	var wg sync.WaitGroup

	// Mix of adds, gets, removes, cleanups running concurrently.
	for range 20 {
		wg.Go(func() {
			store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("d")})
		})
		wg.Go(func() {
			store.Get("some-id")
		})
		wg.Go(func() {
			store.Remove("some-id")
		})
		wg.Go(func() {
			store.Cleanup()
		})
	}

	wg.Wait()
	// No panics or data races = success (run with -race flag).
}

func TestGeneratedImageStore_Len(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	if store.Len() != 0 {
		t.Errorf("empty store Len() = %d, want 0", store.Len())
	}

	store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("d")})
	store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("d")})

	if store.Len() != 2 {
		t.Errorf("Len() = %d, want 2", store.Len())
	}
}

func TestGeneratedImageStore_ByteSliceIsolation(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)

	original := []byte("original-data")
	id, err := store.Add(GeneratedImage{Provider: "p", Style: "s", Data: original})
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Mutate the original slice after Add — stored data must not change.
	original[0] = 'X'

	got, ok := store.Get(id)
	if !ok {
		t.Fatal("Get() returned false for existing image")
	}
	if got.Data[0] == 'X' {
		t.Error("stored data was mutated through original slice (Add did not copy)")
	}
	if string(got.Data) != "original-data" {
		t.Errorf("stored data = %q, want %q", string(got.Data), "original-data")
	}

	// Mutate the slice returned by Get — stored data must not change.
	got.Data[0] = 'Z'

	got2, _ := store.Get(id)
	if got2.Data[0] == 'Z' {
		t.Error("stored data was mutated through Get return value (Get did not copy)")
	}
}

func TestGeneratedImageStore_MaxItemsCap(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)
	store.maxItems = 3 // low cap for testing

	for i := range 3 {
		_, err := store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte{byte(i)}})
		if err != nil {
			t.Fatalf("Add() #%d error: %v", i, err)
		}
	}

	if store.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", store.Len())
	}

	// Fourth add should fail with ErrStoreFull.
	_, err := store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("overflow")})
	if err == nil {
		t.Fatal("expected ErrStoreFull, got nil")
	}
	if err != ErrStoreFull {
		t.Errorf("error = %v, want ErrStoreFull", err)
	}

	// Store size should remain at cap.
	if store.Len() != 3 {
		t.Errorf("Len() = %d, want 3 after rejected add", store.Len())
	}
}

func TestGeneratedImageStore_CapAfterRemove(t *testing.T) {
	store := NewGeneratedImageStore(5 * time.Minute)
	store.maxItems = 2

	id1, _ := store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("1")})
	store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("2")})

	// Store is full.
	_, err := store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("3")})
	if err != ErrStoreFull {
		t.Fatalf("expected ErrStoreFull, got %v", err)
	}

	// Remove one, then add should succeed.
	store.Remove(id1)
	_, err = store.Add(GeneratedImage{Provider: "p", Style: "s", Data: []byte("3")})
	if err != nil {
		t.Fatalf("Add() after Remove() error: %v", err)
	}
}
