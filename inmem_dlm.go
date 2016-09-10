package dlm

import (
	"sync"
	"time"
)

// InMemDLM is a local implementation of a DLM that it's intended to be used during development.
type InMemDLM struct {
	mutex sync.Mutex
	keys  map[string]inMemEntry
}

// NewInMemDLM creates a new InMemDLM.
func NewInMemDLM() DLM {
	return &InMemDLM{keys: make(map[string]inMemEntry)}
}

// NewLock creates a lock for the given key. The returned lock is not held
// and must be adquired with a call to .Lock.
func (d *InMemDLM) NewLock(key string, opts *LockOptions) (Locker, error) {
	if opts == nil {
		opts = &LockOptions{}
	}

	opts = opts.WithDefaults()

	token, err := randstr(32)
	if err != nil {
		return nil, err
	}

	lock := inMemLock{
		ttl:       opts.TTL,
		waitTime:  opts.WaitTime,
		retryTime: opts.RetryTime,
		dlm:       d,
		key:       key,
		token:     token,
	}

	return &lock, nil
}

func (d *InMemDLM) adquire(key, token string, ttl time.Duration) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := time.Now()

	entry, ok := d.keys[key]
	if ok && entry.validUntil.After(now) {
		return false
	}

	d.keys[key] = inMemEntry{token, now.Add(ttl)}

	return true
}

func (d *InMemDLM) release(key, token string) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	entry, ok := d.keys[key]
	if !ok || entry.token != token {
		return false
	}

	delete(d.keys, key)

	return true
}

type inMemEntry struct {
	token      string
	validUntil time.Time
}

type inMemLock struct {
	mutex sync.Mutex // Used while manipulating the internal state of the lock itself

	dlm *InMemDLM

	ttl       time.Duration
	waitTime  time.Duration
	retryTime time.Duration

	key    string
	token  string // A random string used to safely release the lock
	isHeld bool
}

func (l *inMemLock) Lock() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.isHeld {
		return ErrLockHeld
	}

	if l.dlm.adquire(l.key, l.token, l.ttl) {
		l.isHeld = true
		return nil
	}

	timeout := time.After(l.waitTime)
	retry := time.Tick(l.retryTime)

	for {
		select {
		case <-timeout:
			return ErrCannotLock
		case <-retry:
			if l.dlm.adquire(l.key, l.token, l.ttl) {
				l.isHeld = true
				return nil
			}
		}
	}
}

func (l *inMemLock) Unlock() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.isHeld || !l.dlm.release(l.key, l.token) {
		return ErrLockNotHeld
	}

	l.isHeld = false
	return nil
}
