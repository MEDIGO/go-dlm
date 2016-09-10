package dlm

import "testing"

func TestInmemDLMLock(t *testing.T) {
	dlm := NewInMemDLM()

	RunDLMLockTest(t, dlm)
	RunDLMLockTTLTest(t, dlm)
	RunDLMLockWaitTimeTest(t, dlm)
}
