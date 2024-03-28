package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/config"
)

// Compile the regular expression once and store it in a global variable
var statusCodeRegex = regexp.MustCompile(`code=(\d+)`)

func Retry(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.ConfigFile.TezosHostRetry == "" ||
				strings.Contains(c.Request().URL.Path, "mempool") ||
				strings.Contains(c.Request().URL.Path, "monitor") {
				return next(c)
			}

			writer := newDelayedResponseWriter(c)
			delayedResponse := echo.NewResponse(&writer, c.Echo())
			c.SetResponse(delayedResponse)

			// Read the request body
			bodyBytes, _ := io.ReadAll(c.Request().Body)
			// Replace the request body with a buffer
			c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
			statusFromMsg := 0
			err = next(c)

			status := c.Response().Status
			// Extract the status code from the error message
			if err != nil {
				match := statusCodeRegex.FindStringSubmatch(err.Error())
				if len(match) > 1 {
					statusFromMsg, _ = strconv.Atoi(match[1])
				}
			}
			// Check the request method and path
			method := c.Request().Method
			path := c.Request().URL.Path
			shouldRetry := (method == http.MethodGet && (status == http.StatusNotFound || status == http.StatusForbidden)) ||
				(method == http.MethodPost && statusFromMsg == 502 &&
				strings.Contains(path, "/chains/main/blocks/head/helpers/scripts/"))

			if err != nil && statusFromMsg == 502 {
				c.Logger().Infof("Error occurred with status 502. Request method: %s, path: %s, response status: %d", method, path, statusFromMsg)
				c.Logger().Infof("Should retry: %v", shouldRetry)
			}

			if shouldRetry {
				c.Logger().Infof("Triggering retry for http status %d", status)
				writer.Reset()
				delayedResponse.Committed = false
				delayedResponse.Size = 0
				delayedResponse.Status = 0
				c.Set("retry", status)

				// This _error key comes from Echo (referenced in ProxyWithConfig middleware)
				// so we need to reset this as well.
				c.Set("_error", nil)

				// Reset the request body
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				err = next(c)
				// Log the status after the retry attempt
				c.Logger().Infof("Retry attempt http status: %d", c.Response().Status)
				// Extract the status code from the response body for the retry attempt
				statusFromMsgRetry := 0
				if err != nil {
					match := statusCodeRegex.FindStringSubmatch(err.Error())
					if len(match) > 1 {
						statusFromMsgRetry, _ = strconv.Atoi(match[1])
					}
				}
				retryFailed := statusFromMsgRetry == 502 || c.Response().Status == http.StatusNotFound || c.Response().Status == http.StatusForbidden
				if retryFailed {
					c.Logger().Infof("Retry attempt http status: %d, code found in body: %d", c.Response().Status, statusFromMsgRetry)
				} else {
					c.Logger().Infof("Retry attempt succeeded with http status: %d, code found in body: %d", c.Response().Status, statusFromMsgRetry)
				}
			}

			c.Set("retry", nil)
			if delayedWriter, ok := c.Response().Writer.(*delayedResponseWriter); ok {
				if commitErr := delayedWriter.Commit(); commitErr != nil {
					c.Logger().Error(commitErr)
				}
			}
			return err
		}
	}
}

type delayedResponseWriter struct {
	originalResponse http.ResponseWriter
	statusCode       int
	buf              *bytes.Buffer
}

// Header returns the map of header fields.
func (d *delayedResponseWriter) Header() http.Header {
	return d.originalResponse.Header()
}

// Write records the body content to send when Commit is called.
func (d *delayedResponseWriter) Write(bytes []byte) (int, error) {
	return d.buf.Write(bytes)
}

// WriteHeader records the statusCode to send when Commit is called.
func (d *delayedResponseWriter) WriteHeader(statusCode int) {
	d.statusCode = statusCode
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
// This is required for some content types such as octet-stream (for files download)
func (d *delayedResponseWriter) Flush() {
	d.originalResponse.(http.Flusher).Flush()
}

// Reset resets the internal buffer and status code.
func (d *delayedResponseWriter) Reset() {
	d.buf.Reset()
	d.statusCode = 0
}

// Commit sends the header and body content to the original response writer.
func (d *delayedResponseWriter) Commit() (err error) {
	if d.statusCode != 0 {
		d.originalResponse.WriteHeader(d.statusCode)
	}
	_, err = d.originalResponse.Write(d.buf.Bytes())
	return
}

func newDelayedResponseWriter(c echo.Context) delayedResponseWriter {
	return delayedResponseWriter{
		originalResponse: c.Response(),
		buf:              new(bytes.Buffer)}
}
