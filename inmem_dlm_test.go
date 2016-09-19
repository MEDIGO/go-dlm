package dlm

import "testing"

func TestInmemDLMLock(t *testing.T) {
	dlm := NewInMemDLM(&Options{
		Namespace: "test/",
	})

	RunDLMLockTest(t, dlm)
	RunDLMLockTTLTest(t, dlm)
	RunDLMLockWaitTimeTest(t, dlm)
}
