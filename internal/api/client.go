package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// APIResponse describes the wrapper autodns uses for their API response.
type APIResponse[T any] struct {
	Data []T `json:"data"`
}

// Client provides an api client implementation for the AutoDNS API.
type Client struct {
	HTTPClient *http.Client

	// writeMu guards zone mutations. The AutoDNS _stream endpoint is a
	// read-modify-write against the whole zone, so concurrent adds/rems can
	// lose updates. Reads hold it for reading, which keeps them from
	// observing (and caching) a zone midway through a write, the way the
	// single global lock used to.
	writeMu sync.RWMutex

	// cacheMu guards zoneCache and zoneFetch only. It is never held across a
	// network call.
	cacheMu   sync.RWMutex
	zoneCache map[string][]Record

	// zoneFetch holds one lock per zone so that a cold-cache wave of parallel
	// reads collapses into a single fetch instead of one per resource.
	zoneFetch map[string]*sync.Mutex

	HostURL  string
	Context  string
	Username string
	Password string
}

// NewClient returns a new instance of the client.
func NewClient(host, context, username, password string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    "https://" + host,
		Username:   username,
		Password:   password,
		Context:    context,
		zoneCache:  map[string][]Record{},
		zoneFetch:  map[string]*sync.Mutex{},
	}
}

// cachedRecords returns the cached records for a zone, if present.
func (c *Client) cachedRecords(zoneID string) ([]Record, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	records, ok := c.zoneCache[zoneID]

	return records, ok
}

// cacheRecords stores the records for a zone.
func (c *Client) cacheRecords(zoneID string, records []Record) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if c.zoneCache == nil {
		c.zoneCache = map[string][]Record{}
	}

	c.zoneCache[zoneID] = records
}

// fetchLock returns the dedup lock for a zone, creating it on first use.
func (c *Client) fetchLock(zoneID string) *sync.Mutex {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if c.zoneFetch == nil {
		c.zoneFetch = map[string]*sync.Mutex{}
	}

	if _, ok := c.zoneFetch[zoneID]; !ok {
		c.zoneFetch[zoneID] = &sync.Mutex{}
	}

	return c.zoneFetch[zoneID]
}

// invalidateZone drops the cached records for a zone after a mutation.
func (c *Client) invalidateZone(zoneID string) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	delete(c.zoneCache, zoneID)
}

func request[T any](c *Client, req *http.Request) ([]T, error) {
	// Add authentication header
	req.SetBasicAuth(c.Username, c.Password)

	// Add DomainRobot context header
	req.Header.Set("X-Domainrobot-Context", c.Context)

	// Send the request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Read the server response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Only proceed if it's 200 ok
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	// Unmarshel the api response into the proper struct
	resp := &APIResponse[T]{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, err
}
