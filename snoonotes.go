package snoonotes

import (
	"io"

	"github.com/sirupsen/logrus"
)

const baseURL = "https://snoonotes.com/"

var tokens tokensStruct
var log = logrus.WithField("prefix", "SnooNotes")

func init() {
	tokens = tokensStruct{t: make(map[string]authToken)}
}

func safeClose(c io.Closer) {
	if cerr := c.Close(); cerr != nil {
		log.WithError(cerr).Error("closing body")
	}
}
