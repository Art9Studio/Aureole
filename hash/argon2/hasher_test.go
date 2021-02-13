package argon2

import (
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"gouth/hash"
	"testing"
)

func Test_Argon2_Hash(t *testing.T) {
	rawHashData := hash.RawHashData{
		"mode":        "argon2i",
		"iterations":  1,
		"parallelism": 1,
		"salt_length": 1,
		"key_length":  16,
		"memory":      16384,
	}
	hasher, err := hash.New("argon2", rawHashData)
	assert.NoError(t, err)

	h, err := hasher.Hash("passwd")
	assert.NoError(t, err)
	assert.NotEmpty(t, h)
	_, _ = pretty.Print(h)
}
