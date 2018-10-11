package dlm

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/redis.v4"
)

const redisReleaseScript = `
	if redis.call("get",KEYS[1]) == ARGV[1] then
		return redis.call("del",KEYS[1])
	else
		return 0
	end`

// RedisDLM is a DLM that uses Redis as a backend. The implementation is based in
// the single node algorithm described here: http://redis.io/topics/distlock
type RedisDLM struct {
	client    *redis.Client
	namespace string
}

// NewRedisDLM creates a new RedisDLM.
func NewRedisDLM(addr string, opts *Options) (DLM, error) {
	if opts == nil {
		opts = &Options{}
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisDLM{client, opts.Namespace}, nil
}

// NewLock creates a lock for the given key. The returned lock is not held
// and must be acquired with a call to .Lock.
func (d *RedisDLM) NewLock(key string, opts *LockOptions) (Locker, error) {
	if opts == nil {
		opts = &LockOptions{}
	}

	opts = opts.WithDefaults()

	token, err := randstr(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random token: %v", err)
	}

	lock := redisLock{
		ttl:       opts.TTL,
		waitTime:  opts.WaitTime,
		retryTime: opts.RetryTime,
		client:    d.client,
		namespace: d.namespace,
		key:       key,
		token:     token,
	}

	return &lock, nil
}

type redisLock struct {
	mutex sync.Mutex // Used while manipulating the internal state of the lock itself

	client *redis.Client

	ttl       time.Duration
	waitTime  time.Duration
	retryTime time.Duration

	namespace string
	key       string
	token     string // A random string used to safely release the lock
	isHeld    bool
}

func (l *redisLock) Key() string {
	return l.key
}

func (l *redisLock) Namespace() string {
	return l.namespace
}

func (l *redisLock) Lock() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.isHeld {
		return ErrLockHeld
	}

	key := l.namespace + l.key

	ok, err := l.client.SetNX(key, l.token, l.ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	if ok {
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
			ok, err := l.client.SetNX(key, l.token, l.ttl).Result()
			if err != nil {
				timeout.Stop()
				return fmt.Errorf("failed to acquire lock: %v", err)
			}

			if ok {
				l.isHeld = true
				timeout.Stop()
				return nil
			}
		}
	}
}

func (l *redisLock) Unlock() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.isHeld {
		return ErrLockNotHeld
	}

	key := l.namespace + l.key

	n, err := l.client.Eval(redisReleaseScript, []string{key}, l.token).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %v", err)
	}

	if n.(int64) == 0 {
		// the lock has already expired
		return ErrLockNotHeld
	}

	l.isHeld = false
	return nil
}
