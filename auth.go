package snoonotes

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// Auth authenticates a user with SnooNotes.
// You can get the required key from https://snoonotes.com/#!/userkey
func Auth(username, key string) error {
	l := log.WithField("action", "Auth").WithField("username", username)

	l.Debug("sending auth request")

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", key)
	form.Set("client_id", "bots")

	resp, err := http.PostForm(baseURL+"auth/connect/token", form)
	if err != nil {
		return errors.Wrap(err, "http request failed")
	}

	defer safeClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("http request returned status [%s]", resp.Status)
	}

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "http body read failed")
	}

	var t authToken
	err = json.Unmarshal(body, &t)
	if err != nil {
		return errors.Wrap(err, "json unmarshal failed")
	}

	t.Expires = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	t.Key = key

	log.Debug("got auth token")

	tokens.set(username, t)

	return nil
}

func refreshAccessToken(username string, t authToken) (authToken, error) {
	if err := Auth(username, t.Key); err != nil {
		return authToken{}, errors.Wrap(err, "refreshing token failed")
	}
	return tokens.get(username)
}

func getAuthedRequest(username, method, url string, body io.Reader) (*http.Request, error) {
	l := log.WithField("username", username).WithField("method", method).WithField("url", url)

	t, err := tokens.get(username)
	if err != nil {
		return nil, err
	}
	if time.Now().After(t.Expires) {
		l.Debug("token expired, refreshing")
		if t, err = refreshAccessToken(username, t); err != nil {
			return nil, errors.Wrap(err, "could not refresh token")
		}
	}

	var req *http.Request
	req, err = http.NewRequest(method, baseURL+url, body)
	if err != nil {
		return nil, errors.Wrap(err, "creating request failed")
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", t.TokenType, t.AccessToken))
	//log.WithFields(logrus.Fields{"uid": uid, "method": method, "url": url}).Debug("OAuth2Request")
	l.Debug("AuthedRequest created")

	return req, nil
}
