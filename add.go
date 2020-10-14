package snoonotes

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Add adds a new note as a specific user.
func Add(as string, note NewNote) error {
	l := log.WithField("action", "Add").WithField("username", as)

	js, err := json.Marshal(note)
	if err != nil {
		return errors.Wrap(err, "json marshal failed")
	}
	var r *http.Request
	r, err = getAuthedRequest(as, "POST", "api/note", strings.NewReader(string(js)))
	if err != nil {
		return errors.Wrap(err, "getting authed request failed")
	}
	client := http.DefaultClient
	var resp *http.Response
	resp, err = client.Do(r)
	if err != nil {
		return errors.Wrap(err, "http request failed")
	}
	defer safeClose(resp.Body)

	l.Debug("added note")

	// nothing returned

	return nil
}
