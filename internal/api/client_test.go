package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testZoneID = "example.test@a.ns14.net"

// newTestClient returns a client pointed at a local test server. No request
// ever leaves the process, so these tests never touch the real AutoDNS API.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *int64) {
	t.Helper()

	var calls int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(strings.TrimPrefix(srv.URL, "http://"), "4", "user", "pass")
	c.HostURL = srv.URL

	return c, &calls
}

const oneRecord = `{"name":"www","ttl":60,"type":"A","value":"1.1.1.1"}`

func zoneBody() string {
	return fmt.Sprintf(`{"data":[{"origin":"example.test","virtualNameServer":"a.ns14.net","resourceRecords":[%s]}]}`, oneRecord)
}

func TestGetRecordsCachesZone(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, zoneBody())
	})

	for i := range 5 {
		records, err := c.GetRecords(context.Background(), testZoneID)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %s", i, err)
		}

		if len(records) != 1 || records[0].Name != "www" {
			t.Fatalf("call %d: unexpected records: %+v", i, records)
		}
	}

	// The whole point of the change: 5 reads, 1 HTTP request.
	if got := atomic.LoadInt64(calls); got != 1 {
		t.Errorf("expected 1 HTTP call for 5 reads, got %d", got)
	}
}

func TestGetRecordsSeparateZonesAreCachedSeparately(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, zoneBody())
	})

	for _, zone := range []string{testZoneID, "other.test@a.ns14.net", testZoneID} {
		if _, err := c.GetRecords(context.Background(), zone); err != nil {
			t.Fatalf("zone %s: unexpected error: %s", zone, err)
		}
	}

	if got := atomic.LoadInt64(calls); got != 2 {
		t.Errorf("expected 2 HTTP calls for 2 distinct zones, got %d", got)
	}
}

func TestMutationsInvalidateCache(t *testing.T) {
	tests := map[string]func(*Client) error{
		"create": func(c *Client) error {
			return c.CreateRecords(context.Background(), testZoneID, []Record{{Name: "a", Type: "A", Value: "1.1.1.1"}})
		},
		"update": func(c *Client) error {
			return c.UpdateRecords(context.Background(), testZoneID,
				[]Record{{Name: "a", Type: "A", Value: "1.1.1.1"}},
				[]Record{{Name: "a", Type: "A", Value: "2.2.2.2"}})
		},
		"delete": func(c *Client) error {
			return c.DeleteRecords(context.Background(), testZoneID, []Record{{Name: "a", Type: "A", Value: "1.1.1.1"}})
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, zoneBody())
			})

			// Warm the cache, mutate, then read again.
			if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
				t.Fatalf("warm read: %s", err)
			}

			if err := mutate(c); err != nil {
				t.Fatalf("mutate: %s", err)
			}

			if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
				t.Fatalf("read after mutate: %s", err)
			}

			// read (1) + mutate (1) + re-read because cache was dropped (1)
			if got := atomic.LoadInt64(calls); got != 3 {
				t.Errorf("expected 3 HTTP calls, got %d (cache not invalidated?)", got)
			}
		})
	}
}

func TestGetRecordsEmptyDataReturnsErrorNotPanic(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	})

	// Before the fix this panicked on res[0] and crashed the provider.
	_, err := c.GetRecords(context.Background(), testZoneID)
	if err == nil {
		t.Fatal("expected an error for an empty data array, got nil")
	}
}

