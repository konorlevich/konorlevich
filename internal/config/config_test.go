package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     Config
		wantErr  bool
	}{
		{name: "empty name",
			filename: "",
			want:     Config{},
			wantErr:  true,
		},
		{name: "empty file",
			filename: "test_data/empty.yaml",
			want:     Config{},
		},
		{name: "invalid file",
			filename: "test_data/invalid.yaml",
			want:     Config{},
			wantErr:  true,
		},
		{name: "valid file",
			filename: "test_data/valid.yaml",
			want: Config{
				App: Server{
					Address: "asg:13",
				},
			},
		},
		{name: "with pprof",
			filename: "test_data/with_pprof.yaml",
			want: Config{
				App: Server{
					Address: "asg:13",
				},
				Pprof: Server{
					Address: "hg:15",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
