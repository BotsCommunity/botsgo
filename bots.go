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

type Client struct {
	Client *http.Client
	API    string
	Token  string
	Logger *zap.Logger
}

type Call struct {
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

func NewClient(apiURL string) (*Client, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: http.DefaultClient,
		API:    apiURL,
		Logger: logger,
	}, err
}

func (c *Client) Call(options Call) error {
	var (
		url          = c.API + "/" + options.Path
		request, err = http.NewRequestWithContext(context.Background(), options.Method, url, bytes.NewReader(options.Body))
	)

	if err != nil {
		c.Logger.Error(fmt.Sprint("NewRequestWithContext error ", err.Error()))
		return err
	}

	if err := c.multiPart(request, options.File); err != nil {
		return err
	}

	response, err := c.Client.Do(request)
	if err != nil {
		c.Logger.Error(fmt.Sprint("Do from Client error ", err.Error()))
		return err
	}

	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		c.Logger.Error(fmt.Sprint("ParseMediaType from mime error ", err.Error()))
		return err
	}

	switch mediaType {
	case "application/x-msgpack":
		decoder := msgpack.NewDecoder(response.Body)
		decoder.SetCustomStructTag("json")

		if decodeErr := decoder.Decode(&options.Response); decodeErr != nil {
			c.Logger.Error(fmt.Sprint("Decode from MessagePack error ", decodeErr.Error()))

			if closeErr := response.Body.Close(); closeErr != nil {
				c.Logger.Error(fmt.Sprint("CloseBody from HTP response error ", closeErr.Error()))
				return closeErr
			}
		}
	case "application/json":
		if decodeErr := json.NewDecoder(response.Body).Decode(&options.Response); decodeErr != nil {
			c.Logger.Error(fmt.Sprint("Decode from JSON error ", decodeErr.Error()))

			if closeErr := response.Body.Close(); closeErr != nil {
				c.Logger.Error(fmt.Sprint("CloseBody from HTTP response error  ", closeErr.Error()))
				return closeErr
			}
		}
	}

	c.Logger.Info(fmt.Sprintf("Response from request %s, Body: %s, Response: %+v", url, string(options.Body), options.Response))

	return err
}

func (c *Client) multiPart(request *http.Request, file File) error {
	if (File{}) != file {
		var (
			buffer    = &bytes.Buffer{}
			newWriter = multipart.NewWriter(buffer)
		)

		writer, err := newWriter.CreateFormFile(file.Field, file.Data.Name())
		if err != nil {
			c.Logger.Error(fmt.Sprint("CreateFormFile from multipart error ", err.Error()))
			return err
		}

		_, err = io.Copy(writer, file.Data)
		if err != nil {
			c.Logger.Error(fmt.Sprint("Copy from io error ", err.Error()))
			return err
		}

		if err := newWriter.Close(); err != nil {
			c.Logger.Error(fmt.Sprint("Close from multipart error ", err.Error()))
			return err
		}

		request.Body = io.NopCloser(buffer)
		request.Method = http.MethodPost
	}

	return nil
}
