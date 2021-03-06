// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	types "github.com/cosmos/cosmos-sdk/types"
	mock "github.com/stretchr/testify/mock"

	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

// ValsetKeeper is an autogenerated mock type for the ValsetKeeper type
type ValsetKeeper struct {
	mock.Mock
}

// FindSnapshotByID provides a mock function with given fields: ctx, id
func (_m *ValsetKeeper) FindSnapshotByID(ctx types.Context, id uint64) (*valsettypes.Snapshot, error) {
	ret := _m.Called(ctx, id)

	var r0 *valsettypes.Snapshot
	if rf, ok := ret.Get(0).(func(types.Context, uint64) *valsettypes.Snapshot); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*valsettypes.Snapshot)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.Context, uint64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCurrentSnapshot provides a mock function with given fields: ctx
func (_m *ValsetKeeper) GetCurrentSnapshot(ctx types.Context) (*valsettypes.Snapshot, error) {
	ret := _m.Called(ctx)

	var r0 *valsettypes.Snapshot
	if rf, ok := ret.Get(0).(func(types.Context) *valsettypes.Snapshot); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*valsettypes.Snapshot)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewValsetKeeper interface {
	mock.TestingT
	Cleanup(func())
}

// NewValsetKeeper creates a new instance of ValsetKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewValsetKeeper(t mockConstructorTestingTNewValsetKeeper) *ValsetKeeper {
	mock := &ValsetKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
