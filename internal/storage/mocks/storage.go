// Code generated by MockGen. DO NOT EDIT.
// Source: internal/storage/storage.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/pluhe7/shortener/internal/models"
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

// Get mocks base method.
func (m *MockStorage) Get(shortURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", shortURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStorageMockRecorder) Get(shortURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorage)(nil).Get), shortURL)
}

// GetByOriginal mocks base method.
func (m *MockStorage) GetByOriginal(originalURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByOriginal", originalURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByOriginal indicates an expected call of GetByOriginal.
func (mr *MockStorageMockRecorder) GetByOriginal(originalURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByOriginal", reflect.TypeOf((*MockStorage)(nil).GetByOriginal), originalURL)
}

// PingContext mocks base method.
func (m *MockStorage) PingContext(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingContext", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingContext indicates an expected call of PingContext.
func (mr *MockStorageMockRecorder) PingContext(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingContext", reflect.TypeOf((*MockStorage)(nil).PingContext), ctx)
}

// Save mocks base method.
func (m *MockStorage) Save(record models.ShortURLRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", record)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockStorageMockRecorder) Save(record interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockStorage)(nil).Save), record)
}

// SaveBatch mocks base method.
func (m *MockStorage) SaveBatch(records []models.ShortURLRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBatch", records)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveBatch indicates an expected call of SaveBatch.
func (mr *MockStorageMockRecorder) SaveBatch(records interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBatch", reflect.TypeOf((*MockStorage)(nil).SaveBatch), records)
}
