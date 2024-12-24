package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/api/handler"
	"github.com/nu-kotov/URLcompressor/api/utils"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/sqids/sqids-go"
	"github.com/stretchr/testify/assert"
)

func TestCompressURL(t *testing.T) {
	config := config.ParseConfig()
	service := handler.InitService(config)

	handler := http.HandlerFunc(service.CompressURL)
	server := httptest.NewServer(handler)

	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	assert.NoError(t, err, "parsing server URL error")

	config.RunAddr = parsedURL.Host
	config.BaseURL = server.URL

	testURL := "https://stackoverflow.com"
	URLHash := utils.Hash([]byte(testURL))

	sqids, err := sqids.New()
	assert.NoError(t, err, "error making sqids variable")

	shortID, err := sqids.Encode([]uint64{URLHash})
	assert.NoError(t, err, "error making %v expected variable", shortID)

	type want struct {
		statusCode      int
		contentType     string
		method          string
		expectedShortID string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Creating short url",
			want: want{
				statusCode:      201,
				contentType:     "text/plain",
				method:          http.MethodPost,
				expectedShortID: shortID,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:      400,
				contentType:     "",
				method:          http.MethodGet,
				expectedShortID: "",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:      400,
				contentType:     "",
				method:          http.MethodPut,
				expectedShortID: "",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:      400,
				contentType:     "",
				method:          http.MethodDelete,
				expectedShortID: "",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:      400,
				contentType:     "",
				method:          http.MethodPatch,
				expectedShortID: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req := resty.New().R()
			req.Method = test.want.method
			req.URL = server.URL
			req.Body = testURL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, test.want.statusCode, resp.StatusCode(), "Response statusCode didn't match expected")

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, resp.Header()["Content-Type"][0])
			}
			if test.want.expectedShortID != "" {
				assert.True(t, strings.HasSuffix(string(resp.Body()), test.want.expectedShortID))
			}
		})
	}
}

func TestRedirectByShortUrlID(t *testing.T) {
	var config config.Config
	config.RunAddr = "localhost:8080"
	config.BaseURL = "http://localhost:8080"
	service := handler.InitService(config)

	router := mux.NewRouter()
	router.HandleFunc(`/`, service.CompressURL)
	router.HandleFunc(`/{id:\w+}`, service.RedirectByShortUrlID)
	server := httptest.NewServer(router)

	parsedURL, err := url.Parse(server.URL)
	assert.NoError(t, err, "parsing server URL error")

	config.RunAddr = parsedURL.Host
	config.BaseURL = server.URL

	defer server.Close()

	testURL := "https://stackoverflow.com"

	req := resty.New().R()
	req.URL = server.URL
	req.Body = testURL
	resp, err := req.Post(server.URL)
	assert.NoError(t, err, "error making post HTTP request to %v", server.URL)

	compressedURLSuffix := "/" + strings.Split(string(resp.Body()), "/")[3]

	type want struct {
		statusCode    int
		compressedURL string
		location      string
		method        string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Get decompressed url - success",
			want: want{
				statusCode:    http.StatusOK,
				compressedURL: compressedURLSuffix,
				location:      testURL,
				method:        http.MethodGet,
			},
		},
		{
			name: "Get decompressed url - non-existent",
			want: want{
				statusCode:    http.StatusBadRequest,
				compressedURL: "/nonExistent",
				location:      "",
				method:        http.MethodGet,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:    http.StatusBadRequest,
				compressedURL: compressedURLSuffix,
				location:      "",
				method:        http.MethodPost,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:    http.StatusBadRequest,
				compressedURL: compressedURLSuffix,
				location:      "",
				method:        http.MethodPut,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:    http.StatusBadRequest,
				compressedURL: compressedURLSuffix,
				location:      "",
				method:        http.MethodDelete,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:    http.StatusBadRequest,
				compressedURL: compressedURLSuffix,
				location:      "",
				method:        http.MethodPatch,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = test.want.method
			req.URL = server.URL + test.want.compressedURL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, test.want.statusCode, resp.StatusCode(), "Response statusCode didn't match expected")

			if test.want.compressedURL != "" && test.want.statusCode != http.StatusBadRequest {
				assert.Equal(t, testURL, "https://"+resp.RawResponse.Request.URL.Host, "Response Host didn't match expected")
			}
		})
	}
}
