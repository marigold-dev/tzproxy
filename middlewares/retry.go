package middlewares

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/config"
)

func Retry(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.ConfigFile.TezosHostRetry == "" {
				return next(c)
			}

			writer := newDelayedResponseWriter(c)
			delayedResponse := echo.NewResponse(&writer, c.Echo())
			c.SetResponse(delayedResponse)

			err = next(c)

			status := c.Response().Status
			if status == http.StatusNotFound || status == http.StatusForbidden {
				writer.Reset()
				delayedResponse.Committed = false
				delayedResponse.Size = 0
				delayedResponse.Status = 0
				c.Set("retry", status)

				// This _error key comes from Echo (referenced in ProxyWithConfig middleware)
				// so we need to reset this as well.
				c.Set("_error", nil)

				err = next(c)
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
