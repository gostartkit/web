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

// Get sends an HTTP GET request using http.DefaultClient.
//
// accessToken, when non-empty, is sent as a Bearer token. v receives the decoded
// response body; pass *RawBody to copy raw bytes instead of JSON decoding.
func Get(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(nil, ctx, http.MethodGet, url, accessToken, nil, v, before...)
}

// GetWithClient sends an HTTP GET request using client.
//
// When client is nil, http.DefaultClient is used. before callbacks can customize
// the request before it is sent.
func GetWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(client, ctx, http.MethodGet, url, accessToken, nil, v, before...)
}

// Post sends an HTTP POST request with data encoded as JSON.
//
// The request body is encoded with json.Encoder and defaults to JSON
// Content-Type/Accept headers unless before callbacks are supplied.
func Post(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(nil, ctx, http.MethodPost, url, accessToken, data, v, before...)
}

// PostWithClient sends an HTTP POST request with a JSON body using client.
//
// Use this when you need custom timeouts, transports, or connection pooling.
func PostWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(client, ctx, http.MethodPost, url, accessToken, data, v, before...)
}

// PostBytes sends an HTTP POST request with a pre-encoded byte body.
//
// This avoids JSON encoding when the payload is already serialized.
func PostBytes(ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, http.MethodPost, url, accessToken, body, v, before...)
}

// PostBytesWithClient sends an HTTP POST request with a byte body using client.
func PostBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(client, ctx, http.MethodPost, url, accessToken, body, v, before...)
}

// Put sends an HTTP PUT request with data encoded as JSON.
func Put(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(nil, ctx, http.MethodPut, url, accessToken, data, v, before...)
}

// PutWithClient sends an HTTP PUT request with a JSON body using client.
func PutWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(client, ctx, http.MethodPut, url, accessToken, data, v, before...)
}

// PutBytes sends an HTTP PUT request with a pre-encoded byte body.
func PutBytes(ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, http.MethodPut, url, accessToken, body, v, before...)
}

// PutBytesWithClient sends an HTTP PUT request with a byte body using client.
func PutBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(client, ctx, http.MethodPut, url, accessToken, body, v, before...)
}

// Patch sends an HTTP PATCH request with data encoded as JSON.
func Patch(ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(nil, ctx, http.MethodPatch, url, accessToken, data, v, before...)
}

// PatchWithClient sends an HTTP PATCH request with a JSON body using client.
func PatchWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, before ...func(r *http.Request)) error {
	return doWithJSONBody(client, ctx, http.MethodPatch, url, accessToken, data, v, before...)
}

// PatchBytes sends an HTTP PATCH request with a pre-encoded byte body.
func PatchBytes(ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, http.MethodPatch, url, accessToken, body, v, before...)
}

// PatchBytesWithClient sends an HTTP PATCH request with a byte body using client.
func PatchBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(client, ctx, http.MethodPatch, url, accessToken, body, v, before...)
}

// Delete sends an HTTP DELETE request using http.DefaultClient.
func Delete(ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(nil, ctx, http.MethodDelete, url, accessToken, nil, v, before...)
}

// DeleteWithClient sends an HTTP DELETE request using client.
func DeleteWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, before ...func(r *http.Request)) error {
	return DoWithClient(client, ctx, http.MethodDelete, url, accessToken, nil, v, before...)
}

// Do sends an HTTP request with the supplied method and body.
//
// It creates a request with ctx, applies default JSON headers when no before
// callbacks are supplied, and decodes the response into v.
func Do(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, before ...func(r *http.Request)) error {
	return DoWithClient(nil, ctx, method, url, accessToken, body, v, before...)
}

// DoWithClient sends an HTTP request with the supplied method and body using client.
//
// before callbacks can modify headers or other request fields. When at least one
// callback is provided, default Content-Type and Accept headers are not applied.
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

