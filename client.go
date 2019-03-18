package fortnitego

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
 	"os"
	"time"
	"runtime"
	"golang.org/x/net/proxy"
)

// Client is our HTTP client for this package to interface with Epic's API.
type Client struct {
	client *http.Client
}

// Version is the package version.
const Version = "0.1"

// userAgent to represent ourselves on request. Is not spoofed due to uncertainty on usage of game API.
var userAgent = fmt.Sprintf(
	"fortnite-go/v%s Go-http-client/%s (%s %s)",
	Version, runtime.Version(), runtime.GOOS, runtime.GOARCH,
)

func newClient(use_proxy bool) *Client {
	// Return default HTTP client for now. @todo replace with defined client
	if use_proxy {
		// Setup localhost TOR proxy
	 	torProxyUrl, err := url.Parse("socks5://127.0.0.1:9050") // port 9150 is for Tor Browser
	 	if err != nil {
	 		fmt.Println("Unable to parse URL:", err)
	 		os.Exit(-1)
	 	}

	 	// Setup a proxy dialer
	 	torDialer, err := proxy.FromURL(torProxyUrl, proxy.Direct)
	 	if err != nil {
	 		fmt.Println("Unable to setup Tor proxy:", err)
	 		os.Exit(-1)
	 	}

	 	torTransport := &http.Transport{Dial: torDialer.Dial}
		return &Client{client: &http.Client{Transport: torTransport, Timeout: time.Second * 120}}
	}
	return &Client{client: &http.Client{}}
}

// NewRequest prepares a new HTTP request and sets the necessary headers.
func (c *Client) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	// Prepare new request.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// If we're sending data, set appropriate content type.
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Set user agent.
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

// Authentication header types
var (
	AuthBearer = "Bearer"
	AuthBasic  = "Basic"
)

// Do processes a prepared HTTP request with the client provided. An interface is passed in to decode the response into.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// Process request using session's client. Collect response.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Check response status codes to determine success/failure.
	err = checkStatus(resp)
	if err != nil {
		return nil, err
	}

	// If an interface was provided, decode response body into it.
	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil && err != io.EOF {
			return resp, err
		}
	}

	return resp, nil
}

// checkStatus checks the HTTP response status code for unsuccessful requests.
// @todo decode error into Epic Error-JSON object to determine better errors.go?
func checkStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New("unsuccessful response returned and cannot read body: " + err.Error())
		}
		defer resp.Body.Close()

		return errors.New(fmt.Sprintf("unsuccessful response returned: %v %v", resp.StatusCode, string(b)))
	}
}
