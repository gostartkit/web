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

// Get http get
func Get(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return Do(ctx, http.MethodGet, url, accessToken, nil, v, before...)
}

// Post http post
func Post(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(ctx, http.MethodPost, url, accessToken, data, v, before...)
}

// Put http put
func Put(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(ctx, http.MethodPut, url, accessToken, data, v, before...)
}

// Patch http patch
func Patch(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(ctx, http.MethodPatch, url, accessToken, data, v, before...)
}

// Delete http delete
func Delete(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return Do(ctx, http.MethodDelete, url, accessToken, nil, v, before...)
}

// Do do http request
func Do(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, before ...func(r *http.Request)) error {

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

	return DoReq(req, v, nil)
}

// DoReq do http request
func DoReq(req *http.Request, v any, failure func(statusCode int, body io.ReadCloser) error) error {

	resp, err := http.DefaultClient.Do(req)

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
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
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
		if err := json.NewDecoder(resp.Body).Decode(&errMessage); err != nil {
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
	return retryLoop(ctx, retry, func() error {
		return Get(ctx, url, accessToken, v, before...)
	})
}

// TryPost
func TryPost(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return Post(ctx, url, accessToken, data, v, before...)
	})
}

// TryPut
func TryPut(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return Put(ctx, url, accessToken, data, v, before...)
	})
}

// TryPatch
func TryPatch(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return Patch(ctx, url, accessToken, data, v, before...)
	})
}

// TryDelete
func TryDelete(ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return Delete(ctx, url, accessToken, v, before...)
	})
}

// TryDo
func TryDo(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, retry int, before ...func(r *http.Request)) error {
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
		return Do(ctx, method, url, accessToken, reqBody, v, before...)
	})
}

func retryLoop(ctx context.Context, retry int, fn func() error) error {
	attempts := retry
	if attempts <= 0 {
		attempts = 1
	}

	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if isNonRetriable(err) || i == attempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return err
}

func isNonRetriable(err error) bool {
	return err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest)
}

func doWithJSONBody(ctx context.Context, method string, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	body := _bodyBufferPool.Get().(*bytes.Buffer)
	body.Reset()

	err := json.NewEncoder(body).Encode(data)
	if err == nil {
		err = Do(ctx, method, url, accessToken, bytes.NewReader(body.Bytes()), v, before...)
	}

	body.Reset()
	_bodyBufferPool.Put(body)
	return err
}
