package tests_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVK(t *testing.T) {
	t.Run("VK bot", func(t *testing.T) {
		bot := NewVKBot()
		user := bot.GetUser(User{UserID: 1})
		assert.Equal(t, true, user.Error == nil, user.Error)
	})
}
