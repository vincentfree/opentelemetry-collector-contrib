package headerprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/headerprocessor/internal/metadata"
)

func TestCreateDefaultConfig(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig()
	if _, ok := cfg.(*Config); !ok {
		t.Fatalf("unexpected config type: %T", cfg)
	}
}

func TestFactoryConstructors(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig()
	settings := processortest.NewNopSettings(metadata.Type)

	// With a non-nil next consumer the processors should be created without error
	_, err := f.CreateTraces(t.Context(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	_, err = f.CreateLogs(t.Context(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	_, err = f.CreateMetrics(t.Context(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)
}
