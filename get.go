package snoonotes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Get returns notes as a specific user about another user.
// If sub is empty notes for all subreddits are returned.
func Get(sub, as, about string) (*[]Note, error) {
	l := log.WithField("action", "Get").WithField("username", as).WithField("target", about)

	r, err := getAuthedRequest(as, "POST", "api/Note/GetNotes", strings.NewReader(fmt.Sprintf(`["%s"]`, about)))
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	var resp *http.Response
	resp, err = client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "http request failed")
	}
	defer safeClose(resp.Body)

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "http body read failed")
	}

	var ns getNotes
	err = json.Unmarshal(body, &ns)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	for k, n := range ns {
		if strings.EqualFold(k, about) {
			// filter out notes for one subreddit only!
			fn := n[:0]
			for _, sn := range n {
				if len(sub) == 0 || sn.SubName == sub {
					fn = append(fn, sn)
				}
			}
			l.Debugf("got %d (of %d) notes", len(fn), len(n))
			return &fn, nil
		}
	}

	l.Debug("no notes found")

	// no notes
	return nil, nil
}

var configCache = make(map[string]subConfigCache)

type subConfigCache struct {
	t time.Time
	s SubSettings
}

// GetConfig returns the configuration for a given subreddit.
// The result is cached for 24 hours to prevent repeat requests.
func GetConfig(as, sub string) (*SubSettings, error) {
	l := log.WithField("action", "GetConfig").WithField("username", as).WithField("sub", sub)

	if s, ok := configCache[sub]; ok && time.Now().Before(s.t) {
		l.Debugf("got cached, still valid until %s", s.t)
		return &s.s, nil
	}

	url := "restapi/Subreddit"
	if sub == "" {
		return nil, errors.New("can only fetch config for one subreddit at a time")
	}
	url += "/" + sub

	r, err := getAuthedRequest(as, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	var resp *http.Response
	resp, err = client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "http request failed")
	}
	defer safeClose(resp.Body)

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "http body read failed")
	}

	var s []SubSettings
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	for _, subreddit := range s {
		if strings.EqualFold(subreddit.SubName, sub) {
			l.Debug("got config")
			c := subConfigCache{t: time.Now().Add(24 * time.Hour), s: subreddit}
			configCache[sub] = c
			return &subreddit, nil
		}
	}

	l.Warn("no config found")

	return nil, errors.New("no config found")
}

// GetNoteTypes returns all possible note types.
// Sub is required.
func GetNoteTypes(as, sub string) (*[]NoteType, error) {
	l := log.WithField("action", "GetNoteTypes").WithField("username", as).WithField("sub", sub)

	settings, err := GetConfig(as, sub)
	if err != nil {
		l.WithError(err).Warn("couldn't get sub config")
		return nil, err
	}
	ret := []NoteType{}
	for _, note := range settings.Settings.NoteTypes {
		if len(sub) > 0 && !strings.EqualFold(note.SubName, sub) {
			continue
		}
		ret = append(ret, note)
	}

	l.Debugf("found %d note types", len(ret))

	return &ret, nil
}

// GetNoteTypeMap is a helper that returns note type IDs mapped their main values (DisplayName & Color)
// for easy display in applications. Use GetNoteTypes if you need detailed informations.
// Sub is required.
func GetNoteTypeMap(as, sub string) (map[int]SimpleNoteType, error) {
	l := log.WithField("action", "GetNoteTypeMap").WithField("username", as).WithField("sub", sub)

	notes, err := GetNoteTypes(as, sub)
	if err != nil {
		l.WithError(err).Warn("couldn't get note types")
		return nil, err
	}
	ret := make(map[int]SimpleNoteType)
	for _, note := range *notes {
		ret[note.NoteTypeID] = SimpleNoteType{note.DisplayName, note.ColorCode}
	}

	l.Debugf("found %d note types", len(ret))

	return ret, nil
}
