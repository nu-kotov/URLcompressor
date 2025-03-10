package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/utils"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"github.com/sqids/sqids-go"
	"github.com/stretchr/testify/assert"
)

func TestCompressURL(t *testing.T) {
	config := config.ParseConfig()
	store, err := storage.NewStorage(config)
	assert.NoError(t, err, "storage initializing error")

	service := InitService(config, store)
	assert.NoError(t, err, "Init service error")

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

func TestGetShortURL(t *testing.T) {
	var jsonBody models.ShortenURLResponse

	var config config.Config
	config.RunAddr = "localhost:8080"
	config.BaseURL = "http://localhost:8080"
	config.FileStoragePath = "/tmp/test_file.json"
	store, err := storage.NewStorage(config)
	assert.NoError(t, err, "storage initializing error")

	service := InitService(config, store)

	handler := http.HandlerFunc(service.GetShortURL)
	server := httptest.NewServer(handler)

	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	assert.NoError(t, err, "parsing server URL error")

	config.RunAddr = parsedURL.Host
	config.BaseURL = server.URL

	testURL := "https://practicum.yandex.ru"
	URLHash := utils.Hash([]byte(testURL))

	sqids, err := sqids.New()
	assert.NoError(t, err, "error making sqids variable")

	shortID, err := sqids.Encode([]uint64{URLHash})
	assert.NoError(t, err, "error making %v expected variable", shortID)

	type want struct {
		statusCode     int
		contentType    string
		method         string
		reqBody        string
		expectedResult string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Return response with short URL",
			want: want{
				statusCode:     201,
				contentType:    "application/json",
				method:         http.MethodPost,
				reqBody:        `{ "url": "https://practicum.yandex.ru" }`,
				expectedResult: "http://localhost:8080/kWmQM25LjF5r",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:     400,
				contentType:    "",
				method:         http.MethodGet,
				reqBody:        `{ "url": "https://practicum.yandex.ru" }`,
				expectedResult: "",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:     400,
				contentType:    "",
				method:         http.MethodPut,
				reqBody:        `{ "url": "https://practicum.yandex.ru" }`,
				expectedResult: "",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:     400,
				contentType:    "",
				method:         http.MethodDelete,
				reqBody:        `{ "url": "https://practicum.yandex.ru" }`,
				expectedResult: "",
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				statusCode:     400,
				contentType:    "",
				method:         http.MethodPatch,
				reqBody:        `{ "url": "https://practicum.yandex.ru" }`,
				expectedResult: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req := resty.New().R()
			req.Method = test.want.method
			req.URL = server.URL
			req.Body = test.want.reqBody
			req.Header.Set("Content-Type", "application/json")

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, test.want.statusCode, resp.StatusCode(), "Response statusCode didn't match expected")

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, resp.Header()["Content-Type"][0])
			}
			if test.want.expectedResult != "" {
				json.Unmarshal(resp.Body(), &jsonBody)
				assert.Equal(t, jsonBody.Result, test.want.expectedResult)
			}
		})
	}
}
