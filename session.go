package fortnitego

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"net/http/cookiejar"
)

// Session holds connection information regarding a successful authentication with an Epic account to Epic's API. Will
// hold a Client to use for interfacing with said API, and retain information about our authenticated session with them.
type Session struct {
	client *Client

	AccessToken  string
	ExpiresAt    string
	RefreshToken string
	AccountID    string
	ClientID     string

	email         string
	password      string
	launcherToken string
	gameToken     string

	mux sync.Mutex
}

// Create opens a new connection to Epic and authenticates into the game to obtain the necessary access tokens.
func Create(email string, password string, launcherToken string, gameToken string, use_proxy bool) (*Session, int, []*http.Cookie, error) {
	// Initialize a new client for this session to make requests with.
	c := newClient(use_proxy, nil)

	// CSRF
	req, err := c.NewRequest(http.MethodGet, csrfUrl, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	resp, _, err := c.Do(req, nil)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, nil, err
	}
	resp.Body.Close()
	cookieJar, _ := cookiejar.New(nil)
	cookie_url, _ := url.Parse("http://epicgames.com/id")
	cookieJar.SetCookies(cookie_url, resp.Cookies())

	csrf := ""
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "XSRF-TOKEN" {
			csrf = cookie.Value
		}
	}

	// FIRST TOKEN
	data := url.Values{}
	data.Add("email", email)
	data.Add("password", password)
	data.Add("rememberMe", "true")

	req, err = c.NewRequest(http.MethodPost, loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, nil, err
	}
	req.Header.Set("x-xsrf-token", csrf)

	tr := &tokenResponse{}
	resp, status_code, err := c.Do(req, tr)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, status_code, cookies, err
	}
	resp.Body.Close()

	// EXCHANGE
	req, err = c.NewRequest(http.MethodGet, oauthExchangeURL, nil)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, nil, err
	}
	req.Header.Set("x-xsrf-token", csrf)

	er := &exchangeResponse{}
	resp, _, err = c.Do(req, er)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, nil, err
	}
	resp.Body.Close()

	// TOKEN 2
	data = url.Values{}
	data.Add("grant_type", "exchange_code")
	data.Add("exchange_code", er.Code)
	data.Add("includePerms", "true")
	data.Add("token_type", "eg1")

	req, err = c.NewRequest(http.MethodPost, oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBasic, launcherToken))

	tr = &tokenResponse{}
	resp, _, err = c.Do(req, tr)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, 0, nil, err
	}
	resp.Body.Close()

	// Create new session object from data retrieved.
	ret := &Session{
		client:       c,
		AccessToken:  tr.AccessToken,
		ExpiresAt:    tr.ExpiresAt,
		RefreshToken: tr.RefreshToken,
		AccountID:    tr.AccountID,
		ClientID:     tr.ClientID,

		email:         email,
		password:      password,
		launcherToken: launcherToken,
		gameToken:     gameToken,
	}

	// Spawn goroutine to handle automatic renewal of access token.
	go ret.renewProcess()

	log.Println("Session successfully created.")
	return ret, 0, nil, nil
}

