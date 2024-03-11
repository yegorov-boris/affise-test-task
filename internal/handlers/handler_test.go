package handlers

import (
	"fmt"
	"math/rand"
	"testing"
)

func Test_parseID(t *testing.T) {
	type args struct {
		basePath string
		path     string
	}
	validID := rand.Uint64()
	basePath := "/api/v1/links"
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "should parse a valid ID",
			args: args{
				basePath: basePath,
				path:     fmt.Sprintf("%s/%d", basePath, validID),
			},
			want:    validID,
			wantErr: false,
		},
		{
			name: "should fail for an empty ID",
			args: args{
				basePath: basePath,
				path:     fmt.Sprintf("%s/", basePath),
			},
			wantErr: true,
		},
		{
			name: "should fail for an incorrect ID",
			args: args{
				basePath: basePath,
				path:     fmt.Sprintf("%s/foo", basePath),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseID(tt.args.basePath, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
