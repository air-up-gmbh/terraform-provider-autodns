package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ZoneFilter is used to add a filter to the zones API request.
type ZoneFilter struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Operator string `json:"operator"`
}

// ZoneFIlterReq is a set of filters to send with the API request.
type ZoneFilterReq struct {
	Filters []ZoneFilter `json:"filters"`
}

// Zone desribes the zone object in the autodns API response.
type Zone struct {
	Origin            string   `json:"origin"`
	NameServerGroup   string   `json:"nameServerGroup"`
	VirtualNameServer string   `json:"virtualNameServer"`
	Records           []Record `json:"resourceRecords"`
}

// GetZone returns the zone in autodns matching the origin.
func (c *Client) GetZone(ctx context.Context, origin string) (*Zone, error) {
	zf, err := json.Marshal(&ZoneFilterReq{
		Filters: []ZoneFilter{
			{
				Key:      "origin",
				Value:    origin,
				Operator: "EQUAL",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/zone/_search", c.HostURL), strings.NewReader(string(zf)))
	if err != nil {
		return nil, err
	}

	res, err := request[Zone](c, req)
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("origin does not exist or more than one result has been returned by the API")
	}

	return &res[0], nil
}
