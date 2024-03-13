package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/botscommunity/botsgo"
)

type DiscordAPI struct {
	Client         *botsgo.Client
	Version        int
	Token          string
	remainingLimit int
	resetAfterTime float64
	mutex          *sync.Mutex
}

const (
	DiscordAPIDefaultVersion = 10
	DiscordAPIDefaultLimit   = 1
	floatBitSize             = 64
	StatusOK                 = 200
)

var (
	errMessageEmpty      = errors.New("error message is not empty")
	errUnknownStatusCode = errors.New("received a status code other than OK")
)

func NewDiscordAPI(token string) (*DiscordAPI, error) {
	client, err := botsgo.NewClient("https://discord.com")
	client.Logger = zap.NewNop()

	return &DiscordAPI{
		Client:         client,
		Version:        DiscordAPIDefaultVersion,
		Token:          token,
		remainingLimit: DiscordAPIDefaultLimit,
		mutex:          new(sync.Mutex),
	}, err
}

type DiscordAPIMessage struct {
	Code      int    `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	ID        string `json:"ID,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

func (api *DiscordAPI) GetMessage(channelID int, messageID int) (DiscordAPIMessage, error) {
	response := DiscordAPIMessage{}

	if err := api.Call(http.MethodGet, fmt.Sprintf("channels/%d/messages/%d", channelID, messageID), &response); err != nil {
		return response, err
	}

	if response.Message != "" {
		return response, fmt.Errorf("%w: %s", errMessageEmpty, response.Message)
	}

	return response, nil
}

func (api *DiscordAPI) Call(method, path string, response any) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api.remainingLimit <= 0 {
		time.Sleep(time.Duration(api.resetAfterTime * float64(time.Second)))
	}

	req, err := api.Client.NewRequest(context.Background())
	if err != nil {
		return err
	}

	req.Method(method)
	req.Path(fmt.Sprintf("/api/v%d/%s", api.Version, path))
	req.Response(&response)

	req.SetHeader("Authorization", "Bot "+api.Token)
	req.SetHeader("User-Agent", "DiscordBot (https://github.com/botscommunity/dsgo, 1.0.0)")

	res, err := req.Do()
	if err != nil {
		if api.Client.Logger != nil {
			api.Client.Logger.Error("[ERR] Do request")
		}

		return err
	}

	if res.StatusCode != StatusOK {
		api.Client.Logger.Error(fmt.Sprintf("[ERR] %s", res.Status))

		return fmt.Errorf("%w: %s", errUnknownStatusCode, res.Status)
	}

	remainingLimit, err := strconv.Atoi(res.Header.Get("X-RateLimit-Remaining"))
	if err != nil {
		if api.Client.Logger != nil {
			api.Client.Logger.Error("[ERR] Atoi X-RateLimit-Remaining")
		}

		return err
	}

	resetAfterTime, err := strconv.ParseFloat(res.Header.Get("X-RateLimit-Reset-After"), floatBitSize)
	if err != nil {
		if api.Client.Logger != nil {
			api.Client.Logger.Error("[ERR] ParseFloat X-RateLimit-Reset-After")
		}

		return err
	}

	api.remainingLimit = remainingLimit
	api.resetAfterTime = resetAfterTime

	return res.Body.Close()
}
