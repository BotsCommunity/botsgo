package tests_test

import (
	"os"
	"testing"

	"github.com/botscommunity/botsgo/tests"

	"github.com/stretchr/testify/assert"
)

func BenchmarkVKAPI(b *testing.B) {
	api, err := tests.NewVKAPI(os.Getenv("VK_TOKEN"))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := api.GetUser(1); err != nil {
			b.Fatal(err)
		}
	}
}

func TestVKAPI(t *testing.T) {
	api, apiError := tests.NewVKAPI(os.Getenv("VK_TOKEN"))
	if apiError != nil {
		t.Fatal(apiError)
	}

	users, err := api.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(users) == 0 {
		t.Fatal("user length <= 0")
	}

	assert.Equal(t, "Павел", users[0].Name, users)
}
