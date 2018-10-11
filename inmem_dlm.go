package dlm

import (
	"sync"
	"time"
)

// InMemDLM is a local implementation of a DLM that it's intended to be used during development.
type InMemDLM struct {
	mutex     sync.Mutex
	keys      map[string]inMemEntry
	namespace string
}

// NewInMemDLM creates a new InMemDLM.
func NewInMemDLM(opts *Options) DLM {
	if opts == nil {
		opts = &Options{}
	}

	keys := make(map[string]inMemEntry)

	return &InMemDLM{keys: keys, namespace: opts.Namespace}
}

// NewLock creates a lock for the given key. The returned lock is not held
// and must be acquired with a call to .Lock.
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
		namespace: d.namespace,
		key:       key,
		token:     token,
	}

	return &lock, nil
}

func (d *InMemDLM) acquire(key, token string, ttl time.Duration) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := time.Now()
	key = d.namespace + key

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

	key = d.namespace + key

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

	namespace string
	key       string
	token     string // A random string used to safely release the lock
	isHeld    bool
}

func (l *inMemLock) Key() string {
	return l.key
}

func (l *inMemLock) Namespace() string {
	return l.namespace
}

func (l *inMemLock) Lock() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.isHeld {
		return ErrLockHeld
	}

	if l.dlm.acquire(l.key, l.token, l.ttl) {
		l.isHeld = true
		return nil
	}

	timeout := time.NewTimer(l.waitTime)
	retry := time.NewTicker(l.retryTime)
	defer retry.Stop()

	for {
		select {
		case <-timeout.C:
			return ErrCannotLock
		case <-retry.C:
			if l.dlm.acquire(l.key, l.token, l.ttl) {
				l.isHeld = true
				timeout.Stop()
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
