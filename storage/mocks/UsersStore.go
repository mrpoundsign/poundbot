// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import storage "github.com/poundbot/poundbot/storage"
import "github.com/poundbot/poundbot/pkg/models"

// UsersStore is an autogenerated mock type for the UsersStore type
type UsersStore struct {
	mock.Mock
}

// GetByDiscordID provides a mock function with given fields: snowflake
func (_m *UsersStore) GetByDiscordID(snowflake string) (models.User, error) {
	ret := _m.Called(snowflake)

	var r0 models.User
	if rf, ok := ret.Get(0).(func(string) models.User); ok {
		r0 = rf(snowflake)
	} else {
		r0 = ret.Get(0).(models.User)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(snowflake)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByPlayerID provides a mock function with given fields: PlayerID
func (_m *UsersStore) GetByPlayerID(PlayerID string) (models.User, error) {
	ret := _m.Called(PlayerID)

	var r0 models.User
	if rf, ok := ret.Get(0).(func(string) models.User); ok {
		r0 = rf(PlayerID)
	} else {
		r0 = ret.Get(0).(models.User)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(PlayerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPlayerIDsByDiscordIDs provides a mock function with given fields: snowflakes
func (_m *UsersStore) GetPlayerIDsByDiscordIDs(snowflakes []string) ([]string, error) {
	ret := _m.Called(snowflakes)

	var r0 []string
	if rf, ok := ret.Get(0).(func([]string) []string); ok {
		r0 = rf(snowflakes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(snowflakes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemovePlayerID provides a mock function with given fields: snowflake, playerID
func (_m *UsersStore) RemovePlayerID(snowflake string, playerID string) error {
	ret := _m.Called(snowflake, playerID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(snowflake, playerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpsertPlayer provides a mock function with given fields: info
func (_m *UsersStore) UpsertPlayer(info storage.UserInfoGetter) error {
	ret := _m.Called(info)

	var r0 error
	if rf, ok := ret.Get(0).(func(storage.UserInfoGetter) error); ok {
		r0 = rf(info)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
