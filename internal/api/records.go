package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// Record defines the JSON structure of a DNS record in the AutoDNS API.
type Record struct {
	ID    string
	Name  string `json:"name"`
	TTL   int64  `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Pref  int32  `json:"pref,omitempty"`
}

// ZoneStream as the payload in create/update/delete operations.
type ZoneStream struct {
	Adds []Record `json:"adds"`
	Rems []Record `json:"rems"`
}

// CreateRecords sends an API request to create the records in the JSON payload.
func (c *Client) CreateRecords(ctx context.Context, zoneID string, records []Record) error {
	zoneInfo := strings.Split(zoneID, "@")
	if len(zoneInfo) != 2 {
		return fmt.Errorf("the zone is must have the format origin@virtualNameServer")
	}

	// Serialize mutations: _stream is a read-modify-write on the whole zone.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// Drop any cached copy before the write too: a create existence-check or
	// refresh earlier in this run may have cached a snapshot that this write
	// is about to make stale.
	c.invalidateZone(zoneID)

	zs, err := json.Marshal(&ZoneStream{
		Adds: records,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/zone/%s/_stream", c.HostURL, zoneInfo[0]), strings.NewReader(string(zs)))
	if err != nil {
		return err
	}

	_, err = request[any](c, req)
	if err != nil {
		return err
	}

	// The zone changed, so any cached copy is now stale.
	c.invalidateZone(zoneID)

	return nil
}

// GetRecords Fetches all the records in the zone.
func (c *Client) GetRecords(ctx context.Context, zoneID string) ([]Record, error) {
	zoneInfo := strings.Split(zoneID, "@")
	if len(zoneInfo) != 2 {
		return nil, fmt.Errorf("the zone is must have the format origin@virtualNameServer")
	}

	// Serve from the per-zone cache when we already fetched this zone. The
	// API has no per-record read, so without this every record resource
	// re-downloads the whole zone. Callers filter the result in place with
	// slices.DeleteFunc, so hand out a copy and never the cached slice.
	if records, ok := c.cachedRecords(zoneID); ok {
		return slices.Clone(records), nil
	}

	// Only one fetch per zone: parallel readers that all miss the cache queue
	// here, and every one after the winner finds the zone already cached.
	fetch := c.fetchLock(zoneID)
	fetch.Lock()
	defer fetch.Unlock()

	if records, ok := c.cachedRecords(zoneID); ok {
		return slices.Clone(records), nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/zone/%s/%s", c.HostURL, zoneInfo[0], zoneInfo[1]), nil)
	if err != nil {
		return nil, err
	}

	res, err := request[Zone](c, req)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("no zone returned by the API for %s", zoneID)
	}

	c.cacheRecords(zoneID, res[0].Records)

	return slices.Clone(res[0].Records), nil
}

// UpdateRecords sends an API request to update the records in the JSON payload.
func (c *Client) UpdateRecords(ctx context.Context, zoneID string, oldRecords, newRecords []Record) error {
	zoneInfo := strings.Split(zoneID, "@")
	if len(zoneInfo) != 2 {
		return fmt.Errorf("the zone is must have the format origin@virtualNameServer")
	}

	// Serialize mutations: _stream is a read-modify-write on the whole zone.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// Drop any cached copy before the write too: a create existence-check or
	// refresh earlier in this run may have cached a snapshot that this write
	// is about to make stale.
	c.invalidateZone(zoneID)

	zs, err := json.Marshal(&ZoneStream{
		Adds: newRecords,
		Rems: oldRecords,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/zone/%s/_stream", c.HostURL, zoneInfo[0]), strings.NewReader(string(zs)))
	if err != nil {
		return err
	}

	_, err = request[any](c, req)
	if err != nil {
		return err
	}

	// The zone changed, so any cached copy is now stale.
	c.invalidateZone(zoneID)

	return nil
}

// DeleteRecord sends a delete request to the API.
func (c *Client) DeleteRecords(ctx context.Context, zoneID string, records []Record) error {
	zoneInfo := strings.Split(zoneID, "@")
	if len(zoneInfo) != 2 {
		return fmt.Errorf("the zone is must have the format origin@virtualNameServer")
	}

	// Serialize mutations: _stream is a read-modify-write on the whole zone.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// Drop any cached copy before the write too: a create existence-check or
	// refresh earlier in this run may have cached a snapshot that this write
	// is about to make stale.
	c.invalidateZone(zoneID)

	zs, err := json.Marshal(&ZoneStream{
		Rems: records,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/zone/%s/_stream", c.HostURL, zoneInfo[0]), strings.NewReader(string(zs)))
	if err != nil {
		return err
	}

	_, err = request[any](c, req)
	if err != nil {
		return err
	}

	// The zone changed, so any cached copy is now stale.
	c.invalidateZone(zoneID)

	return nil
}
