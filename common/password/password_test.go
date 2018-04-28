package password

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAccountPassword(t *testing.T) {
	var password, err = GetAccountPassword()
	assert.Nil(t, password)
	assert.NotNil(t, err)
	password, err = GetPassword()
	assert.Nil(t, password)
	assert.NotNil(t, err)
	password, err = GetConfirmedPassword()
	assert.Nil(t, password)
	assert.NotNil(t, err)
}
