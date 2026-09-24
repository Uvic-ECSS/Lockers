package admin

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Uvic-ECSS/Lockers/internal/crypto"
	"github.com/Uvic-ECSS/Lockers/internal/time"
	"github.com/stretchr/testify/assert"
)

func TestAdminToken(t *testing.T) {
	t.Parallel()

	adminUsername = "foo"
	adminPassword = "bar"

	token, err := makeToken()
	assert.Nil(t, err)
	assert.True(t, validToken(token))

	// token test
	assert.False(t, validToken(""))
	assert.False(t, validToken("not-a-real-token"))
	assert.False(t, validToken(token+"tampered"))

	// expired token
	old := make([]byte, 8)
	binary.BigEndian.PutUint64(old, uint64(time.Now().Unix())-uint64(adminSessionMaxAge)-100)
	ct, err := crypto.Encrypt(crypto.CipherKey[:], old, []byte(adminUsername))
	assert.Nil(t, err)
	assert.False(t, validToken(crypto.Base64.EncodeToString(ct)))
}

func TestAdminTokenCheckerHtmxRedirect(t *testing.T) {
	called := false
	h := AdminTokenChecker(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/history", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.False(t, called, "handler ran without a valid session")
	assert.Equal(t, "/auth/admin", rec.Header().Get("HX-Redirect"))
}
