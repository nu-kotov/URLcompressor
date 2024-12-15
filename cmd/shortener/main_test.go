package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/sqids/sqids-go"
	"github.com/stretchr/testify/assert"
)

func TestCompressURLHandler(t *testing.T) {

	handler := http.HandlerFunc(CompressURLHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	testURL := "https://ya.ru"
	URLHash := hash([]byte(testURL))

	sqids, newSqidsError := sqids.New()
	assert.NoError(t, newSqidsError, "error making sqids variable")

	shortID, IDCreatingError := sqids.Encode([]uint64{URLHash})
	assert.NoError(t, IDCreatingError, "error making %v expected variable", shortID)

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

func TestShortURLByID(t *testing.T) {

	router := mux.NewRouter()
	router.HandleFunc(`/`, CompressURLHandler)
	router.HandleFunc(`/{id:\w+}`, ShortURLByID)
	server := httptest.NewServer(router)

	defer server.Close()

	testURL := "http://ya.ru"

	req := resty.New().R()
	req.URL = server.URL
	req.Body = testURL
	resp, creatingShortIDError := req.Post(server.URL)
	assert.NoError(t, creatingShortIDError, "error making HTTP request")
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
				assert.Equal(t, testURL, "http://"+resp.RawResponse.Request.URL.Host, "Response Host didn't match expected")
			}
		})
	}
}
