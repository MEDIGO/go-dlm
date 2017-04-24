// Package dlm provides an abstraction library for multiple Distributed Lock Manager backends.
package dlm

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

const (
	// DefaultTTL is the TTL of the lock after wich it will be automatically released.
	DefaultTTL = 15 * time.Second

	// DefaultWaitTime is how long will we wait to acquire a lock.
	DefaultWaitTime = 15 * time.Second

	// DefaultRetryTime is how long will we wait after a failed lock adquisition
	// before attempting to acquire the lock again.
	DefaultRetryTime = 5 * time.Second
)

var (
	// ErrLockHeld is returned if we attempt to acquire a lock that is already held.
	ErrLockHeld = fmt.Errorf("lock already held")

	// ErrLockNotHeld is returned if we attempt to release a lock that is not held.
	ErrLockNotHeld = fmt.Errorf("lock not held")

	// ErrCannotLock is returned when it's not possible to acquire the lock before
	// the configured wait time ends.
	ErrCannotLock = fmt.Errorf("timeout while trying to acquire lock")
)

// DLM describes a Distributed Lock Manager.
type DLM interface {
	// NewLock creates a lock for the given key. The returned lock is not held
	// and must be acquired with a call to .Lock.
	NewLock(string, *LockOptions) (Locker, error)
}

// Options parameterizes a DLM.
type Options struct {
	Namespace string // Optional namespace prefixed to each lock key.
}

// LockOptions parameterizes a lock.
type LockOptions struct {
	TTL       time.Duration // Optional, defaults to DefaultTTL
	WaitTime  time.Duration // Optional, defaults to DefaultWaitTime
	RetryTime time.Duration // Optional, defaults to DefaultRetryTime
}

// WithDefaults returns the options with all default values set.
func (lo *LockOptions) WithDefaults() *LockOptions {
	if lo.TTL == 0 {
		lo.TTL = DefaultTTL
	}

	if lo.WaitTime == 0 {
		lo.WaitTime = DefaultWaitTime
	}

	if lo.RetryTime == 0 {
		lo.RetryTime = DefaultRetryTime
	}

	return lo
}

// Locker describes a lock that can be locked or unlocked.
type Locker interface {
	// Key returns the key to be locked.
	Key() string

	// Namespace returns the the prefix to be added to the key. If there is not a namespace,
	// it returns the empty string.
	Namespace() string

	// Lock acquires the lock. It fails with error if the lock is already held.
	Lock() error

	// Unlock releases the lock. It fails with error if the lock is not currently held.
	Unlock() error
}

func randstr(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buffer), nil
}
