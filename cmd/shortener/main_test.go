package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressURLHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		method      string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Creating short url",
			want: want{
				code:        201,
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				code:        400,
				contentType: "",
				method:      http.MethodGet,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				code:        400,
				contentType: "",
				method:      http.MethodPut,
			},
		},
		{
			name: "Wrong method - resp status 400",
			want: want{
				code:        400,
				contentType: "",
				method:      http.MethodDelete,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.want.method, "/", nil)
			w := httptest.NewRecorder()
			CompressURLHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
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
