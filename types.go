package snoonotes

import (
	"sync"
	"time"
)

// ErrNoUser is returned when a user account was not found.
type ErrNoUser struct{}

func (e *ErrNoUser) Error() string {
	return "Unknown user."
}

type tokensStruct struct {
	sync.RWMutex
	t map[string]authToken
}

func (t *tokensStruct) set(user string, token authToken) {
	t.RLock()
	// if token is not changed, return
	if val, ok := t.t[user]; ok {
		if val == token {
			t.RUnlock()
			return
		}
	}
	t.RUnlock()

	t.Lock()
	t.t[user] = token
	t.Unlock()
}
func (t *tokensStruct) get(user string) (authToken, error) {
	t.RLock()
	defer t.RUnlock()

	if val, ok := t.t[user]; ok {
		return val, nil
	}

	return authToken{}, &ErrNoUser{}
}

// Response from auth/connect/token
type authToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint   `json:"expires_in"`
	TokenType   string `json:"token_type"`

	Expires time.Time `json:"-"`
	Key     string    `json:"-"`
}

// Response from api/Note/GetNotes
type getNotes map[string][]Note

// Note is a note returned from SnooNotes
type Note struct {
	NoteID          int
	NoteTypeID      int
	SubName         string
	Submitter       string
	Message         string
	URL             string `json:"Url"`
	TimeStamp       string
	ParentSubreddit *string
}

// SubSettings returns the settings for a specific subreddit.
// Response from restapi/Subreddit
type SubSettings struct {
	SubredditID int
	SubName     string
	Active      bool
	BotSettings *struct {
		DirtBagURL      string `json:"DirtbagUrl"`
		DirtbagUsername string
	}
	Settings struct {
		AccessMask int
		NoteTypes  []NoteType
		PermBanID  *int
		TempBanID  *int
	}
}

// NoteType defines a type of note.
type NoteType struct {
	NoteTypeID   int
	SubName      string
	DisplayName  string
	ColorCode    string
	DisplayOrder int
	Bold         bool
	Italic       bool
}

// SimpleNoteType defines a simple type of note.
type SimpleNoteType struct {
	DisplayName string
	ColorCode   string
}

// NewNote is used to submit new notes to SnooNotes.
// It has some different fields than Note.
// Request to api/note
type NewNote struct {
	NoteTypeID        int
	SubName           string
	Message           string
	AppliesToUsername string
	URL               string `json:"url"`
}
