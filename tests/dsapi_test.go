package tests_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/botscommunity/botsgo/tests"
)

func BenchmarkDiscordAPI(b *testing.B) {
	api, err := tests.NewDiscordAPI(os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		b.Fatal(err)
	}

	channelID, err := strconv.Atoi(os.Getenv("DISCORD_CHANNEL_ID"))
	if err != nil {
		b.Fatal(err)
	}

	messageID, err := strconv.Atoi(os.Getenv("DISCORD_MESSAGE_ID"))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := api.GetMessage(channelID, messageID); err != nil {
			b.Fatal(err)
		}
	}
}

func TestDiscordAPI(t *testing.T) {
	api, err := tests.NewDiscordAPI(os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}

	channelID, err := strconv.Atoi(os.Getenv("DISCORD_CHANNEL_ID"))
	if err != nil {
		t.Fatal(err)
	}

	messageID, err := strconv.Atoi(os.Getenv("DISCORD_MESSAGE_ID"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := api.GetMessage(channelID, messageID); err != nil {
		t.Fatal(err)
	}
}
