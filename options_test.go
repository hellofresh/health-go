package health

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
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

type mockTracerProvider struct {
	mock.Mock
}

func (m *mockTracerProvider) Tracer(instrumentationName string, opts ...trace.TracerOption) trace.Tracer {
	args := m.Called(instrumentationName, opts)
	return args.Get(0).(trace.Tracer)
}

func TestWithTracerProvider(t *testing.T) {
	h1, err := New()
	require.NoError(t, err)
	assert.Equal(t, "trace.noopTracerProvider", fmt.Sprintf("%T", h1.tp))
	assert.Equal(t, "", h1.instrumentationName)

	tp := new(mockTracerProvider)
	instrumentationName := "test.test"

	h2, err := New(WithTracerProvider(tp, instrumentationName))
	require.NoError(t, err)
	assert.Same(t, tp, h2.tp)
	assert.Equal(t, instrumentationName, h2.instrumentationName)
}

func TestWithComponent(t *testing.T) {
	h1, err := New()
	require.NoError(t, err)
	assert.Empty(t, h1.component.Name)
	assert.Empty(t, h1.component.Version)

	c := new(Component)
	c.Name = "test"
	c.Version = "1.0"

	h2, err := New(WithComponent(*c))
	require.NoError(t, err)
	assert.Equal(t, "test", h2.component.Name)
	assert.Equal(t, "1.0", h2.component.Version)
}
