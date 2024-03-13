package botsgo

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type Client struct {
	APIURL     string
	HTTPClient *http.Client
	JSONIter   jsoniter.API
	Logger     *zap.Logger
}

func NewClient(apiURL string) (*Client, error) {
	devLogger, err := zap.NewDevelopment()

	return &Client{
		APIURL:     apiURL,
		HTTPClient: http.DefaultClient,
		JSONIter:   jsoniter.ConfigCompatibleWithStandardLibrary,
		Logger:     devLogger,
	}, err
}
