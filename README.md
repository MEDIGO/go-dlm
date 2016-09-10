# go-dlm

[![CircleCI](https://circleci.com/gh/MEDIGO/go-dlm.svg?style=shield)](https://circleci.com/gh/MEDIGO/go-dlm)

This package provides an abstraction library for multiple Distributed Lock Manager backends.

As the moment, go-dlm only supports [Redis](http://redis.io/) and an in-memory implementation intended to be used during development.

## Usage

```go

// Create a DLM
dlm, err := NewRedisLock("localhost:6379")
if err != nil {
    panic(err)
}

// Create a lock
lock, err := dlm.NewLock("resource", nil)
if err != nil {
    panic(err)
}

// Adquire the lock
if err := lock.Lock(); err != nil {
    panic(err)
}

// Release the lock
if err := lock.Unlock(); err != nil {
    panic(err)
}
```

## Copyright and license

Copyright © 2016 MEDIGO GmbH. go-dlm is licensed under the Apache License, Version 2.0. See LICENSE for the full license text.
