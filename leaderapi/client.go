package leaderapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

// Client is a client for the leader API socket.
type Client struct {
	cli *http.Client
}

// NewClient creates a new Client.
func NewClient(path string) (*Client, error) {
	// Check the socket path exists and is a socket.
	// Note that os.ModeSocket might not be set on Windows.
	// (https://github.com/golang/go/issues/33357)
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(LeaderSocketPath)
		if err != nil {
			return nil, fmt.Errorf("stat socket: %w", err)
		}
		if fi.Mode()&os.ModeSocket == 0 {
			return nil, fmt.Errorf("%q is not a socket", LeaderSocketPath)
		}
	}

	// Try to connect to the socket.
	test, err := net.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("socket test connection: %w", err)
	}
	test.Close()

	dialer := net.Dialer{}
	return &Client{
		cli: &http.Client{
			Transport: &http.Transport{
				// Ignore arguments, dial socket
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return dialer.DialContext(ctx, "unix", LeaderSocketPath)
				},
			},
		},
	}, nil
}

// do implements the common bits of an API call. req is serialised to JSON and
// passed as the request body if not nil. The method is called, with the token
// added in the Authorization header. The response is deserialised, either into
// the object passed into resp if the status is 200 OK, otherwise into an error.
func (c *Client) do(ctx context.Context, method, url string, req, resp any) error {
	var body io.Reader
	if req != nil {
		buf, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshalling request: %w", err)
		}
		body = bytes.NewReader(buf)
	}

	hreq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("creating a request: %w", err)
	}

	hresp, err := c.cli.Do(hreq)
	if err != nil {
		return err
	}
	defer hresp.Body.Close()
	dec := json.NewDecoder(hresp.Body)

	if hresp.StatusCode != 200 {
		var er ErrorResponse
		if err := dec.Decode(&er); err != nil {
			return fmt.Errorf("decoding error response: %w", err)
		}
		return fmt.Errorf("error from job executor: %s", er.Error)
	}

	if resp == nil {
		return nil
	}
	if err := dec.Decode(resp); err != nil {
		return fmt.Errorf("decoding response: %w:", err)
	}
	return nil
}

// Get gets the current value of the lock key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	uk := url.PathEscape(key)

	var resp ValueResponse
	if err := c.do(ctx, "GET", "http://agent/api/leader/v0/lock/"+uk, nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// CompareAndSwap atomically compares-and-swaps the old value for the new value
// or performs no modification. It returns the most up-to-date value for the
// key, and reports whether the new value was written.
func (c *Client) CompareAndSwap(ctx context.Context, key, old, new string) (string, bool, error) {
	uk := url.PathEscape(key)

	req := LockCASRequest{
		Old: old,
		New: new,
	}
	var resp LockCASResponse
	if err := c.do(ctx, "GET", "http://agent/api/leader/v0/lock/"+uk, &req, &resp); err != nil {
		return "", false, err
	}
	return resp.Value, resp.Swapped, nil
}
