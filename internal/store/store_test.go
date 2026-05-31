package store

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(duration)
}

func TestStoreSetGetExpireAndLazyExpiration(t *testing.T) {
	clock := newFakeClock()
	store := New(clock.Now)

	store.Set("session:1", "alive", time.Time{})
	if value, ok := store.Get("session:1"); !ok || value != "alive" {
		t.Fatalf("Get returned %q/%v", value, ok)
	}

	if !store.Expire("session:1", 2*time.Second) {
		t.Fatalf("Expire returned false")
	}
	ttl, status := store.TTL("session:1")
	if status != TTLHasExpiry || ttl != 2*time.Second {
		t.Fatalf("TTL = %v/%v", ttl, status)
	}

	clock.Advance(3 * time.Second)
	if _, ok := store.Get("session:1"); ok {
		t.Fatalf("expired key was returned")
	}
	stats := store.Stats()
	if stats.ExpiredTotal != 1 {
		t.Fatalf("ExpiredTotal = %d, want 1", stats.ExpiredTotal)
	}
}

func TestStorePersistRemovesExpiration(t *testing.T) {
	clock := newFakeClock()
	store := New(clock.Now)
	store.Set("feature:flag", "on", clock.Now().Add(time.Second))

	if !store.Persist("feature:flag") {
		t.Fatalf("Persist returned false")
	}
	_, status := store.TTL("feature:flag")
	if status != TTLNoExpiry {
		t.Fatalf("TTL status = %v, want TTLNoExpiry", status)
	}
	clock.Advance(2 * time.Second)
	if value, ok := store.Get("feature:flag"); !ok || value != "on" {
		t.Fatalf("persistent key missing after clock advance")
	}
}

func TestStoreKeysAreBoundedSortedAndSkipExpired(t *testing.T) {
	clock := newFakeClock()
	store := New(clock.Now)
	store.Set("user:2", "B", time.Time{})
	store.Set("user:1", "A", time.Time{})
	store.Set("order:1", "O", time.Time{})
	store.Set("user:expired", "X", clock.Now().Add(time.Second))

	clock.Advance(2 * time.Second)
	keys := store.Keys("user:*", 2)
	want := []string{"user:1", "user:2"}
	if fmt.Sprint(keys) != fmt.Sprint(want) {
		t.Fatalf("keys = %#v, want %#v", keys, want)
	}
}

func TestStoreCleanupExpired(t *testing.T) {
	clock := newFakeClock()
	store := New(clock.Now)
	for i := 0; i < 10; i++ {
		store.Set(fmt.Sprintf("k:%d", i), "v", clock.Now().Add(time.Second))
	}
	clock.Advance(2 * time.Second)
	if cleaned := store.CleanupExpired(4); cleaned != 4 {
		t.Fatalf("first cleanup removed %d, want 4", cleaned)
	}
	if cleaned := store.CleanupExpired(0); cleaned != 6 {
		t.Fatalf("second cleanup removed %d, want 6", cleaned)
	}
}

func TestSnapshotRestoreSkipsExpiredEntries(t *testing.T) {
	clock := newFakeClock()
	store := New(clock.Now)
	store.Set("live", "1", time.Time{})
	store.Set("ttl", "2", clock.Now().Add(time.Second))
	snapshot := store.Snapshot()

	restored := New(clock.Now)
	clock.Advance(2 * time.Second)
	restored.Restore(snapshot)

	if value, ok := restored.Get("live"); !ok || value != "1" {
		t.Fatalf("live key missing after restore")
	}
	if _, ok := restored.Get("ttl"); ok {
		t.Fatalf("expired snapshot key restored as live")
	}
}

func TestConcurrentExpireGetAndCleanupDoNotRace(t *testing.T) {
	clock := newFakeClock()
	store := New(clock.Now)
	for i := 0; i < 1000; i++ {
		store.Set(fmt.Sprintf("key:%d", i), "value", clock.Now().Add(time.Second))
	}

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				_, _ = store.Get(fmt.Sprintf("key:%d", (id+j)%1000))
			}
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = store.Expire(fmt.Sprintf("key:%d", i), time.Millisecond)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			store.Del(fmt.Sprintf("key:%d", i))
			_ = store.CleanupExpired(25)
		}
	}()
	wg.Wait()
}
