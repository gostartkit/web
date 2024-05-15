package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Get http get
func Get(url string, accessToken string, v any, cbs ...HttpDoCallback) error {
	return httpDo(http.MethodGet, url, accessToken, nil, v, cbs...)
}

// Post http post
func Post(url string, accessToken string, data any, v any, cbs ...HttpDoCallback) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpDo(http.MethodPost, url, accessToken, body, v, cbs...)
}

// Put http put
func Put(url string, accessToken string, data any, v any, cbs ...HttpDoCallback) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpDo(http.MethodPut, url, accessToken, body, v, cbs...)
}

// Patch http patch
func Patch(url string, accessToken string, data any, v any, cbs ...HttpDoCallback) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpDo(http.MethodPatch, url, accessToken, body, v, cbs...)
}

// Delete http delete
func Delete(url string, accessToken string, v any, cbs ...HttpDoCallback) error {
	return httpDo(http.MethodDelete, url, accessToken, nil, v, cbs...)
}

// httpDo do http request
func httpDo(method string, url string, accessToken string, body io.Reader, v any, cbs ...HttpDoCallback) error {

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return err
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	for _, cb := range cbs {
		cb(req)
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
