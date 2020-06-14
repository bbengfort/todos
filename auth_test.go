package todos_test

import (
	"testing"

	. "github.com/bbengfort/todos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPasswordDerivedKey(t *testing.T) {
	pw, err := CreateDerivedKey("supersecretpassword")
	require.NoError(t, err)

	t.Log(pw)

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

func TestAuthTokens(t *testing.T) {
	token, err := CreateAuthToken(42)
	require.NoError(t, err)
	require.NotZero(t, token, "no token struct was returned")

	at, err := token.AccessToken()
	require.NoError(t, err)

	rt, err := token.RefreshToken()
	require.NoError(t, err)
	require.NotEqual(t, at, rt, "access and refresh tokens are identical")

	aid, err := VerifyAuthToken(at, true, false)
	require.NoError(t, err)
	require.Equal(t, token.ID, aid)

	// The refresh token will not be valid until the future
	rid, err := VerifyAuthToken(rt, false, true)
	require.Error(t, err)
	require.Equal(t, uuid.Nil, rid)
}