// Create opens a new connection to Epic and authenticates into the game to obtain the necessary access tokens.
func Create2fa(code string, cookies []*http.Cookie, launcherToken string, gameToken string) (*Session, error) {
	// Initialize a new client for this session to make requests with.
	c := newClient(false, cookies)

	cookieJar, _ := cookiejar.New(nil)
	cookie_url, _ := url.Parse("http://epicgames.com/id")
	cookieJar.SetCookies(cookie_url, cookies)

	csrf := ""
	for _, cookie := range cookies {
		if cookie.Name == "XSRF-TOKEN" {
			csrf = cookie.Value
		}
	}

	// CSRF
	req, err := c.NewRequest(http.MethodGet, csrfUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-xsrf-token", csrf)

	resp, _, err := c.Do(req, nil)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	resp.Body.Close()

	// cookieJar.SetCookies(cookie_url, resp.Cookies())
	cookies = resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "XSRF-TOKEN" {
			csrf = cookie.Value
		}
	}

	// FIRST TOKEN
	data := url.Values{}
	data.Add("code", code)
	data.Add("method", "authenticator")
	data.Add("rememberDevice", "false")

	req, err = c.NewRequest(http.MethodPost, mfaUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	req.Header.Set("x-xsrf-token", csrf)

	tr := &tokenResponse{}
	resp, _, err = c.Do(req, tr)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	resp.Body.Close()

	// EXCHANGE
	req, err = c.NewRequest(http.MethodGet, oauthExchangeURL, nil)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	req.Header.Set("x-xsrf-token", csrf)

	er := &exchangeResponse{}
	resp, _, err = c.Do(req, er)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	resp.Body.Close()

	// TOKEN 2
	data = url.Values{}
	data.Add("grant_type", "exchange_code")
	data.Add("exchange_code", er.Code)
	data.Add("includePerms", "true")
	data.Add("token_type", "eg1")

	req, err = c.NewRequest(http.MethodPost, oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBasic, launcherToken))

	tr = &tokenResponse{}
	resp, _, err = c.Do(req, tr)
	if err != nil {
		log.Println("ERR: ", err)
		return nil, err
	}
	resp.Body.Close()

	// Create new session object from data retrieved.
	ret := &Session{
		client:       c,
		AccessToken:  tr.AccessToken,
		ExpiresAt:    tr.ExpiresAt,
		RefreshToken: tr.RefreshToken,
		AccountID:    tr.AccountID,
		ClientID:     tr.ClientID,

		launcherToken: launcherToken,
		gameToken:     gameToken,
	}

	// Spawn goroutine to handle automatic renewal of access token.
	go ret.renewProcess()

	log.Println("Session successfully created.")
	return ret, nil
}

// Refresh renews a session by obtaining a new access token, and replacing the hold one. Intended use it for an
// automatic goroutine to handle scheduling of renewal. Previous token is automatically invalidated on Epic's end.
func (s *Session) Refresh() error {
	data := url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", s.RefreshToken)
	data.Add("includePerms", "true")

	// Prepare new request containing encoded body.
	req, err := s.client.NewRequest(http.MethodPost, oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	// Use game token for authorization.
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBasic, s.gameToken))

	// Perform request and collect response into response object.
	tr := &tokenResponse{}
	resp, _, err := s.client.Do(req, tr)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	// Assign token information to session.
	s.mux.Lock()
	s.AccessToken = tr.AccessToken
	s.RefreshToken = tr.RefreshToken
	s.ExpiresAt = tr.ExpiresAt
	s.mux.Unlock()

	return nil
}

// Kill terminates an existing session by sending a DELETE request to deactivate the session on Epic's servers.
func (s *Session) Kill() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	req, err := s.client.NewRequest(http.MethodDelete, killSessionURL+"/"+s.AccessToken, nil)
	if err != nil {
		return err
	}

	// Set authentication header to use access token.
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBearer, s.AccessToken))

	_, _, err = s.client.Do(req, nil)
	if err != nil {
		return err
	}

	// Clear session information.
	s.AccessToken = ""
	s.ExpiresAt = ""
	s.RefreshToken = ""

	log.Println("Session token successfully deactivated.")
	return nil
}

// renewProcess is a goroutine intended to be running during the lifetime of a Session. Its intention is to handle
// automatic renewal of access token within a necessary time for renewal to ensure the API stays connected and
// functional.
func (s *Session) renewProcess() {
	// Check every 20 seconds if we need to update access token.
	updateChecker := time.NewTimer(time.Second * 20)

	// Locks until timer above has passed.
	<-updateChecker.C

	// Parse expiration time into Time object.
	expiresAt, err := time.Parse(time.RFC3339, s.ExpiresAt)
	if err != nil {
		log.Println(err)
		return
	}

	// If the token expiration time does not expire within the next minute, wait and try again.
	if !time.Now().After(expiresAt.Add(-time.Minute - 1)) {
		defer s.renewProcess()
		return
	}

	// Token expiry is imminent, renew.
	err = s.Refresh()
	if err != nil {
		log.Println("Token renewal unsuccessful: " + err.Error())
	}

	log.Println("Token renewed successfully.")
	defer s.renewProcess()
}
