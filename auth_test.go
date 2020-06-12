package todos_test

import (
	"testing"

	. "github.com/bbengfort/todos"
	"github.com/stretchr/testify/require"
)

func TestPasswordDerivedKey(t *testing.T) {
	pw, err := CreateDerivedKey("supersecretpassword")
	require.NoError(t, err)

	dk, salt, tm, mem, proc, err := ParseDerivedKey(pw)
	require.NoError(t, err)
	require.Len(t, dk, 32)
	require.Len(t, salt, 16)
	require.Equal(t, uint32(1), tm)
	require.Equal(t, uint32(64*1024), mem)
	require.Equal(t, uint8(2), proc)

	valid, err := VerifyDerivedKey(pw, "hackerishere")
	require.NoError(t, err)
	require.False(t, valid)

	valid, err = VerifyDerivedKey(pw, "supersecretpassword")
	require.NoError(t, err)
	require.True(t, valid)

}
