package botsgo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"

	"github.com/vmihailenco/msgpack/v5"
)

type Requester struct {
	client      *Client
	HTTPRequest *http.Request
	BaseURL     *url.URL
	path        string
	response    any
}

func (c *Client) NewRequest(context context.Context) (*Requester, error) {
	req, err := http.NewRequestWithContext(context, "", "", nil)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(c.APIURL)
	if err != nil {
		return nil, err
	}

	return &Requester{
		client:      c,
		HTTPRequest: req,
		BaseURL:     baseURL,
	}, nil
}

func (r *Requester) Method(method string) *Requester {
	r.HTTPRequest.Method = method
	return r
}

func (r *Requester) Path(path string) *Requester {
	r.path = path
	return r
}

func (r *Requester) Body(body []byte) *Requester {
	r.HTTPRequest.Body = io.NopCloser(bytes.NewReader(body))

	return r
}

func (r *Requester) GetHeader(key string) string {
	return r.HTTPRequest.Header.Get(key)
}

func (r *Requester) SetHeader(key, value string) *Requester {
	r.HTTPRequest.Header.Set(key, value)
	return r
}

func (r *Requester) Response(res any) *Requester {
	r.response = res
	return r
}

func (r *Requester) Do() (*http.Response, error) {
	URL, err := url.Parse(r.path)
	if err != nil {
		return nil, err
	}

	r.HTTPRequest.URL = r.BaseURL.ResolveReference(URL)

	res, err := http.DefaultClient.Do(r.HTTPRequest)
	if err != nil {
		return nil, err
	}

	contentType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if err := r.decodeContentTypes(contentType, res); err != nil {
		return nil, err
	}

	if r.client.Logger != nil {
		r.client.Logger.Info(
			fmt.Sprintf("\n[%s] %s\n[%s] %s", r.HTTPRequest.Method, r.HTTPRequest.URL.String(), contentType, res.Status),
		)
	}

	return res, nil
}

func (r *Requester) decodeContentTypes(contentType string, res *http.Response) error {
	switch contentType {
	case "application/json":
		if err := r.client.JSONIter.NewDecoder(res.Body).Decode(&r.response); err != nil {
			if closeError := res.Body.Close(); closeError != nil {
				return closeError
			}

			return err
		}
	case "application/x-msgpack", "application/msgpack":
		decoder := msgpack.NewDecoder(res.Body)
		decoder.SetCustomStructTag("json")

		if err := decoder.Decode(&r.response); err != nil {
			if closeError := res.Body.Close(); closeError != nil {
				return closeError
			}

			return err
		}
	}

	return nil
}
