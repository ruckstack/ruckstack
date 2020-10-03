package test_util

import (
	"github.com/stretchr/testify/assert"
	"os/user"
	"testing"
)

var (
	currentUser *user.User
)

func GetCurrentUser(t *testing.T) *user.User {
	currentUser, err := user.Current()
	assert.NoError(t, err)

	return currentUser
}

func GetCurrentUserGroup(t *testing.T) *user.Group {
	currentUserGroup, err := user.LookupGroupId(GetCurrentUser(t).Gid)
	assert.NoError(t, err)

	return currentUserGroup
}