// DoBytes sends a request with a pre-encoded byte body.
//
// The default Content-Type is application/octet-stream and the default Accept is
// application/json unless before callbacks are supplied.
func DoBytes(ctx context.Context, method string, url string, accessToken string, body []byte, v any, before ...func(r *http.Request)) error {
	return DoBytesWithClient(nil, ctx, method, url, accessToken, body, v, before...)
}

// DoBytesWithClient sends a request with a pre-encoded byte body using client.
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

// DoReq executes an already constructed request using http.DefaultClient.
//
// Successful JSON responses are decoded into v. Pass *RawBody to v to receive
// raw bytes. failure handles HTTP 400 responses when provided.
func DoReq(req *http.Request, v any, failure func(statusCode int, body io.ReadCloser) error) error {
	return DoReqWithClient(nil, req, v, failure)
}

// DoReqWithClient executes an already constructed request using client.
//
// Status 200, 201, and 202 are treated as success and decoded into v. Status
// 204 succeeds without decoding. Common API errors are mapped to package errors;
// unexpected statuses return ErrUnexpected.
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

// TryGet sends an HTTP GET request with retry support using http.DefaultClient.
//
// retry is the maximum number of attempts; values <= 0 still perform one
// attempt. Retries stop early for non-retriable API errors such as
// ErrUnauthorized, ErrForbidden, and ErrBadRequest.
func TryGet(ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return TryGetWithClient(nil, ctx, url, accessToken, v, retry, before...)
}

// TryGetWithClient sends an HTTP GET request with retry support using client.
//
// When client is nil, http.DefaultClient is used. The context controls both the
// request and the one-second delay between attempts.
func TryGetWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return GetWithClient(client, ctx, url, accessToken, v, before...)
	})
}

// TryPost sends a JSON POST request with retry support using http.DefaultClient.
//
// data is JSON-encoded for each attempt. Use TryPostBytes when the request body
// is already encoded or when avoiding repeated JSON encoding matters.
func TryPost(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return TryPostWithClient(nil, ctx, url, accessToken, data, v, retry, before...)
}

// TryPostWithClient sends a JSON POST request with retry support using client.
//
// The response is decoded into v using the same rules as PostWithClient. Pass
// *RawBody to v when the response should be copied as bytes instead of decoded.
func TryPostWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PostWithClient(client, ctx, url, accessToken, data, v, before...)
	})
}

// TryPostBytes sends a POST request with a pre-encoded byte body and retries.
//
// The same byte slice is reused for every attempt, making this the preferred
// retry helper when the payload is already serialized.
func TryPostBytes(ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryPostBytesWithClient(nil, ctx, url, accessToken, body, v, retry, before...)
}

// TryPostBytesWithClient sends a byte-body POST request with retry support using client.
//
// Default headers are application/octet-stream for Content-Type and
// application/json for Accept unless before callbacks are supplied.
func TryPostBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PostBytesWithClient(client, ctx, url, accessToken, body, v, before...)
	})
}

// TryPut sends a JSON PUT request with retry support using http.DefaultClient.
//
// retry counts total attempts, not extra attempts. For example, retry=3 performs
// at most three HTTP requests.
func TryPut(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return TryPutWithClient(nil, ctx, url, accessToken, data, v, retry, before...)
}

// TryPutWithClient sends a JSON PUT request with retry support using client.
//
// Use this variant when the caller owns a tuned http.Client with timeouts,
// tracing, custom transport, or connection pooling.
func TryPutWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PutWithClient(client, ctx, url, accessToken, data, v, before...)
	})
}

// TryPutBytes sends a PUT request with a pre-encoded byte body and retries.
//
// This avoids JSON encoding and is useful for binary, compressed, or already
// serialized payloads.
func TryPutBytes(ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryPutBytesWithClient(nil, ctx, url, accessToken, body, v, retry, before...)
}

