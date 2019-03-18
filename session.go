package fortnitego

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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

	username      string
	password      string
	launcherToken string
	gameToken     string

	mux sync.Mutex
}

// Create opens a new connection to Epic and authenticates into the game to obtain the necessary access tokens.
func Create(username, password, launcherToken, gameToken string, use_proxy bool) (*Session, error) {
	// Initialize a new client for this session to make requests with.
	c := newClient(use_proxy)

	// Prepare form to request access token for launcher.
	data := url.Values{}
	data.Add("grant_type", "password")
	data.Add("username", username)
	data.Add("password", password)
	data.Add("includePerms", "true")

	// Prepare request.
	req, err := c.NewRequest(http.MethodPost, oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// Set authorization header to use launcher token.
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBasic, launcherToken))

	// Process request and decode response into tokenResponse.
	tr := &tokenResponse{}
	resp, err := c.Do(req, tr)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	///////////////////
	// Prepare new request for OAUTH exchange.
	req, err = c.NewRequest(http.MethodGet, oauthExchangeURL, nil)
	if err != nil {
		return nil, err
	}

	// Set authorization header to use the access token just retrieved.
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBearer, tr.AccessToken))

	// Process request and decode response into exchangeResponse.
	er := &exchangeResponse{}
	resp, err = c.Do(req, er)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	////////////////////
	// Prepare new form for 2nd OAUTH token request for game client.
	data = url.Values{}
	data.Add("grant_type", "exchange_code")
	data.Add("exchange_code", er.Code)
	data.Add("includePerms", "true")
	data.Add("token_type", "eg1") // should this be eg1???

	req, err = c.NewRequest(http.MethodPost, oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// Set authorization header to use the game token.
	req.Header.Set("Authorization", fmt.Sprintf("%v %v", AuthBasic, gameToken))

	// Perform request.
	resp, err = c.Do(req, tr)
	if err != nil {
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

		username:      username,
		password:      password,
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
	resp, err := s.client.Do(req, tr)
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

	_, err = s.client.Do(req, nil)
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
