package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Get do http get
func Get(client *http.Client, url string, accessToken string, v interface{}) error {
	return httpDo(client, http.MethodGet, url, accessToken, nil, v)
}

// Post do http post
func Post(client *http.Client, url string, accessToken string, data interface{}, v interface{}) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpDo(client, http.MethodPost, url, accessToken, body, v)
}

// Put do http put
func Put(client *http.Client, url string, accessToken string, data interface{}, v interface{}) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpDo(client, http.MethodPut, url, accessToken, body, v)
}

// Patch do http patch
func Patch(client *http.Client, url string, accessToken string, data interface{}, v interface{}) error {

	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(data)

	if err != nil {
		return err
	}

	return httpDo(client, http.MethodPatch, url, accessToken, body, v)
}

// Delete do http delete
func Delete(client *http.Client, url string, accessToken string, v interface{}) error {
	return httpDo(client, http.MethodDelete, url, accessToken, nil, v)
}

// httpDo do http request
func httpDo(client *http.Client, method string, url string, accessToken string, body io.Reader, v interface{}) error {

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := client.Do(req)

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
