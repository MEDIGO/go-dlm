package dlm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisDLMLock(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.SkipNow()
	}

	dlm, err := NewRedisDLM(os.Getenv("REDIS_ADDR"))
	require.NoError(t, err)

	RunDLMLockTest(t, dlm)
	RunDLMLockTTLTest(t, dlm)
	RunDLMLockWaitTimeTest(t, dlm)
}
