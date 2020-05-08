package threebotconnect

import (
	"crypto/ed25519"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	_, private, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	auth := Auth{
		AppID:      "appID",
		PrivateKey: private,
	}

	auth.Login(w, req)

	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)

	u, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	q := u.Query()
	assert.Equal(t, q.Get("appid"), "appID")
	assert.Equal(t, q.Get("redirecturl"), "/callback_threebot")
	assert.NotEmpty(t, q.Get("state"))
}
