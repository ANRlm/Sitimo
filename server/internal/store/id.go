package store

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}
