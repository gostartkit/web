package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Get http get
func Get(url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return Do(http.MethodGet, url, accessToken, nil, v, before...)
}

// Post http post
func Post(url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(http.MethodPost, url, accessToken, body, v, before...)
}

// Put http put
func Put(url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(http.MethodPut, url, accessToken, body, v, before...)
}

// Patch http patch
func Patch(url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(http.MethodPatch, url, accessToken, body, v, before...)
}

// Delete http delete
func Delete(url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return Do(http.MethodDelete, url, accessToken, nil, v, before...)
}

// Do do http request
func Do(method string, url string, accessToken string, body io.Reader, v any, before ...func(r *http.Request)) error {

	req, err := http.NewRequest(method, url, body)

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
		return ErrUnExpected
	}
}

// TryGet
func TryGet(url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Get(url, accessToken, v, before...); err == nil {
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
func TryPost(url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Post(url, accessToken, data, v, before...); err == nil {
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
func TryPut(url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Put(url, accessToken, data, v, before...); err == nil {
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
func TryPatch(url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Patch(url, accessToken, data, v, before...); err == nil {
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
func TryDelete(url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Delete(url, accessToken, v, before...); err == nil {
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
func TryDo(method string, url string, accessToken string, body io.Reader, v any, retry int, before ...func(r *http.Request)) error {

	var err error

	for i := retry; i > 0; i-- {

		if err = Do(method, url, accessToken, body, v, before...); err == nil {
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
