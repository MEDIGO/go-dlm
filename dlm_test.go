package dlm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func RunDLMLockTest(t *testing.T, dlm DLM) {
	// create a lock for a key
	lockA, err := dlm.NewLock("dlm-lock-test", nil)
	require.NoError(t, err)

	// it should adquire the lock
	err = lockA.Lock()
	require.NoError(t, err)

	// it should fail to adquire the lock twice
	err = lockA.Lock()
	require.Equal(t, ErrLockHeld, err)

	// it should release the lock
	err = lockA.Unlock()
	require.NoError(t, err)

	// it should fail to release the lock twice
	err = lockA.Unlock()
	require.Equal(t, ErrLockNotHeld, err)

	// it should adquire the lock again
	err = lockA.Lock()
	require.NoError(t, err)

	// create a second lock for the same key
	lockB, err := dlm.NewLock("dlm-lock-test", nil)
	require.NoError(t, err)

	// it should block until the lock A is released
	locked := make(chan bool)

	go func(locked chan<- bool) {
		err := lockB.Lock()
		require.NoError(t, err)
		locked <- true
	}(locked)

	select {
	case <-locked:
		require.Fail(t, "lock succeded on a key locked by another client")
	case <-time.After(3 * time.Second):
		err := lockA.Unlock()
		require.NoError(t, err)
		break
	}
}

func RunDLMLockTTLTest(t *testing.T, dlm DLM) {
	// create a lock for a key
	lockA, err := dlm.NewLock("dlm-lock-ttl-test", &LockOptions{
		TTL: 2 * time.Second,
	})
	require.NoError(t, err)

	// it should adquire the lock
	err = lockA.Lock()
	require.NoError(t, err)

	// create a second lock for the same key
	lockB, err := dlm.NewLock("dlm-lock-ttl-test", &LockOptions{
		RetryTime: 1 * time.Second,
	})
	require.NoError(t, err)

	// it should lock after the TTL of the lock A has expired
	locked := make(chan bool)

	go func(locked chan<- bool) {
		err := lockB.Lock()
		require.NoError(t, err)
		locked <- true
	}(locked)

	select {
	case <-locked:
		err := lockB.Unlock()
		require.NoError(t, err)
	case <-time.After(4 * time.Second):
		require.Fail(t, "failed to expire lock")
	}
}

func RunDLMLockWaitTimeTest(t *testing.T, dlm DLM) {
	// create a lock for a key
	lockA, err := dlm.NewLock("dlm-lock-wait-time-test", nil)
	require.NoError(t, err)

	// it should adquire the lock
	err = lockA.Lock()
	require.NoError(t, err)

	// create a second lock for the same key
	lockB, err := dlm.NewLock("dlm-lock-wait-time-test", &LockOptions{
		WaitTime: 2 * time.Second,
	})
	require.NoError(t, err)

	// it should fail after the wait time has ended
	done := make(chan bool)

	go func(done chan<- bool) {
		err := lockB.Lock()
		require.Equal(t, ErrCannotLock, err)
		done <- true
	}(done)

	select {
	case <-done:
		err := lockA.Unlock()
		require.NoError(t, err)
		break
	case <-time.After(3 * time.Second):
		require.Fail(t, "failed to timeout during lock")
	}
}
