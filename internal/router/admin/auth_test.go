package admin

import (
	"encoding/binary"
	"testing"

	"github.com/parsa222/ECSS-Lockers/internal/crypto"
	"github.com/parsa222/ECSS-Lockers/internal/time"
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
