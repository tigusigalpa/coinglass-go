package coinglass

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// apiResponse is the envelope returned by every Coinglass API v4 endpoint.
type apiResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// get performs an HTTP GET request against the Coinglass API and decodes
// the JSON data payload into out (if non-nil). It handles authentication,
// query string construction, envelope decoding, error mapping, and
// automatic retry with exponential backoff for HTTP 429 responses.
//
// path must be an absolute path beginning with "/" relative to the
// client's configured base URL, e.g. "/api/futures/openInterest/ohlc-history".
func (c *Client) get(ctx context.Context, path string, query url.Values, out interface{}) error {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	maxAttempts := c.maxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return fmt.Errorf("coinglass: failed to build request: %w", err)
		}
		req.Header.Set("CG-API-KEY", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = fmt.Errorf("coinglass: request failed: %w", err)
			return lastErr
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("coinglass: failed to read response body: %w", readErr)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return c.decodeEnvelope(respBody, out)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxAttempts {
			delay := retryDelay(resp.Header.Get("Retry-After"), c.baseDelay, attempt)
			lastErr = newAPIError(resp.StatusCode, "", extractMessage(respBody), respBody)

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}

		return newAPIError(resp.StatusCode, "", extractMessage(respBody), respBody)
	}

	if lastErr != nil {
		return lastErr
	}
	return newAPIError(http.StatusTooManyRequests, "", "rate limit exceeded", nil)
}

// decodeEnvelope parses the Coinglass response envelope and unmarshals
// the data payload into out when out is non-nil.
func (c *Client) decodeEnvelope(body []byte, out interface{}) error {
	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("coinglass: failed to decode response envelope: %w", err)
	}

	if envelope.Code != "" && envelope.Code != "0" {
		return &APIError{
			StatusCode: 200,
			Code:       envelope.Code,
			Message:    envelope.Msg,
			RawBody:    body,
		}
	}

	if out != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("coinglass: failed to decode response data: %w", err)
		}
	}
	return nil
}

// extractMessage attempts to pull a human-readable message field out of a
// JSON error response body. If the body cannot be parsed, a generic
// message including a truncated body snippet is returned.
func extractMessage(body []byte) string {
	var parsed struct {
		Message string `json:"message"`
		Msg     string `json:"msg"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if parsed.Message != "" {
			return parsed.Message
		}
		if parsed.Msg != "" {
			return parsed.Msg
		}
		if parsed.Error != "" {
			return parsed.Error
		}
	}
	if len(body) == 0 {
		return "unknown error"
	}
	snippet := string(body)
	if len(snippet) > 200 {
		snippet = snippet[:200]
	}
	return snippet
}

// retryDelay computes the delay to wait before the next retry attempt.
// It honors a Retry-After header (in seconds) if present and valid,
// otherwise it falls back to exponential backoff based on baseDelay and
// the current attempt number.
func retryDelay(retryAfterHeader string, baseDelay time.Duration, attempt int) time.Duration {
	if retryAfterHeader != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(retryAfterHeader)); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	if baseDelay <= 0 {
		baseDelay = DefaultBaseDelay
	}
	multiplier := 1 << (attempt - 1)
	return baseDelay * time.Duration(multiplier)
}

// buildQuery converts a struct value into url.Values using struct tags
// of the form `url:"name,omitempty"`. Supported field kinds:
//   - string / *string
//   - int, int64 / *int, *int64
//   - float64 / *float64
//   - bool / *bool
//   - []string / *[]string
//
// The omitempty option causes zero or nil values to be skipped.
func buildQuery(v interface{}) url.Values {
	values := url.Values{}
	if v == nil {
		return values
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return values
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		tag, ok := field.Tag.Lookup("url")
		if !ok {
			continue
		}

		name, opts := parseTag(tag)
		if name == "-" {
			continue
		}

		omitempty := strings.Contains(opts, "omitempty")
		if str, ok := stringValue(fv, omitempty); ok {
			values.Set(name, str)
		}
	}

	return values
}

// parseTag splits a struct tag into its key name and options.
func parseTag(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx >= 0 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

// stringValue returns the string representation of a supported field value.
// The second return value is false when the value should be omitted.
func stringValue(v reflect.Value, omitempty bool) (string, bool) {
	switch v.Kind() {
	case reflect.String:
		if omitempty && v.String() == "" {
			return "", false
		}
		return v.String(), true
	case reflect.Int, reflect.Int64:
		if omitempty && v.Int() == 0 {
			return "", false
		}
		return strconv.FormatInt(v.Int(), 10), true
	case reflect.Float64:
		if omitempty && v.Float() == 0 {
			return "", false
		}
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), true
	case reflect.Bool:
		if omitempty && !v.Bool() {
			return "", false
		}
		return strconv.FormatBool(v.Bool()), true
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.String {
			if omitempty && v.Len() == 0 {
				return "", false
			}
			parts := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				parts[i] = v.Index(i).String()
			}
			return strings.Join(parts, ","), true
		}
	case reflect.Ptr:
		if v.IsNil() {
			return "", false
		}
		return stringValue(v.Elem(), false)
	}
	return "", false
}
