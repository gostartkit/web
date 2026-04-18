package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var _bodyBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func httpClientOrDefault(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return http.DefaultClient
}

// Get http get
func Get(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(nil, ctx, http.MethodGet, url, accessToken, nil, v, before...)
}

// GetWithClient http get with explicit client
func GetWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(client, ctx, http.MethodGet, url, accessToken, nil, v, before...)
}

// Post http post
func Post(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(nil, ctx, http.MethodPost, url, accessToken, data, v, before...)
}

// PostWithClient http post with explicit client
func PostWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(client, ctx, http.MethodPost, url, accessToken, data, v, before...)
}

// PostBytes http post with raw bytes body
func PostBytes(ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, http.MethodPost, url, accessToken, body, v, before...)
}

// PostBytesWithClient http post with raw bytes body and explicit client
func PostBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(client, ctx, http.MethodPost, url, accessToken, body, v, before...)
}

// Put http put
func Put(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(nil, ctx, http.MethodPut, url, accessToken, data, v, before...)
}

// PutWithClient http put with explicit client
func PutWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(client, ctx, http.MethodPut, url, accessToken, data, v, before...)
}

// PutBytes http put with raw bytes body
func PutBytes(ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, http.MethodPut, url, accessToken, body, v, before...)
}

// PutBytesWithClient http put with raw bytes body and explicit client
func PutBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(client, ctx, http.MethodPut, url, accessToken, body, v, before...)
}

// Patch http patch
func Patch(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(nil, ctx, http.MethodPatch, url, accessToken, data, v, before...)
}

// PatchWithClient http patch with explicit client
func PatchWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(client, ctx, http.MethodPatch, url, accessToken, data, v, before...)
}

// PatchBytes http patch with raw bytes body
func PatchBytes(ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, http.MethodPatch, url, accessToken, body, v, before...)
}

// PatchBytesWithClient http patch with raw bytes body and explicit client
func PatchBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(client, ctx, http.MethodPatch, url, accessToken, body, v, before...)
}

// Delete http delete
func Delete(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(nil, ctx, http.MethodDelete, url, accessToken, nil, v, before...)
}

// DeleteWithClient http delete with explicit client
func DeleteWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(client, ctx, http.MethodDelete, url, accessToken, nil, v, before...)
}

// Do do http request
func Do(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, before ...func(r *http.Request)) error {
	return DoWithClient(nil, ctx, method, url, accessToken, body, v, before...)
}

// DoWithClient do http request with explicit client
func DoWithClient(client *http.Client, ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, before ...func(r *http.Request)) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	if err != nil {
		return err
	}

	if len(before) == 0 {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	} else {
		for _, fn := range before {
			if fn != nil {
				fn(req)
			}
		}
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	return DoReqWithClient(client, req, v, nil)
}

// DoBytes sends a request with a pre-encoded body.
func DoBytes(ctx context.Context, method string, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, method, url, accessToken, body, v, before...)
}

// DoBytesWithClient sends a request with a pre-encoded body and explicit client.
func DoBytesWithClient(client *http.Client, ctx context.Context, method string, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}

	if len(before) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Accept", "application/json")
	} else {
		for _, fn := range before {
			if fn != nil {
				fn(req)
			}
		}
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	return DoReqWithClient(client, req, v, nil)
}

// DoReq do http request
func DoReq(req *http.Request, v any, failure func(statusCode int, body io.ReadCloser) error) error {
	return DoReqWithClient(nil, req, v, failure)
}

// DoReqWithClient do http request with explicit client
func DoReqWithClient(client *http.Client, req *http.Request, v any, failure func(statusCode int, body io.ReadCloser) error) error {
	resp, err := httpClientOrDefault(client).Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		if v == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			return nil
		}
		if err := decodeResponseBody(resp.Body, v); err != nil {
			return err
		}
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusBadRequest:
		if failure != nil {
			return failure(resp.StatusCode, resp.Body)
		}
		errMessage := ""
		if err := decodeJSONBody(resp.Body, &errMessage); err != nil {
			return fmt.Errorf("%w: %s", ErrBadRequest, err)
		}
		return fmt.Errorf("%w: %s", ErrBadRequest, errMessage)
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return ErrUnexpected
	}
}

// TryGet
func TryGet(ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return TryGetWithClient(nil, ctx, url, accessToken, v, retry, before...)
}

// TryGetWithClient
func TryGetWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return GetWithClient(client, ctx, url, accessToken, v, before...)
	})
}

// TryPost
func TryPost(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return TryPostWithClient(nil, ctx, url, accessToken, data, v, retry, before...)
}

// TryPostWithClient
func TryPostWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PostWithClient(client, ctx, url, accessToken, data, v, before...)
	})
}

