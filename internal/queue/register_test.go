package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBroker(t *testing.T) {
	assert := assert.New(t)

	b, err := NewBroker("badproto://badurl")
	assert.NotNil(err)
	assert.Nil(b)

	b, err = NewBroker("foo://host%10")
	assert.Error(err)
	assert.Nil(b)
}
