// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/nu-kotov/URLcompressor/internal/app/models"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStorage) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStorageMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStorage)(nil).Close))
}

// DeleteURLs mocks base method.
func (m *MockStorage) DeleteURLs(ctx context.Context, data []models.URLForDeleteMsg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteURLs", ctx, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteURLs indicates an expected call of DeleteURLs.
func (mr *MockStorageMockRecorder) DeleteURLs(ctx, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteURLs", reflect.TypeOf((*MockStorage)(nil).DeleteURLs), ctx, data)
}

// InsertURLsData mocks base method.
func (m *MockStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertURLsData", ctx, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertURLsData indicates an expected call of InsertURLsData.
func (mr *MockStorageMockRecorder) InsertURLsData(ctx, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertURLsData", reflect.TypeOf((*MockStorage)(nil).InsertURLsData), ctx, data)
}

// InsertURLsDataBatch mocks base method.
func (m *MockStorage) InsertURLsDataBatch(ctx context.Context, data []models.URLsData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertURLsDataBatch", ctx, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertURLsDataBatch indicates an expected call of InsertURLsDataBatch.
func (mr *MockStorageMockRecorder) InsertURLsDataBatch(ctx, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertURLsDataBatch", reflect.TypeOf((*MockStorage)(nil).InsertURLsDataBatch), ctx, data)
}

// Ping mocks base method.
func (m *MockStorage) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockStorageMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockStorage)(nil).Ping))
}

// SelectOriginalURLByShortURL mocks base method.
func (m *MockStorage) SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectOriginalURLByShortURL", ctx, shortURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectOriginalURLByShortURL indicates an expected call of SelectOriginalURLByShortURL.
func (mr *MockStorageMockRecorder) SelectOriginalURLByShortURL(ctx, shortURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectOriginalURLByShortURL", reflect.TypeOf((*MockStorage)(nil).SelectOriginalURLByShortURL), ctx, shortURL)
}

// SelectURLs mocks base method.
func (m *MockStorage) SelectURLs(ctx context.Context, userID string) ([]models.GetUserURLsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectURLs", ctx, userID)
	ret0, _ := ret[0].([]models.GetUserURLsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectURLs indicates an expected call of SelectURLs.
func (mr *MockStorageMockRecorder) SelectURLs(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectURLs", reflect.TypeOf((*MockStorage)(nil).SelectURLs), ctx, userID)
}
