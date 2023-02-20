package base_swagger_service

import "net/http"

type LoggingResponseWriter struct {
	wrapped    http.ResponseWriter
	StatusCode int
	Content    []byte
}

func NewLoggingResponseWriter(wrapped http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{wrapped: wrapped}
}

func (lrw *LoggingResponseWriter) Header() http.Header {
	return lrw.wrapped.Header()
}

func (lrw *LoggingResponseWriter) Write(content []byte) (int, error) {
	lrw.Content = content
	return lrw.wrapped.Write(content)
}

func (lrw *LoggingResponseWriter) WriteHeader(statusCode int) {
	lrw.StatusCode = statusCode
	lrw.wrapped.WriteHeader(statusCode)
}

type ResponseWriterWithHTTPCode struct {
	wrapped http.ResponseWriter
	Code    int
}

func NewResponseWriterWithHTTPCode(wrapped http.ResponseWriter) *ResponseWriterWithHTTPCode {
	return &ResponseWriterWithHTTPCode{wrapped: wrapped}
}

func (rw *ResponseWriterWithHTTPCode) Header() http.Header {
	return rw.wrapped.Header()
}

func (rw *ResponseWriterWithHTTPCode) Write(content []byte) (int, error) {
	return rw.wrapped.Write(content)
}

func (rw *ResponseWriterWithHTTPCode) WriteHeader(statusCode int) {
	rw.Code = statusCode
	rw.wrapped.WriteHeader(statusCode)
}
