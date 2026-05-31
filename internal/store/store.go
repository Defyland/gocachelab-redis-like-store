package store

import (
	"path"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Clock func() time.Time

type Entry struct {
	Value     string
	ExpiresAt time.Time
}

type Stats struct {
	LiveKeys     int
	PhysicalKeys int
	ExpiredTotal uint64
}

type Store struct {
	mu           sync.RWMutex
	items        map[string]Entry
	clock        Clock
	expiredTotal atomic.Uint64
}

func New(clock Clock) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{
		items: make(map[string]Entry),
		clock: clock,
	}
}

// Set stores a value and optional absolute expiration. A zero expiration means
// the key is persistent until DEL or a later EXPIRE command.
func (s *Store) Set(key string, value string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = Entry{Value: value, ExpiresAt: expiresAt}
}

func (s *Store) Get(key string) (string, bool) {
	now := s.clock()

	s.mu.RLock()
	entry, ok := s.items[key]
	if !ok {
		s.mu.RUnlock()
		return "", false
	}
	if !entry.expired(now) {
		s.mu.RUnlock()
		return entry.Value, true
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok = s.items[key]
	if ok && entry.expired(now) {
		delete(s.items, key)
		s.expiredTotal.Add(1)
		return "", false
	}
	if !ok {
		return "", false
	}
	return entry.Value, true
}

func (s *Store) Del(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		if _, ok := s.items[key]; ok {
			delete(s.items, key)
			deleted++
		}
	}
	return deleted
}

func (s *Store) Exists(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	count := 0
	for _, key := range keys {
		entry, ok := s.items[key]
		if !ok {
			continue
		}
		if entry.expired(now) {
			delete(s.items, key)
			s.expiredTotal.Add(1)
			continue
		}
		count++
	}
	return count
}

func (s *Store) Expire(key string, ttl time.Duration) bool {
	if ttl <= 0 {
		return s.Del(key) == 1
	}
	return s.ExpireAt(key, s.clock().Add(ttl))
}

func (s *Store) ExpireAt(key string, expiresAt time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	entry, ok := s.items[key]
	if !ok {
		return false
	}
	if entry.expired(now) {
		delete(s.items, key)
		s.expiredTotal.Add(1)
		return false
	}
	entry.ExpiresAt = expiresAt
	s.items[key] = entry
	return true
}

func (s *Store) Persist(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	entry, ok := s.items[key]
	if !ok {
		return false
	}
	if entry.expired(now) {
		delete(s.items, key)
		s.expiredTotal.Add(1)
		return false
	}
	if entry.ExpiresAt.IsZero() {
		return false
	}
	entry.ExpiresAt = time.Time{}
	s.items[key] = entry
	return true
}

type TTLStatus int

const (
	TTLMissing TTLStatus = iota
	TTLNoExpiry
	TTLHasExpiry
)

func (s *Store) TTL(key string) (time.Duration, TTLStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	entry, ok := s.items[key]
	if !ok {
		return 0, TTLMissing
	}
	if entry.expired(now) {
		delete(s.items, key)
		s.expiredTotal.Add(1)
		return 0, TTLMissing
	}
	if entry.ExpiresAt.IsZero() {
		return 0, TTLNoExpiry
	}
	return entry.ExpiresAt.Sub(now), TTLHasExpiry
}

func (s *Store) Keys(pattern string, limit int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	keys := make([]string, 0, min(len(s.items), positiveLimit(limit, len(s.items))))
	for key, entry := range s.items {
		if entry.expired(now) {
			delete(s.items, key)
			s.expiredTotal.Add(1)
			continue
		}
		matched, err := path.Match(pattern, key)
		if err != nil {
			matched = key == pattern
		}
		if matched {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if limit > 0 && len(keys) > limit {
		return keys[:limit]
	}
	return keys
}

func (s *Store) CleanupExpired(limit int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	cleaned := 0
	for key, entry := range s.items {
		if !entry.expired(now) {
			continue
		}
		delete(s.items, key)
		cleaned++
		if limit > 0 && cleaned >= limit {
			break
		}
	}
	if cleaned > 0 {
		s.expiredTotal.Add(uint64(cleaned))
	}
	return cleaned
}

func (s *Store) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	entries := make(map[string]SnapshotEntry, len(s.items))
	for key, entry := range s.items {
		if entry.expired(now) {
			delete(s.items, key)
			s.expiredTotal.Add(1)
			continue
		}
		entries[key] = SnapshotEntry{
			Value:             entry.Value,
			ExpiresAtUnixNano: unixNano(entry.ExpiresAt),
		}
	}
	return Snapshot{
		Version:   SnapshotVersion,
		CreatedAt: now.UTC(),
		Entries:   entries,
	}
}

func (s *Store) Restore(snapshot Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	s.items = make(map[string]Entry, len(snapshot.Entries))
	for key, entry := range snapshot.Entries {
		expiresAt := timeFromUnixNano(entry.ExpiresAtUnixNano)
		if !expiresAt.IsZero() && !expiresAt.After(now) {
			continue
		}
		s.items[key] = Entry{Value: entry.Value, ExpiresAt: expiresAt}
	}
}

func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := s.clock()
	live := 0
	for _, entry := range s.items {
		if !entry.expired(now) {
			live++
		}
	}
	return Stats{
		LiveKeys:     live,
		PhysicalKeys: len(s.items),
		ExpiredTotal: s.expiredTotal.Load(),
	}
}

func (e Entry) expired(now time.Time) bool {
	return !e.ExpiresAt.IsZero() && !e.ExpiresAt.After(now)
}

func unixNano(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}

func timeFromUnixNano(value int64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	return time.Unix(0, value).UTC()
}

func positiveLimit(limit int, fallback int) int {
	if limit > 0 {
		return limit
	}
	return fallback
}