// TryPostBytes
func TryPostBytes(ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryPostBytesWithClient(nil, ctx, url, accessToken, body, v, retry, before...)
}

// TryPostBytesWithClient
func TryPostBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PostBytesWithClient(client, ctx, url, accessToken, body, v, before...)
	})
}

// TryPut
func TryPut(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return TryPutWithClient(nil, ctx, url, accessToken, data, v, retry, before...)
}

// TryPutWithClient
func TryPutWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PutWithClient(client, ctx, url, accessToken, data, v, before...)
	})
}

// TryPutBytes
func TryPutBytes(ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryPutBytesWithClient(nil, ctx, url, accessToken, body, v, retry, before...)
}

// TryPutBytesWithClient
func TryPutBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PutBytesWithClient(client, ctx, url, accessToken, body, v, before...)
	})
}

// TryPatch
func TryPatch(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return TryPatchWithClient(nil, ctx, url, accessToken, data, v, retry, before...)
}

// TryPatchWithClient
func TryPatchWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PatchWithClient(client, ctx, url, accessToken, data, v, before...)
	})
}

// TryPatchBytes
func TryPatchBytes(ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryPatchBytesWithClient(nil, ctx, url, accessToken, body, v, retry, before...)
}

// TryPatchBytesWithClient
func TryPatchBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PatchBytesWithClient(client, ctx, url, accessToken, body, v, before...)
	})
}

// TryDelete
func TryDelete(ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return TryDeleteWithClient(nil, ctx, url, accessToken, v, retry, before...)
}

// TryDeleteWithClient
func TryDeleteWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return DeleteWithClient(client, ctx, url, accessToken, v, before...)
	})
}

// TryDo
func TryDo(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, retry int, before ...func(r *http.Request)) error {
	return TryDoWithClient(nil, ctx, method, url, accessToken, body, v, retry, before...)
}

// TryDoWithClient
func TryDoWithClient(client *http.Client, ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, retry int, before ...func(r *http.Request)) error {
	var payload []byte
	var err error

	if body != nil {
		payload, err = io.ReadAll(body)
		if err != nil {
			return err
		}
	}

	return retryLoop(ctx, retry, func() error {
		var reqBody io.Reader
		if payload != nil {
			reqBody = bytes.NewReader(payload)
		}
		return DoWithClient(client, ctx, method, url, accessToken, reqBody, v, before...)
	})
}

// TryDoBytes
func TryDoBytes(ctx context.Context, method string, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryDoBytesWithClient(nil, ctx, method, url, accessToken, body, v, retry, before...)
}

// TryDoBytesWithClient
func TryDoBytesWithClient(client *http.Client, ctx context.Context, method string, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return DoBytesWithClient(client, ctx, method, url, accessToken, body, v, before...)
	})
}

func retryLoop(ctx context.Context, retry int, fn func() error) error {
	attempts := retry
	if attempts <= 0 {
		attempts = 1
	}

	var err error
	var timer *time.Timer
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			if timer != nil {
				timer.Stop()
			}
			return nil
		}

		if isNonRetriable(err) || i == attempts-1 {
			break
		}

		if timer == nil {
			timer = time.NewTimer(time.Second)
		} else {
			timer.Reset(time.Second)
		}

		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return ctx.Err()
		case <-timer.C:
		}
	}

	if timer != nil {
		timer.Stop()
	}
	return err
}

func isNonRetriable(err error) bool {
	return err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest)
}

func doWithJSONBody(client *http.Client, ctx context.Context, method string, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	body := _bodyBufferPool.Get().(*bytes.Buffer)
	body.Reset()

	err := json.NewEncoder(body).Encode(data)
	if err == nil {
		err = DoWithClient(client, ctx, method, url, accessToken, bytes.NewReader(body.Bytes()), v, before...)
	}

	body.Reset()
	_bodyBufferPool.Put(body)
	return err
}

func decodeJSONBody(body io.ReadCloser, v any) error {
	buf := _bodyReadBufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	_, err := buf.ReadFrom(body)
	if err == nil {
		err = json.Unmarshal(buf.Bytes(), v)
	}

	buf.Reset()
	_bodyReadBufferPool.Put(buf)
	return err
}

func decodeResponseBody(body io.ReadCloser, v any) error {
	switch out := v.(type) {
	case *RawBody:
		buf := _bodyReadBufferPool.Get().(*bytes.Buffer)
		buf.Reset()

		_, err := buf.ReadFrom(body)
		if err == nil {
			*out = append((*out)[:0], buf.Bytes()...)
		}

		buf.Reset()
		_bodyReadBufferPool.Put(buf)
		return err
	default:
		return decodeJSONBody(body, v)
	}
}
