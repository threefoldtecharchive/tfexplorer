package phonebook

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

func TestErrWrongPubKey(t *testing.T) {
	err := errWrongPubKey{
		name:   "foo",
		pubkey: "bar",
	}

	assert.True(t, errors.As(err, &errWrongPubKey{}))
}
