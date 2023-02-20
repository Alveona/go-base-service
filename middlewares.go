package base_swagger_service

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	fingerprintHeader       = "X-Session-Fingerprint"
	formFileBodyPlaceholder = "\r\n\r\n<--FILE BODY REPLACED-->\r\n--"
)

var (
	expAuthorization      = regexp.MustCompile(`(?m)^Authorization:.*$`)
	expFormFieldsBoundary = regexp.MustCompile(`(?Um)^Content-Type: multipart/form-data; boundary=(.+)\r\n`)
	expFileBody           = regexp.MustCompile(`(?Us)Content-Disposition: form-data.+filename=.+\r\nContent-Type:.*(\r\n\r\n.*\r\n--)`)
	expToken              = regexp.MustCompile(`(?m)"token":\s*".*?"`)
)

func (srv *baseServiceImpl) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Panic recovered: %+v, %s", r, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, req)
	})
}

func (srv *baseServiceImpl) loggerMiddleware(mCtx *middleware.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		r = r.WithContext(ctx)

		route, _, exists := mCtx.RouteInfo(r)
		if exists && route != nil && route.Operation != nil {
			r = r.WithContext(ctx)
		}

		if srv.middlewareLogger.IsEnable(ctx) {
			srv.middlewareLogger.LogRequest(ctx, fmt.Sprintf("Handling incoming HTTP request: %.4096s", dumpRequest(r)))
		}

		wrapped := NewLoggingResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		srv.middlewareLogger.LogResponse(ctx, fmt.Sprintf("HTTP Code %d, response - %.4096s", wrapped.StatusCode, wrapped.Content), "status_code", wrapped.StatusCode)
	})
}

func dumpRequest(r *http.Request) []byte {
	dump, _ := httputil.DumpRequest(r, true)
	if r.Method == "POST" {
		dump = replaceFormFileBody(dump)
	}
	dump = expAuthorization.ReplaceAll(dump, []byte("Authorization: *****"))
	dump = expToken.ReplaceAll(dump, []byte(`"token": "*****"`))
	return dump
}

func replaceFormFileBody(text []byte) []byte {
	res := expFormFieldsBoundary.FindAllSubmatch(text, 2)
	if len(res) == 0 || len(res[0]) < 2 {
		return text
	}
	boundary := res[0][1]
	parts := bytes.Split(text, boundary)
	for i := range parts {
		res := expFileBody.FindAllSubmatch(parts[i], 2)
		if len(res) == 0 || len(res[0]) < 2 {
			continue
		}
		parts[i] = bytes.Replace(parts[i], res[0][1], []byte(formFileBodyPlaceholder), 1)
	}
	text = bytes.Join(parts, boundary)
	return text
}

func (srv *baseServiceImpl) metricsMiddleware(ctx *middleware.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts := time.Now()
		var handlerName string
		respWriter := NewResponseWriterWithHTTPCode(w)
		route, _, exists := ctx.RouteInfo(r)
		if exists && route != nil && route.Operation != nil {
			handlerName = route.Operation.ID
			defer func() {
				srv.statReqDurations.With(prometheus.Labels{"handler": handlerName}).Observe(time.Since(ts).Seconds())
				srv.statReqCount.With(prometheus.Labels{
					"code":    strconv.Itoa(respWriter.Code),
					"method":  r.Method,
					"handler": handlerName,
				}).Inc()
			}()
		}
		next.ServeHTTP(respWriter, r)
	})
}

func (srv *baseServiceImpl) profileMiddleware(ctx *middleware.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, _, exists := ctx.RouteInfo(r)
		if !exists || route == nil || route.Operation == nil {
			next.ServeHTTP(w, r)
			return
		}

		r = r.WithContext(r.Context())

		next.ServeHTTP(w, r)
	})
}
