package health

import (
	"fmt"
	"runtime"
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

	c := Component{
		Name:    "test",
		Version: "1.0",
	}

	h2, err := New(WithComponent(c))
	require.NoError(t, err)
	assert.Equal(t, "test", h2.component.Name)
	assert.Equal(t, "1.0", h2.component.Version)
}

func TestWithMaxConcurrent(t *testing.T) {
	numCPU := runtime.NumCPU()
	t.Logf("Num CPUs: %d", numCPU)

	h1, err := New()
	require.NoError(t, err)
	assert.Equal(t, numCPU, h1.maxConcurrent)

	h2, err := New(WithMaxConcurrent(13))
	require.NoError(t, err)
	assert.Equal(t, 13, h2.maxConcurrent)
}

func TestWithSystemInfo(t *testing.T) {
	h1, err := New()
	require.NoError(t, err)
	assert.False(t, h1.systemInfoEnabled)

	h2, err := New(WithSystemInfo())
	require.NoError(t, err)
	assert.True(t, h2.systemInfoEnabled)
}
