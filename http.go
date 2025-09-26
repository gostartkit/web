package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Get http get
func Get(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return Do(ctx, http.MethodGet, url, accessToken, nil, v, before...)
}

// Post http post
func Post(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(ctx, http.MethodPost, url, accessToken, body, v, before...)
}

// Put http put
func Put(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(ctx, http.MethodPut, url, accessToken, body, v, before...)
}

// Patch http patch
func Patch(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(ctx, http.MethodPatch, url, accessToken, body, v, before...)
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

	if before == nil {
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

	var err error

	for i := retry; i > 0; i-- {

		if err = Get(ctx, url, accessToken, v, before...); err == nil {
			break
		}

		if err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest) {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

// TryPost
func TryPost(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Post(ctx, url, accessToken, data, v, before...); err == nil {
			break
		}

		if err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest) {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

// TryPut
func TryPut(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Put(ctx, url, accessToken, data, v, before...); err == nil {
			break
		}

		if err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest) {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

// TryPatch
func TryPatch(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Patch(ctx, url, accessToken, data, v, before...); err == nil {
			break
		}

		if err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest) {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

// TryDelete
func TryDelete(ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Delete(ctx, url, accessToken, v, before...); err == nil {
			break
		}

		if err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest) {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

// TryDo
func TryDo(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Do(ctx, method, url, accessToken, body, v, before...); err == nil {
			break
		}

		if err == ErrUnauthorized || err == ErrForbidden || errors.Is(err, ErrBadRequest) {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}
