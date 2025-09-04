package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/service"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
)

// ExampleHandler_CompressURL демонстрирует создание короткого URL.
func ExampleHandler_CompressURL() {
	var config config.Config
	store, _ := storage.NewStorage(config)
	service := service.NewURLService(config, store)
	HTTPHandler := NewHandler(config, service, store, nil)
	router := NewRouter(*HTTPHandler)

	reqBody := []byte("https://stackoverflow.com")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	respRec := httptest.NewRecorder()
	router.ServeHTTP(respRec, req)

	fmt.Println(respRec.Code)
	// Output: 201
}

// ExampleService_GetShortURL демонстрирует получение короткого урла в качестве ответа.
func ExampleHandler_GetShortURL() {
	var config config.Config
	store, _ := storage.NewStorage(config)
	service := service.NewURLService(config, store)
	HTTPHandler := NewHandler(config, service, store, nil)
	router := NewRouter(*HTTPHandler)

	reqBody, _ := json.Marshal(models.ShortenURLRequest{URL: "https://stackoverflow.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	respRec := httptest.NewRecorder()
	router.ServeHTTP(respRec, req)

	fmt.Println(respRec.Code)
	// Output: 201
}

// ExampleService_RedirectByShortURLID демонстрирует редирект на полный урл, по id короткого урла.
func ExampleHandler_RedirectByShortURLID() {
	var config config.Config
	store, _ := storage.NewStorage(config)
	service := service.NewURLService(config, store)
	HTTPHandler := NewHandler(config, service, store, nil)
	router := NewRouter(*HTTPHandler)

	reqBody := []byte("https://stackoverflow.com")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	respCreateRec := httptest.NewRecorder()
	router.ServeHTTP(respCreateRec, req)

	shortURL := strings.TrimSpace(respCreateRec.Body.String())
	shortID := strings.TrimPrefix(shortURL, config.BaseURL+"/")

	reqRedirect := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	reqRedirectRec := httptest.NewRecorder()
	router.ServeHTTP(reqRedirectRec, reqRedirect)

	fmt.Println(reqRedirectRec.Code)
	// Output: 307
}

// ExampleService_GetUserURLs демонстрирует получение всех урлов пользователя.
func ExampleHandler_GetUserURLs() {
	var config config.Config
	store, _ := storage.NewStorage(config)
	service := service.NewURLService(config, store)
	HTTPHandler := NewHandler(config, service, store, nil)
	router := NewRouter(*HTTPHandler)

	reqBody := []byte("https://stackoverflow.com")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	respCreateRec := httptest.NewRecorder()
	router.ServeHTTP(respCreateRec, req)

	createResult := respCreateRec.Result()
	defer createResult.Body.Close()

	cookies := createResult.Cookies()

	urlsReq := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	for _, cookie := range cookies {
		urlsReq.AddCookie(cookie)
	}

	respURLsRec := httptest.NewRecorder()
	router.ServeHTTP(respURLsRec, urlsReq)

	defer respURLsRec.Result().Body.Close()

	fmt.Println(respURLsRec.Code)
	// Output: 200
}

// ExampleService_DeleteUserURLs демонстрирует удаления урла пользователя.
func ExampleHandler_DeleteUserURLs() {
	var config config.Config
	store, _ := storage.NewStorage(config)
	service := service.NewURLService(config, store)
	HTTPHandler := NewHandler(config, service, store, nil)
	router := NewRouter(*HTTPHandler)

	reqBody := []byte("https://stackoverflow.com")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	respCreateRec := httptest.NewRecorder()
	router.ServeHTTP(respCreateRec, req)

	createResult := respCreateRec.Result()
	defer createResult.Body.Close()

	cookies := createResult.Cookies()

	shortURL := strings.TrimSpace(respCreateRec.Body.String())
	shortID := strings.TrimPrefix(shortURL, config.BaseURL+"/")
	deleteBody, _ := json.Marshal([]string{shortID})

	delReq := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(deleteBody))

	for _, cookie := range cookies {
		delReq.AddCookie(cookie)
	}

	respDelRec := httptest.NewRecorder()
	router.ServeHTTP(respDelRec, delReq)

	defer respDelRec.Result().Body.Close()

	fmt.Println(respDelRec.Code)
	// Output: 202
}

// ExampleService_GetShortURLsBatch демонстрирует сокращение батча URL'ов.
func ExampleHandler_GetShortURLsBatch() {
	var config config.Config
	store, _ := storage.NewStorage(config)
	service := service.NewURLService(config, store)
	HTTPHandler := NewHandler(config, service, store, nil)
	router := NewRouter(*HTTPHandler)

	reqBody, _ := json.Marshal([]models.GetShortURLsBatchRequest{
		{CorrelationID: "1", OriginalURL: "https://stackoverflow.com"},
		{CorrelationID: "2", OriginalURL: "http://ya.ru"},
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	respRec := httptest.NewRecorder()
	router.ServeHTTP(respRec, req)

	fmt.Println(respRec.Code)
	// Output: 201
}
