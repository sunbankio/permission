package utils

import "testing"

func TestHasPermission(t *testing.T) {
	testCases := []struct {
		userPermissions    []string
		requiredPermission string
		expected           bool
	}{
		{[]string{"user:create", "user:update"}, "user:create", true},
		{[]string{"user:create", "user:update"}, "user:delete", false},
		{[]string{}, "user:create", false},
		{[]string{"user:create"}, "user:create", true},
		{[]string{"admin"}, "user:create", false},                   // test case for admin
		{[]string{"user:read", "user:write"}, "user:read", true},    // test case for read and write
		{[]string{"user:read", "user:write"}, "user:delete", false}, // test case for read, write and delete
	}

	for _, tc := range testCases {
		result := HasPermission(tc.userPermissions, tc.requiredPermission)
		if result != tc.expected {
			t.Errorf("For userPermissions: %v, requiredPermission: %s, expected %v, but got %v",
				tc.userPermissions, tc.requiredPermission, tc.expected, result)
		}
	}
}
