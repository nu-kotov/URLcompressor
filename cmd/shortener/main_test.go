package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sqids/sqids-go"
	"github.com/stretchr/testify/assert"
)

func TestCompressURLHandler(t *testing.T) {

	handler := http.HandlerFunc(CompressURLHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	testURL := []byte("http://ya.ru")
	URLHash := hash(testURL)

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
			req.Body = "http://ya.ru"

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
	type want struct {
		code     int
		url      string
		location string
		method   string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Creating short url",
			want: want{
				code:   400,
				url:    "/compressedID",
				method: http.MethodGet,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				code:   400,
				url:    "",
				method: http.MethodPost,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				code:   400,
				url:    "",
				method: http.MethodPut,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				code:   400,
				url:    "",
				method: http.MethodDelete,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.want.method, "/compressedID", nil)
			w := httptest.NewRecorder()
			ShortURLByID(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			if test.want.url != "" {
				assert.Equal(t, test.want.url, request.URL.String(), "URL запроса не совпадает с ожидаемым")
			}
		})
	}
}
