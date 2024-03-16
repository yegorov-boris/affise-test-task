package dotenv

import (
	"os"
	"strings"
	"testing"
)

func Test_parse(t *testing.T) {
	s := []string{
		"FOO=bar",
		"ABC_DEF=1",
		"",
		"TEST=2s",
	}
	expected := map[string]string{
		"FOO":     "bar",
		"ABC_DEF": "1",
		"TEST":    "2s",
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should parse a simplified .env file (see documentation for Load function)",
			args: args{
				b: []byte(strings.Join(s, "\n")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parse(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			for key, value := range expected {
				if os.Getenv(key) != value {
					t.Errorf("Wrong env var %q: expected %q - got %q", key, expected[key], value)
				}
			}
		})
	}
}
