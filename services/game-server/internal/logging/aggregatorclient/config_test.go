package aggregatorclient

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config != DefaultConfig() {
		t.Fatalf("config = %#v, want defaults %#v", config, DefaultConfig())
	}
	if !config.SpoolEnabled {
		t.Fatal("default config must enable spooling")
	}
}

func TestSpoolCanBeDisabledByEnvironment(t *testing.T) {
	t.Setenv("OBS_AGGREGATOR_SPOOL_ENABLED", "false")
	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.SpoolEnabled {
		t.Fatal("spooling unexpectedly enabled")
	}
}

func TestDisabledConfigAllowsMissingEndpoint(t *testing.T) {
	t.Setenv("OBS_AGGREGATOR_ENABLED", "false")
	t.Setenv("OBS_AGGREGATOR_ENDPOINT_URL", "")
	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.Enabled {
		t.Fatal("config unexpectedly enabled")
	}
}

func TestValidEnabledConfig(t *testing.T) {
	t.Setenv("OBS_AGGREGATOR_ENABLED", "true")
	t.Setenv("OBS_AGGREGATOR_ENDPOINT_URL", "https://aggregator.example.test/v1/events")
	t.Setenv("OBS_AGGREGATOR_QUEUE_CAPACITY", "32")
	t.Setenv("OBS_AGGREGATOR_BATCH_SIZE", "8")
	t.Setenv("OBS_AGGREGATOR_FLUSH_INTERVAL", "250ms")
	t.Setenv("OBS_AGGREGATOR_REQUEST_TIMEOUT", "2s")
	t.Setenv("OBS_AGGREGATOR_SPOOL_BYTE_CAP", "4096")
	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.QueueCapacity != 32 || config.BatchSize != 8 || config.FlushInterval != 250*time.Millisecond || config.RequestTimeout != 2*time.Second || config.SpoolByteCap != 4096 {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestInvalidURL(t *testing.T) {
	t.Setenv("OBS_AGGREGATOR_ENABLED", "true")
	t.Setenv("OBS_AGGREGATOR_ENDPOINT_URL", "not-a-url")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("expected invalid URL error")
	}
}

func TestInvalidNumericAndDurationValues(t *testing.T) {
	cases := []struct{ name, env, value string }{
		{"numeric", "OBS_AGGREGATOR_BATCH_SIZE", "many"},
		{"duration", "OBS_AGGREGATOR_FLUSH_INTERVAL", "soon"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(testCase.env, testCase.value)
			if _, err := ConfigFromEnv(); err == nil {
				t.Fatalf("expected error for %s", testCase.env)
			}
		})
	}
}
