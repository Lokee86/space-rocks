package servicelog

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	validFile := FilePolicy{
		Directory:          "logs",
		Prefix:             "game-server",
		SegmentMaxBytes:    1024,
		SegmentMaxAge:      time.Hour,
		RetentionMaxAge:    24 * time.Hour,
		RetentionMaxBytes:  4096,
		CompressionEnabled: true,
	}

	tests := []struct {
		name      string
		config    Config
		wantValid bool
	}{
		{
			name: "valid console only",
			config: Config{
				Identity:       ServiceIdentity{Name: "game-server"},
				ConsoleEnabled: true,
			},
			wantValid: true,
		},
		{
			name: "valid file enabled",
			config: Config{
				Identity:    ServiceIdentity{Name: "game-server", Version: "dev"},
				File:        validFile,
				FileEnabled: true,
			},
			wantValid: true,
		},
		{
			name: "invalid identity",
			config: Config{
				ConsoleEnabled: true,
			},
		},
		{
			name: "invalid negative flush interval",
			config: Config{
				Identity: ServiceIdentity{Name: "game-server"},
				Flush:    FlushPolicy{Interval: -time.Second},
			},
		},
		{
			name: "invalid segment max bytes",
			config: Config{
				Identity:    ServiceIdentity{Name: "game-server"},
				File:        FilePolicy{Directory: "logs", Prefix: "game-server", SegmentMaxAge: time.Hour, RetentionMaxAge: time.Hour, RetentionMaxBytes: 1},
				FileEnabled: true,
			},
		},
		{
			name: "invalid segment max age",
			config: Config{
				Identity:    ServiceIdentity{Name: "game-server"},
				File:        FilePolicy{Directory: "logs", Prefix: "game-server", SegmentMaxBytes: 1, RetentionMaxAge: time.Hour, RetentionMaxBytes: 1},
				FileEnabled: true,
			},
		},
		{
			name: "invalid retention max age",
			config: Config{
				Identity:    ServiceIdentity{Name: "game-server"},
				File:        FilePolicy{Directory: "logs", Prefix: "game-server", SegmentMaxBytes: 1, SegmentMaxAge: time.Hour, RetentionMaxBytes: 1},
				FileEnabled: true,
			},
		},
		{
			name: "invalid retention max bytes",
			config: Config{
				Identity:    ServiceIdentity{Name: "game-server"},
				File:        FilePolicy{Directory: "logs", Prefix: "game-server", SegmentMaxBytes: 1, SegmentMaxAge: time.Hour, RetentionMaxAge: time.Hour},
				FileEnabled: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if (err == nil) != test.wantValid {
				t.Fatalf("Validate() error = %v, want valid = %v", err, test.wantValid)
			}
		})
	}
}
