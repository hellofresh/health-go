package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitorsalgado/mocha/v2"
	"github.com/vitorsalgado/mocha/v2/expect"
	"github.com/vitorsalgado/mocha/v2/reply"
)

func TestNew(t *testing.T) {
	m := mocha.New(t)
	m.Start()
	m.CloseOnCleanup(t)

	t.Run("service is available", func(t *testing.T) {
		svc200 := m.AddMocks(mocha.Get(expect.URLPath("/test-200")).Reply(reply.OK()))

		check := New(Config{
			URL: m.URL() + "/test-200",
		})

		err := check(context.Background())
		require.NoError(t, err)
		assert.True(t, svc200.Called())
	})

	t.Run("service is not available", func(t *testing.T) {
		svc500 := m.AddMocks(mocha.Get(expect.URLPath("/test-500")).Reply(reply.InternalServerError()))

		check := New(Config{
			URL: m.URL() + "/test-500",
		})

		err := check(context.Background())
		require.Error(t, err)
		assert.True(t, svc500.Called())
	})
}
