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

	mu sync.Mutex

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
	}
}

func request[T any](c *Client, req *http.Request) ([]T, error) {
	// Lock the mutex to avoid concurrent updates
	c.mu.Lock()
	defer c.mu.Unlock()

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
