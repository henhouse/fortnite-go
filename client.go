package fortnitego

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"runtime"
	"time"
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

func newClient(use_proxy bool, cookies []*http.Cookie) *Client {
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
		cookieJar, _ := cookiejar.New(nil)
		if cookies != nil {
			cookie_url, _ := url.Parse("https://epicgames.com/id")
			cookieJar.SetCookies(cookie_url, cookies)
		}
		return &Client{client: &http.Client{Transport: torTransport, Timeout: time.Second * 120, Jar: cookieJar}}
	}
	cookieJar, _ := cookiejar.New(nil)
	if cookies != nil {
		cookie_url, _ := url.Parse("https://epicgames.com/id")
		cookieJar.SetCookies(cookie_url, cookies)
	}
	return &Client{client: &http.Client{Jar: cookieJar}}
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
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, int, error) {
	// Process request using session's client. Collect response.
	resp, err := c.client.Do(req)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, err
	}

	// Check response status codes to determine success/failure.
	status_code, err := checkStatus(resp)
	if err != nil {
		return nil, status_code, err
	}

	// If an interface was provided, decode response body into it.
	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil && err != io.EOF {
			log.Println("ERR: ", err)
			return resp, 0, err
		}
	}

	return resp, 0, nil
}

// checkStatus checks the HTTP response status code for unsuccessful requests.
// @todo decode error into Epic Error-JSON object to determine better errors.go?
func checkStatus(resp *http.Response) (int, error) {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return 0, nil
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.New("unsuccessful response returned and cannot read body: " + err.Error())
		}
		defer resp.Body.Close()

		status_code := resp.StatusCode

		return status_code, errors.New(string(b))
	}
}
