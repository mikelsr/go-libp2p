// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mikelsr/go-libp2p/core/network (interfaces: ProtocolScope)

// Package mocknetwork is a generated GoMock package.
package mocknetwork

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	network "github.com/mikelsr/go-libp2p/core/network"
	protocol "github.com/mikelsr/go-libp2p/core/protocol"
)

// MockProtocolScope is a mock of ProtocolScope interface.
type MockProtocolScope struct {
	ctrl     *gomock.Controller
	recorder *MockProtocolScopeMockRecorder
}

// MockProtocolScopeMockRecorder is the mock recorder for MockProtocolScope.
type MockProtocolScopeMockRecorder struct {
	mock *MockProtocolScope
}

// NewMockProtocolScope creates a new mock instance.
func NewMockProtocolScope(ctrl *gomock.Controller) *MockProtocolScope {
	mock := &MockProtocolScope{ctrl: ctrl}
	mock.recorder = &MockProtocolScopeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProtocolScope) EXPECT() *MockProtocolScopeMockRecorder {
	return m.recorder
}

// BeginSpan mocks base method.
func (m *MockProtocolScope) BeginSpan() (network.ResourceScopeSpan, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginSpan")
	ret0, _ := ret[0].(network.ResourceScopeSpan)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginSpan indicates an expected call of BeginSpan.
func (mr *MockProtocolScopeMockRecorder) BeginSpan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginSpan", reflect.TypeOf((*MockProtocolScope)(nil).BeginSpan))
}

// Protocol mocks base method.
func (m *MockProtocolScope) Protocol() protocol.ID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Protocol")
	ret0, _ := ret[0].(protocol.ID)
	return ret0
}

// Protocol indicates an expected call of Protocol.
func (mr *MockProtocolScopeMockRecorder) Protocol() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Protocol", reflect.TypeOf((*MockProtocolScope)(nil).Protocol))
}

// ReleaseMemory mocks base method.
func (m *MockProtocolScope) ReleaseMemory(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ReleaseMemory", arg0)
}

// ReleaseMemory indicates an expected call of ReleaseMemory.
func (mr *MockProtocolScopeMockRecorder) ReleaseMemory(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseMemory", reflect.TypeOf((*MockProtocolScope)(nil).ReleaseMemory), arg0)
}

// ReserveMemory mocks base method.
func (m *MockProtocolScope) ReserveMemory(arg0 int, arg1 byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReserveMemory", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReserveMemory indicates an expected call of ReserveMemory.
func (mr *MockProtocolScopeMockRecorder) ReserveMemory(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReserveMemory", reflect.TypeOf((*MockProtocolScope)(nil).ReserveMemory), arg0, arg1)
}

// Stat mocks base method.
func (m *MockProtocolScope) Stat() network.ScopeStat {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stat")
	ret0, _ := ret[0].(network.ScopeStat)
	return ret0
}

// Stat indicates an expected call of Stat.
func (mr *MockProtocolScopeMockRecorder) Stat() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stat", reflect.TypeOf((*MockProtocolScope)(nil).Stat))
}