func TestGetRecordsErrorIsNotCached(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)

	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"data":[]}`)

			return
		}
		fmt.Fprint(w, zoneBody())
	})

	if _, err := c.GetRecords(context.Background(), testZoneID); err == nil {
		t.Fatal("expected first call to fail")
	}

	fail.Store(false)

	if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
		t.Fatalf("expected recovery after failure, got: %s", err)
	}

	if got := atomic.LoadInt64(calls); got != 2 {
		t.Errorf("expected 2 HTTP calls (failure must not be cached), got %d", got)
	}
}

func TestConcurrentColdReadsFetchOnce(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, zoneBody())
	})

	// 20 resources hit an empty cache at once, as Terraform does at the start
	// of a refresh. They must collapse into a single fetch.
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			records, err := c.GetRecords(context.Background(), testZoneID)
			if err != nil {
				t.Errorf("concurrent read: %s", err)

				return
			}

			if len(records) != 1 {
				t.Errorf("expected 1 record, got %d", len(records))
			}
		}()
	}

	wg.Wait()

	if got := atomic.LoadInt64(calls); got != 1 {
		t.Errorf("expected 1 HTTP call for 20 concurrent cold reads, got %d", got)
	}
}

// TestCallersCannotCorruptTheCache guards a real regression: callers filter the
// returned slice in place with slices.DeleteFunc, so handing out the cached
// slice lets the first resource destroy the records every later one needs.
func TestCallersCannotCorruptTheCache(t *testing.T) {
	body := `{"data":[{"origin":"example.test","virtualNameServer":"a.ns14.net","resourceRecords":[` +
		`{"name":"www","ttl":60,"type":"A","value":"1.1.1.1"},` +
		`{"name":"mail","ttl":60,"type":"A","value":"2.2.2.2"},` +
		`{"name":"api","ttl":60,"type":"A","value":"3.3.3.3"}]}]}`

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})

	// Warm the cache, then take a second read so the slice under test comes
	// from the cache-hit path, and filter it the way the resource does.
	if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
		t.Fatalf("warm read: %s", err)
	}

	cached, err := c.GetRecords(context.Background(), testZoneID)
	if err != nil {
		t.Fatalf("first read: %s", err)
	}

	_ = slices.DeleteFunc(cached, func(r Record) bool { return r.Name != "www" })

	// Every later resource must still see the untouched zone.
	second, err := c.GetRecords(context.Background(), testZoneID)
	if err != nil {
		t.Fatalf("second read: %s", err)
	}

	if len(second) != 3 {
		t.Fatalf("cache corrupted: expected 3 records, got %d: %+v", len(second), second)
	}

	for _, want := range []string{"www", "mail", "api"} {
		if !slices.ContainsFunc(second, func(r Record) bool { return r.Name == want }) {
			t.Errorf("cache corrupted: record %q missing", want)
		}
	}
}

// A mutating call must drop the cache BEFORE it writes, not only after. The
// post-write invalidate already covers the success path, so this isolates the
// case a successful write hides: a write that FAILS must still have dropped any
// snapshot cached earlier in the same run, so a later read cannot serve stale
// data from before the attempted mutation.
func TestMutationInvalidatesBeforeWrite(t *testing.T) {
	var failWrite atomic.Bool

	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// _stream is the mutating endpoint; fail it while failWrite is set.
		if failWrite.Load() && strings.Contains(r.URL.Path, "_stream") {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"data":[]}`)

			return
		}
		fmt.Fprint(w, zoneBody())
	})

	// Warm the cache, as a create existence-check or refresh would.
	if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
		t.Fatalf("warm read: %s", err)
	}

	// A mutation is attempted and fails at the network. Only a pre-write
	// invalidate can have dropped the snapshot in this path.
	failWrite.Store(true)
	if err := c.CreateRecords(context.Background(), testZoneID, []Record{{Name: "x", Type: "A", Value: "1.1.1.1"}}); err == nil {
		t.Fatal("expected the create to fail")
	}
	failWrite.Store(false)

	// The next read must re-fetch, not serve the pre-write snapshot.
	before := atomic.LoadInt64(calls)
	if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
		t.Fatalf("read after failed mutate: %s", err)
	}

	if atomic.LoadInt64(calls) == before {
		t.Error("read after a failed write was served from a pre-write snapshot; cache was not invalidated before the write")
	}
}

// The released provider held one global lock across every request, so a read
// could never overlap a write. Splitting the locks for caching must not lose
// that: a fetch overlapping an in-flight _stream write can observe, and then
// cache, a zone midway through a mutation.
func TestFetchDoesNotOverlapInFlightWrite(t *testing.T) {
	var (
		writeInFlight atomic.Bool
		overlapped    atomic.Bool
		writeStarted  = make(chan struct{})
	)

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "_stream") {
			writeInFlight.Store(true)
			close(writeStarted)
			// Hold the write open long enough for a racing read to land.
			time.Sleep(60 * time.Millisecond)
			writeInFlight.Store(false)
			fmt.Fprint(w, `{"data":[]}`)

			return
		}

		// A GET must never be served while a write is in flight.
		if writeInFlight.Load() {
			overlapped.Store(true)
		}
		fmt.Fprint(w, zoneBody())
	})

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := c.CreateRecords(context.Background(), testZoneID,
			[]Record{{Name: "a", Type: "A", Value: "1.1.1.1"}}); err != nil {
			t.Errorf("create: %s", err)
		}
	}()

	// Start the read once the write is actually in flight.
	<-writeStarted

	wg.Add(1)

	go func() {
		defer wg.Done()

		if _, err := c.GetRecords(context.Background(), testZoneID); err != nil {
			t.Errorf("read: %s", err)
		}
	}()

	wg.Wait()

	if overlapped.Load() {
		t.Error("a zone fetch was served while a _stream write was in flight")
	}
}
