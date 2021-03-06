// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// AttestTask is an autogenerated mock type for the AttestTask type
type AttestTask struct {
	mock.Mock
}

// Attest provides a mock function with given fields:
func (_m *AttestTask) Attest() {
	_m.Called()
}

type mockConstructorTestingTNewAttestTask interface {
	mock.TestingT
	Cleanup(func())
}

// NewAttestTask creates a new instance of AttestTask. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAttestTask(t mockConstructorTestingTNewAttestTask) *AttestTask {
	mock := &AttestTask{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
