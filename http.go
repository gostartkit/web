package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	_httpClient *http.Client
)

// SetHttpClient set http client
func SetHttpClient(client *http.Client) {
	_httpClient = client
}

// Get http get
func Get(url string, accessToken string, v any) error {
	return httpRequest(http.MethodGet, url, accessToken, nil, v)
}

// Post http post
func Post(url string, accessToken string, data any, v any) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpRequest(http.MethodPost, url, accessToken, body, v)
}

// Put http put
func Put(url string, accessToken string, data any, v any) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpRequest(http.MethodPut, url, accessToken, body, v)
}

// Patch http patch
func Patch(url string, accessToken string, data any, v any) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpRequest(http.MethodPatch, url, accessToken, body, v)
}

// Delete http delete
func Delete(url string, accessToken string, v any) error {
	return httpRequest(http.MethodDelete, url, accessToken, nil, v)
}

// httpRequest http request
func httpRequest(method string, url string, accessToken string, body io.Reader, v any) error {

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	if _httpClient == nil {
		_httpClient = &http.Client{}
	}

	resp, err := _httpClient.Do(req)

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
		return ErrUnExpectedError
	}
}
