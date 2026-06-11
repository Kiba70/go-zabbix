package zabbix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ErrNotFound describes an empty result set for an API call.
var ErrNotFound = errors.New("No results were found matching the given search parameters")

// A Session is an authenticated Zabbix JSON-RPC API client. It must be
// initialized and connected with NewSession.
type Session struct {
	// URL of the Zabbix JSON-RPC API (ending in `/api_jsonrpc.php`).
	URL string `json:"url"`

	// Token is the cached authentication token returned by `user.login` and
	// used to authenticate all API calls in this Session.
	Token string `json:"token"`

	// ApiVersion is the software version string of the connected Zabbix API.
	APIVersion string `json:"apiVersion"`
	ApiVersion APIVersion

	client *http.Client
}

type APIVersion struct {
	Major int
	Minor int
	Build int
}

// NewSession returns a new Session given an API connection URL and an API
// username and password.
//
// An error is returned if there was an HTTP protocol error, the API credentials
// are incorrect or if the API version is indeterminable.
//
// The authentication token returned by the Zabbix API server is cached to
// authenticate all subsequent requests in this Session.
func NewSession(ctx context.Context, url, username, password, token string) (session *Session, err error) {
	// create session
	session = &Session{URL: url}
	err = session.login(ctx, username, password, token)
	return
}

// NewSessionToken returns a new Session given an API connection URL and an API Token
//
// An error is returned if there was an HTTP protocol error, the API credentials
// are incorrect or if the API version is indeterminable.
//
// The authentication token returned by the Zabbix API server is cached to
// authenticate all subsequent requests in this Session.
func NewSessionToken(ctx context.Context, url string, token string) (session *Session, err error) {
	// create session
	session = &Session{URL: url, Token: token}

	// get Zabbix API version
	_, err = session.GetVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve Zabbix API version: %v", err)
	}

	return
}

func (c *Session) login(ctx context.Context, username, password, token string) error {
	// get Zabbix API version
	_, err := c.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("Failed to retrieve Zabbix API version: %v", err)
	}

	if token != "" {
		c.Token = token
		return nil
	}

	// login to API
	params := map[string]string{
		"username": username,
		"password": password,
	}

	res, err := c.Do(ctx, NewRequest("user.login", params))
	if err != nil {
		return fmt.Errorf("Error logging in to Zabbix API: %v", err)
	}

	err = res.Bind(&c.Token)
	if err != nil {
		return fmt.Errorf("Error failed to decode Zabbix login response: %v", err)
	}

	return nil
}

// GetVersion returns the software version string of the connected Zabbix API.
func (c *Session) GetVersion(ctx context.Context) (string, error) {
	if c.APIVersion == "" {
		// get Zabbix API version
		res, err := c.Do(ctx, NewRequest("apiinfo.version", nil))
		if err != nil {
			return "", err
		}

		err = res.Bind(&c.APIVersion)
		if err != nil {
			return "", err
		}

		versions := strings.Split(c.APIVersion, ".")
		c.ApiVersion.Major, _ = strconv.Atoi(versions[0])
		c.ApiVersion.Minor, _ = strconv.Atoi(versions[1])
		c.ApiVersion.Build, _ = strconv.Atoi(versions[2])
	}
	return c.APIVersion, nil
}

// AuthToken returns the authentication token used by this session to
// authentication all API calls.
func (c *Session) AuthToken() string {
	return c.Token
}

// Set new API Token for session
func (c *Session) SetToken(token string) {
	c.Token = token
}

// Do sends a JSON-RPC request and returns an API Response, using connection
// configuration defined in the parent Session.
//
// An error is returned if there was an HTTP protocol error, a non-200 response
// is received, or if an error code is set is the JSON response body.
//
// When err is nil, resp always contains a non-nil resp.Body.
//
// Generally Get or a wrapper function will be used instead of Do.
func (c *Session) Do(ctx context.Context, req *Request) (resp *Response, err error) {
	// configure request
	req.AuthToken = c.Token

	// encode request as json
	b, err := json.Marshal(req)
	if err != nil {
		return
	}

	dprintf("Call     [%s:%d]: %s\n", req.Method, req.RequestID, b)

	// create HTTP request
	r, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(b))
	if err != nil {
		return
	}
	r.ContentLength = int64(len(b))
	r.Header.Add("Content-Type", "application/json-rpc")

	// send request
	client := c.client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(r)
	if err != nil {
		return
	}

	defer res.Body.Close()

	// read response body
	b, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}

	dprintf("Response [%s:%d]: %s\n", req.Method, req.RequestID, b)

	// map HTTP response to Response struct
	resp = &Response{
		StatusCode: res.StatusCode,
	}

	// unmarshal response body
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON response body: %v", err)
	}

	// check for API errors
	if err = resp.Err(); err != nil {
		return
	}

	return
}

// Get calls the given Zabbix API method with the given query parameters and
// unmarshals the JSON response body into the given interface.
//
// An error is return if a transport, marshalling or API error happened.
func (c *Session) Get(ctx context.Context, method string, params interface{}, v interface{}) error {
	req := NewRequest(method, params)
	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}

	err = resp.Bind(v)
	if err != nil {
		return err
	}

	return nil
}
