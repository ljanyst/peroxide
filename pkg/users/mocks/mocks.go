// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ljanyst/peroxide/pkg/users (interfaces: Locator,CredentialsStorer,StoreMaker)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	store "github.com/ljanyst/peroxide/pkg/store"
	credentials "github.com/ljanyst/peroxide/pkg/users/credentials"
)

// MockLocator is a mock of Locator interface.
type MockLocator struct {
	ctrl     *gomock.Controller
	recorder *MockLocatorMockRecorder
}

// MockLocatorMockRecorder is the mock recorder for MockLocator.
type MockLocatorMockRecorder struct {
	mock *MockLocator
}

// NewMockLocator creates a new mock instance.
func NewMockLocator(ctrl *gomock.Controller) *MockLocator {
	mock := &MockLocator{ctrl: ctrl}
	mock.recorder = &MockLocatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLocator) EXPECT() *MockLocatorMockRecorder {
	return m.recorder
}

// Clear mocks base method.
func (m *MockLocator) Clear() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Clear")
	ret0, _ := ret[0].(error)
	return ret0
}

// Clear indicates an expected call of Clear.
func (mr *MockLocatorMockRecorder) Clear() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clear", reflect.TypeOf((*MockLocator)(nil).Clear))
}

// MockCredentialsStorer is a mock of CredentialsStorer interface.
type MockCredentialsStorer struct {
	ctrl     *gomock.Controller
	recorder *MockCredentialsStorerMockRecorder
}

// MockCredentialsStorerMockRecorder is the mock recorder for MockCredentialsStorer.
type MockCredentialsStorerMockRecorder struct {
	mock *MockCredentialsStorer
}

// NewMockCredentialsStorer creates a new mock instance.
func NewMockCredentialsStorer(ctrl *gomock.Controller) *MockCredentialsStorer {
	mock := &MockCredentialsStorer{ctrl: ctrl}
	mock.recorder = &MockCredentialsStorerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCredentialsStorer) EXPECT() *MockCredentialsStorerMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockCredentialsStorer) Add(arg0, arg1, arg2, arg3 string, arg4 []byte, arg5 []string) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Add indicates an expected call of Add.
func (mr *MockCredentialsStorerMockRecorder) Add(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockCredentialsStorer)(nil).Add), arg0, arg1, arg2, arg3, arg4, arg5)
}

// Delete mocks base method.
func (m *MockCredentialsStorer) Delete(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockCredentialsStorerMockRecorder) Delete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockCredentialsStorer)(nil).Delete), arg0)
}

// Get mocks base method.
func (m *MockCredentialsStorer) Get(arg0 string) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCredentialsStorerMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCredentialsStorer)(nil).Get), arg0)
}

// List mocks base method.
func (m *MockCredentialsStorer) List() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockCredentialsStorerMockRecorder) List() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockCredentialsStorer)(nil).List))
}

// Logout mocks base method.
func (m *MockCredentialsStorer) Logout(arg0 string) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logout", arg0)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Logout indicates an expected call of Logout.
func (mr *MockCredentialsStorerMockRecorder) Logout(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logout", reflect.TypeOf((*MockCredentialsStorer)(nil).Logout), arg0)
}

// SwitchAddressMode mocks base method.
func (m *MockCredentialsStorer) SwitchAddressMode(arg0 string) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SwitchAddressMode", arg0)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SwitchAddressMode indicates an expected call of SwitchAddressMode.
func (mr *MockCredentialsStorerMockRecorder) SwitchAddressMode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SwitchAddressMode", reflect.TypeOf((*MockCredentialsStorer)(nil).SwitchAddressMode), arg0)
}

// UpdateEmails mocks base method.
func (m *MockCredentialsStorer) UpdateEmails(arg0 string, arg1 []string) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateEmails", arg0, arg1)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateEmails indicates an expected call of UpdateEmails.
func (mr *MockCredentialsStorerMockRecorder) UpdateEmails(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEmails", reflect.TypeOf((*MockCredentialsStorer)(nil).UpdateEmails), arg0, arg1)
}

// UpdatePassword mocks base method.
func (m *MockCredentialsStorer) UpdatePassword(arg0 string, arg1 []byte) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", arg0, arg1)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdatePassword indicates an expected call of UpdatePassword.
func (mr *MockCredentialsStorerMockRecorder) UpdatePassword(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockCredentialsStorer)(nil).UpdatePassword), arg0, arg1)
}

// UpdateToken mocks base method.
func (m *MockCredentialsStorer) UpdateToken(arg0, arg1, arg2 string) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateToken", arg0, arg1, arg2)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateToken indicates an expected call of UpdateToken.
func (mr *MockCredentialsStorerMockRecorder) UpdateToken(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateToken", reflect.TypeOf((*MockCredentialsStorer)(nil).UpdateToken), arg0, arg1, arg2)
}

// MockStoreMaker is a mock of StoreMaker interface.
type MockStoreMaker struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMakerMockRecorder
}

// MockStoreMakerMockRecorder is the mock recorder for MockStoreMaker.
type MockStoreMakerMockRecorder struct {
	mock *MockStoreMaker
}

// NewMockStoreMaker creates a new mock instance.
func NewMockStoreMaker(ctrl *gomock.Controller) *MockStoreMaker {
	mock := &MockStoreMaker{ctrl: ctrl}
	mock.recorder = &MockStoreMakerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoreMaker) EXPECT() *MockStoreMakerMockRecorder {
	return m.recorder
}

// New mocks base method.
func (m *MockStoreMaker) New(arg0 store.BridgeUser, arg1 bool) (*store.Store, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", arg0, arg1)
	ret0, _ := ret[0].(*store.Store)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New.
func (mr *MockStoreMakerMockRecorder) New(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockStoreMaker)(nil).New), arg0, arg1)
}

// Remove mocks base method.
func (m *MockStoreMaker) Remove(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockStoreMakerMockRecorder) Remove(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockStoreMaker)(nil).Remove), arg0)
}
