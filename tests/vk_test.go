package tests_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/botscommunity/botsgo"
)

type VKBot struct {
	Client *botsgo.Client
	Limit  int
	mutex  sync.Mutex
	time   time.Time
	rps    int
}

func NewVKBot() *VKBot {
	client, err := botsgo.NewClient("https://api.vk.com/method")
	if err != nil {
		panic(err)
	}

	// Disable Logger
	client.Logger = zap.NewNop()

	return &VKBot{
		Client: client,
		Limit:  20,
	}
}

type User struct {
	UserID   int      `json:"user_id,omitempty"`
	UserIDs  []int    `json:"user_ids,omitempty"`
	NameCase string   `json:"name_case,omitempty"`
	Fields   []string `json:"fields,omitempty"`
}

type UsersResponse struct {
	Error *Error       `json:"error"`
	User  UserResponse `json:"response"`
}

type UserResponse []struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"first_name,omitempty"`
	Surname string `json:"last_name,omitempty"`
}

type Error struct {
	Code    int          `json:"error_code,omitempty"`
	Message string       `json:"error_msg,omitempty"`
	Params  []ErrorParam `json:"request_params,omitempty"`
}

type ErrorParam struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (bot *VKBot) GetUser(user User) UsersResponse {
	response := UsersResponse{}

	path := "users.get.msgpack"
	path += fmt.Sprintf("?access_token=%s&v=5.154&user_id=1&random_id=1", os.Getenv("TOKEN"))

	body, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}

	bot.mutex.Lock()

	sleepTime := time.Second - time.Since(bot.time)
	if sleepTime < 0 {
		bot.time = time.Now()
		bot.rps = 0
	} else if bot.rps == bot.Limit {
		time.Sleep(sleepTime)
		bot.time = time.Now()
		bot.rps = 0
	}

	bot.rps++
	bot.mutex.Unlock()

	if err := bot.Client.Call(botsgo.Call{
		Method:   http.MethodPost,
		Path:     path,
		Body:     body,
		Response: &response,
	}); err != nil {
		panic(err)
	}

	return response
}
