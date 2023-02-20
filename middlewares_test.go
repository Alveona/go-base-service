package base_swagger_service

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDumpRequest(t *testing.T) {
	testCases := []struct {
		name          string
		authorization string
		body          string
		contentType   string
		expected      string
		extraHeader   string
	}{
		{
			name:        "empty both",
			expected:    "POST test HTTP/1.1\r\nContent-Type: application/json\r\n\r\n",
			contentType: "application/json",
		},
		{
			name:          "fill authorization",
			authorization: "Legacy eyJjbGllbBiYWU0Yy1mOWEifQ==",
			expected:      "POST test HTTP/1.1\r\nAuthorization: *****\nContent-Type: application/json\r\n\r\n",
			contentType:   "application/json",
		},
		{
			name:        "fill token",
			body:        "{\"token\":\"elNGVkNz==\"}",
			expected:    "POST test HTTP/1.1\r\nContent-Type: application/json\r\n\r\n{\"token\": \"*****\"}",
			contentType: "application/json",
		},
		{
			name:          "fill authorization and token",
			authorization: "Legacy eyJjbGllbBiYWU0Yy1mOWEifQ==",
			body:          "{\"token\":  \"elNGVkNz==\"}",
			expected:      "POST test HTTP/1.1\r\nAuthorization: *****\nContent-Type: application/json\r\n\r\n{\"token\": \"*****\"}",
			contentType:   "application/json",
		},
		{
			name:          "value before token",
			authorization: "Legacy eyJjbGllbBiYWU0Yy1mOWEifQ==",
			body:          "{\"device\":\"ios\",\"token\":\"elNGVkNz==\"}",
			expected:      "POST test HTTP/1.1\r\nAuthorization: *****\nContent-Type: application/json\r\n\r\n{\"device\":\"ios\",\"token\": \"*****\"}",
			contentType:   "application/json",
		},
		{
			name:          "values before and after token",
			authorization: "Legacy eyJjbGllbBiYWU0Yy1mOWEifQ==",
			body:          "{\"device\":\"ios\",\"token\":\"elNGVkNz==\",\"id\":123}",
			expected:      "POST test HTTP/1.1\r\nAuthorization: *****\nContent-Type: application/json\r\n\r\n{\"device\":\"ios\",\"token\": \"*****\",\"id\":123}",
			contentType:   "application/json",
		},
		{
			name:        "multipart form without file",
			body:        "--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n FILE BODY\r\n\r\n TO BE REPLACED \r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
			contentType: "multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3",
			expected:    "POST test HTTP/1.1\r\nContent-Type: multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3\r\n\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
		},
		{
			name:        "multipart form",
			body:        "--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n FILE BODY\r\n\r\n TO BE REPLACED \r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
			contentType: "multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3",
			expected:    "POST test HTTP/1.1\r\nContent-Type: multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3\r\n\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
		},
		{
			name:        "multipart form with extra header",
			body:        "--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n FILE BODY\r TO BE\n REPLACED \r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
			contentType: "multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3",
			extraHeader: "some_header_value",
			expected:    "POST test HTTP/1.1\r\nContent-Type: multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3\r\nX-Extra-Header: some_header_value\r\n\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
		},
		{
			name:        "multipart form with extra header and multiple files",
			body:        "--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n FILE BODY\r TO BE\n REPLACED \r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test2.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n FILE BODY\r\n TO BE\r\n REPLACED \r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
			contentType: "multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3",
			extraHeader: "some_header_value",
			expected:    "POST test HTTP/1.1\r\nContent-Type: multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3\r\nX-Extra-Header: some_header_value\r\n\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test2.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
		},
		{
			name:        "multipart form with extra header and empty files",
			body:        "--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n FILE BODY\r TO BE\n REPLACED \r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test2.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n\r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
			contentType: "multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3",
			extraHeader: "some_header_value",
			expected:    "POST test HTTP/1.1\r\nContent-Type: multipart/form-data; boundary=qlsdvhsdrvv3t3tf3g3kub3u3\r\nX-Extra-Header: some_header_value\r\n\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3\r\nContent-Disposition: form-data; name=\"image\"; filename=\"photo_test2.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n<--FILE BODY REPLACED-->\r\n--qlsdvhsdrvv3t3tf3g3kub3u3--",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "test", bytes.NewBuffer([]byte(testCase.body)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", testCase.contentType)
			if testCase.authorization != "" {
				req.Header.Set("Authorization", testCase.authorization)
			}

			if testCase.extraHeader != "" {
				req.Header.Set("X-Extra-Header", testCase.extraHeader)
			}

			out := dumpRequest(req)
			assert.Equal(t, testCase.expected, string(out))
		})
	}
}
