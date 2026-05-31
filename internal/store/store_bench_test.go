package store

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkStoreSet1MKeys(b *testing.B) {
	cache := New(nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("bench:set:%d", i), "value", time.Time{})
	}
}

func BenchmarkStoreGetMillionKeyDataset(b *testing.B) {
	cache := New(nil)
	const dataset = 1_000_000
	for i := 0; i < dataset; i++ {
		cache.Set(fmt.Sprintf("bench:get:%d", i), "value", time.Time{})
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(fmt.Sprintf("bench:get:%d", i%dataset))
	}
}

func BenchmarkStoreMixed80Read20Write(b *testing.B) {
	cache := New(nil)
	const dataset = 100_000
	for i := 0; i < dataset; i++ {
		cache.Set(fmt.Sprintf("bench:mixed:%d", i), "value", time.Time{})
	}

	var counter atomic.Uint64
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			key := fmt.Sprintf("bench:mixed:%d", n%dataset)
			if n%5 == 0 {
				cache.Set(key, "new-value", time.Time{})
				continue
			}
			_, _ = cache.Get(key)
		}
	})
}

func BenchmarkStore100ConcurrentClients(b *testing.B) {
	cache := New(nil)
	const clients = 100
	var counter atomic.Uint64
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(clients)
		for c := 0; c < clients; c++ {
			go func() {
				defer wg.Done()
				n := counter.Add(1)
				key := fmt.Sprintf("bench:client:%d", n)
				cache.Set(key, "value", time.Time{})
				_, _ = cache.Get(key)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkTTLCleanup(b *testing.B) {
	clock := newFakeClock()
	cache := New(clock.Now)
	const dataset = 100_000
	for i := 0; i < dataset; i++ {
		cache.Set(fmt.Sprintf("bench:ttl:%d", i), "value", clock.Now().Add(time.Second))
	}
	clock.Advance(2 * time.Second)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if cleaned := cache.CleanupExpired(1000); cleaned == 0 {
			b.StopTimer()
			for j := 0; j < dataset; j++ {
				cache.Set(fmt.Sprintf("bench:ttl:%d:%d", i, j), "value", clock.Now().Add(-time.Second))
			}
			b.StartTimer()
		}
	}
}
