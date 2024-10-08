// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	xchain "github.com/palomachain/paloma/v2/internal/x-chain"
)

// EvmKeeper is an autogenerated mock type for the EvmKeeper type
type EvmKeeper struct {
	mock.Mock
}

// PickValidatorForMessage provides a mock function with given fields: ctx, chainReferenceID, requirements
func (_m *EvmKeeper) PickValidatorForMessage(ctx context.Context, chainReferenceID string, requirements *xchain.JobRequirements) (string, string, error) {
	ret := _m.Called(ctx, chainReferenceID, requirements)

	if len(ret) == 0 {
		panic("no return value specified for PickValidatorForMessage")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *xchain.JobRequirements) (string, string, error)); ok {
		return rf(ctx, chainReferenceID, requirements)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *xchain.JobRequirements) string); ok {
		r0 = rf(ctx, chainReferenceID, requirements)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *xchain.JobRequirements) string); ok {
		r1 = rf(ctx, chainReferenceID, requirements)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, *xchain.JobRequirements) error); ok {
		r2 = rf(ctx, chainReferenceID, requirements)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewEvmKeeper creates a new instance of EvmKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEvmKeeper(t interface {
	mock.TestingT
	Cleanup(func())
}) *EvmKeeper {
	mock := &EvmKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
