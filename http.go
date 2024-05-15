package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Get http get
func Get(url string, accessToken string, v any) error {
	return Do(http.MethodGet, url, accessToken, nil, v, nil, nil)
}

// Post http post
func Post(url string, accessToken string, data any, v any) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(http.MethodPost, url, accessToken, body, v, nil, nil)
}

// Put http put
func Put(url string, accessToken string, data any, v any) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(http.MethodPut, url, accessToken, body, v, nil, nil)
}

// Patch http patch
func Patch(url string, accessToken string, data any, v any) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return Do(http.MethodPatch, url, accessToken, body, v, nil, nil)
}

// Delete http delete
func Delete(url string, accessToken string, v any) error {
	return Do(http.MethodDelete, url, accessToken, nil, v, nil, nil)
}

// Do do http request
func Do(method string, url string, accessToken string, body io.Reader, v any, cb func(r *http.Request), badRequest func(body io.ReadCloser) error) error {

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return err
	}

	if cb == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	} else {
		cb(req)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return err
		}
		return nil
	case http.StatusBadRequest:
		if badRequest != nil {
			return badRequest(resp.Body)
		}

		errMessage := ""
		if err := json.NewDecoder(resp.Body).Decode(&errMessage); err != nil {
			return err
		}
		return errors.New(errMessage)
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	default:
		return ErrUnExpected
	}
}
