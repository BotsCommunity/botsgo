package botsgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

type Bot struct {
	API        string
	HTTPClient *http.Client
	Logger     *zap.Logger
}

type CallOptions struct {
	Method   string
	Path     string
	Body     []byte
	File     File
	Response any
}

type File struct {
	Field string
	Data  *os.File
}

func NewBot(apiURL string) (*Bot, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &Bot{
		API:        apiURL,
		HTTPClient: http.DefaultClient,
		Logger:     logger,
	}, err
}

func (bot *Bot) Call(options CallOptions) error {
	var (
		url          = bot.API + "/" + options.Path
		request, err = http.NewRequestWithContext(context.Background(), options.Method, url, bytes.NewReader(options.Body))
	)

	if err != nil {
		bot.Logger.Error(fmt.Sprint("NewRequestWithContext error ", err.Error()))
		return err
	}

	if err := bot.multiPart(request, options.File); err != nil {
		return err
	}

	response, err := bot.HTTPClient.Do(request)
	if err != nil {
		bot.Logger.Error(fmt.Sprint("Do from Client error ", err.Error()))
		return err
	}

	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		bot.Logger.Error(fmt.Sprint("ParseMediaType from mime error ", err.Error()))
		return err
	}

	switch mediaType {
	case "application/x-msgpack":
		decoder := msgpack.NewDecoder(response.Body)
		decoder.SetCustomStructTag("json")

		if decodeErr := decoder.Decode(&options.Response); decodeErr != nil {
			bot.Logger.Error(fmt.Sprint("Decode from MessagePack error ", decodeErr.Error()))

			if closeErr := response.Body.Close(); closeErr != nil {
				bot.Logger.Error(fmt.Sprint("CloseBody from HTP response error ", closeErr.Error()))
				return closeErr
			}
		}
	case "application/json":
		if decodeErr := json.NewDecoder(response.Body).Decode(&options.Response); decodeErr != nil {
			bot.Logger.Error(fmt.Sprint("Decode from JSON error ", decodeErr.Error()))

			if closeErr := response.Body.Close(); closeErr != nil {
				bot.Logger.Error(fmt.Sprint("CloseBody from HTTP response error  ", closeErr.Error()))
				return closeErr
			}
		}
	}

	bot.Logger.Info(fmt.Sprintf("Response from request %s, Body: %s, Response: %+v", url, string(options.Body), options.Response))

	return err
}

func (bot *Bot) multiPart(request *http.Request, file File) error {
	if (File{}) != file {
		var (
			buffer    = &bytes.Buffer{}
			newWriter = multipart.NewWriter(buffer)
		)

		writer, err := newWriter.CreateFormFile(file.Field, file.Data.Name())
		if err != nil {
			bot.Logger.Error(fmt.Sprint("CreateFormFile from multipart error ", err.Error()))
			return err
		}

		_, err = io.Copy(writer, file.Data)
		if err != nil {
			bot.Logger.Error(fmt.Sprint("Copy from io error ", err.Error()))
			return err
		}

		if err := newWriter.Close(); err != nil {
			bot.Logger.Error(fmt.Sprint("Close from multipart error ", err.Error()))
			return err
		}

		request.Body = io.NopCloser(buffer)
		request.Method = http.MethodPost
	}

	return nil
}