// TryPutBytesWithClient sends a byte-body PUT request with retry support using client.
//
// The response handling and error mapping are identical to DoReqWithClient.
func TryPutBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PutBytesWithClient(client, ctx, url, accessToken, body, v, before...)
	})
}

// TryPatch sends a JSON PATCH request with retry support using http.DefaultClient.
//
// before callbacks can customize headers or request metadata before each
// attempt is sent.
func TryPatch(ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return TryPatchWithClient(nil, ctx, url, accessToken, data, v, retry, before...)
}

// TryPatchWithClient sends a JSON PATCH request with retry support using client.
//
// Non-retriable framework errors are returned immediately so authentication or
// validation failures are not repeated unnecessarily.
func TryPatchWithClient(client *http.Client, ctx context.Context, url string, accessToken string, data any, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PatchWithClient(client, ctx, url, accessToken, data, v, before...)
	})
}

// TryPatchBytes sends a PATCH request with a pre-encoded byte body and retries.
//
// Use this for JSON bytes, merge-patch payloads, binary patches, or any payload
// that has already been serialized by the caller.
func TryPatchBytes(ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryPatchBytesWithClient(nil, ctx, url, accessToken, body, v, retry, before...)
}

// TryPatchBytesWithClient sends a byte-body PATCH request with retry support using client.
//
// Each attempt creates a fresh reader over body, so the payload can be replayed
// safely across retries.
func TryPatchBytesWithClient(client *http.Client, ctx context.Context, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return PatchBytesWithClient(client, ctx, url, accessToken, body, v, before...)
	})
}

// TryDelete sends an HTTP DELETE request with retry support using http.DefaultClient.
//
// v receives the successful response body, or can be nil when the caller only
// cares whether the delete operation succeeded.
func TryDelete(ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return TryDeleteWithClient(nil, ctx, url, accessToken, v, retry, before...)
}

// TryDeleteWithClient sends an HTTP DELETE request with retry support using client.
//
// It is a convenience wrapper around retryLoop and DeleteWithClient.
func TryDeleteWithClient(client *http.Client, ctx context.Context, url string, accessToken string, v any, retry int, before ...func(r *http.Request)) error {
	return retryLoop(ctx, retry, func() error {
		return DeleteWithClient(client, ctx, url, accessToken, v, before...)
	})
}

// TryDo sends an arbitrary HTTP request with retry support using http.DefaultClient.
//
// The body reader, when non-nil, is read into memory once so it can be replayed
// for each retry attempt. Prefer TryDoBytes when the caller already has a byte
// slice or when making this buffering behavior explicit is clearer.
func TryDo(ctx context.Context, method string, url string, accessToken string, body io.Reader, v any, retry int, before ...func(r *http.Request)) error {
	return TryDoWithClient(nil, ctx, method, url, accessToken, body, v, retry, before...)
}

// TryDoWithClient sends an arbitrary HTTP request with retry support using client.
//
// The request body is buffered once before the first attempt. The context is used
// for all attempts and for cancellation while waiting between retries.
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

// TryDoBytes sends an arbitrary HTTP request with a byte body and retry support.
//
// It uses http.DefaultClient and reuses body across attempts without additional
// buffering.
func TryDoBytes(ctx context.Context, method string, url string, accessToken string, body []byte, v any, retry int, before ...func(r *http.Request)) error {
	return TryDoBytesWithClient(nil, ctx, method, url, accessToken, body, v, retry, before...)
}

// TryDoBytesWithClient sends an arbitrary byte-body HTTP request with retry support.
//
// Use this as the lowest-level retry helper when you need a custom method,
// custom client, optional bearer token, and a replayable byte payload.
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
		buf := bytes.NewBuffer((*out)[:cap(*out)])
		buf.Reset()
		_, err := buf.ReadFrom(body)
		if err == nil {
			*out = RawBody(buf.Bytes())
		}
		return err
	default:
		return decodeJSONBody(body, v)
	}
}
