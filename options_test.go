package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithChecks(t *testing.T) {
	h1, err := New()
	require.NoError(t, err)
	assert.Len(t, h1.checks, 0)

	h2, err := New(WithChecks(Config{
		Name: "foo",
	}, Config{
		Name: "bar",
	}))
	require.NoError(t, err)
	assert.Len(t, h2.checks, 2)

	_, err = New(WithChecks(Config{
		Name: "foo",
	}, Config{
		Name: "foo",
	}))
	require.Error(t, err)
}
