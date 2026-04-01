// Package testutil provides testing utilities and helpers for the Refyne backend.
package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestContext creates a context with a test-appropriate timeout.
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// RandomEmail generates a random email for testing.
func RandomEmail() string {
	return fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
}

// RandomUsername generates a random username for testing.
func RandomUsername() string {
	return fmt.Sprintf("user_%s", uuid.New().String()[:8])
}

// RandomUUID generates a random UUID string.
func RandomUUID() string {
	return uuid.New().String()
}

// AssertNoError is a helper that fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}

// AssertError is a helper that fails the test if err is nil.
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
}

// AssertEqual is a helper that fails the test if expected != actual.
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	require.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotNil is a helper that fails the test if obj is nil.
func AssertNotNil(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	require.NotNil(t, obj, msgAndArgs...)
}
