package tests

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/botscommunity/botsgo"
)

type VKAPI struct {
	Client  *botsgo.Client
	Version float64
	Token   string
	Limit   int
	mutex   sync.Mutex
	time    time.Time
	rps     int
}

const (
	VKAPIDefaultVersion    = 5.199
	VKAPICommunityRPSLimit = 19
)

func NewVKAPI(token string) (*VKAPI, error) {
	client, err := botsgo.NewClient("https://api.vk.com")
	client.Logger = zap.NewNop()

	return &VKAPI{
		Client:  client,
		Version: VKAPIDefaultVersion,
		Token:   token,
		Limit:   VKAPICommunityRPSLimit,
	}, err
}

type VKAPIUsers struct {
	Error *VKAPIError `json:"error,omitempty"`
	Users []VKAPIUser `json:"response,omitempty"`
}

type VKAPIUser struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"first_name,omitempty"`
	Surname string `json:"last_name,omitempty"`
}

type VKAPIError struct {
	Code    int          `json:"error_code,omitempty"`
	Message string       `json:"error_msg,omitempty"`
	Params  []ErrorParam `json:"request_params,omitempty"`
}

type ErrorParam struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (api *VKAPI) GetUser(userID int) ([]VKAPIUser, *VKAPIError) {
	response := VKAPIUsers{}

	queryParams := url.Values{}
	queryParams.Set("user_id", fmt.Sprint(userID))

	if err := api.Call(queryParams.Encode(), &response); err != nil {
		return response.Users, &VKAPIError{
			Message: err.Error(),
		}
	}

	if response.Error != nil {
		return response.Users, &VKAPIError{
			Message: response.Error.Message,
		}
	}

	return response.Users, nil
}

func (api *VKAPI) Call(queryParams string, response any) error {
	api.mutex.Lock()

	sleepTime := time.Second - time.Since(api.time)
	if sleepTime < 0 {
		api.time = time.Now()
		api.rps = 0
	} else if api.rps == api.Limit {
		time.Sleep(sleepTime)
		api.time = time.Now()
		api.rps = 0
	}

	api.rps++
	api.mutex.Unlock()

	req, err := api.Client.NewRequest(context.Background())
	if err != nil {
		return err
	}

	req.Method(http.MethodGet)
	req.Path(fmt.Sprintf("/method/users.get?v=%f&random_id=%d&%s", api.Version, time.Now().Unix(), queryParams))
	req.Response(&response)

	req.SetHeader("Authorization", "Bearer "+api.Token)

	res, err := req.Do()
	if err != nil {
		return err
	}

	if err := res.Body.Close(); err != nil {
		return err
	}

	return nil
}
