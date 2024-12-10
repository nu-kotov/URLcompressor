package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestpostCompressURLHandler(t *testing.T) {
	responseBody := "http://localhost:8080/" + generateShortKey()
	type want struct {
		code        int
		response    string
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
				response:    responseBody,
				contentType: "text/plain",
				method:      http.MethodPost,
			},
		},
		{
			name: "Wrong method - 405",
			want: want{
				code:        405,
				response:    "",
				contentType: "",
				method:      http.MethodGet,
			},
		},
		{
			name: "Wrong method - 405",
			want: want{
				code:        405,
				response:    "",
				contentType: "",
				method:      http.MethodPut,
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
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
			if test.want.response != "" {
				assert.Equal(t, test.want.response, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
			assert.JSONEq(t, test.want.response, string(resBody))

		})
	}
}
