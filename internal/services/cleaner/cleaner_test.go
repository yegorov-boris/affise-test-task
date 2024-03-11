package cleaner

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	type args struct {
		storeTimeout time.Duration
		storePath    string
		logger       *slog.Logger
	}
	storePath := "./store/"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should clean outdated files",
			args: args{
				storeTimeout: 2 * time.Second,
				storePath:    storePath,
				logger:       slog.New(slog.NewTextHandler(os.Stdout, nil)),
			},
			want: "2.json",
		},
	}
	for _, tt := range tests {
		//set up
		if err := os.Mkdir(storePath, 0777); err != nil {
			t.Errorf("mkdir %q failed: %s", storePath, err)
		}

		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.args.storeTimeout, tt.args.storePath, tt.args.logger)

			for _, id := range []string{"1", "2"} {
				name := fmt.Sprintf("%s%s.json", storePath, id)
				_, err := os.Create(name)
				if err != nil {
					t.Errorf("failed to create %q: %s", name, err)
				}
				time.Sleep(time.Second)
			}
			time.Sleep(100 * time.Millisecond)

			c.Shutdown()
			time.Sleep(2100 * time.Millisecond)

			entries, err := os.ReadDir(c.storePath)
			if err != nil {
				t.Errorf("failed to list dir %q: %s", storePath, err)
			}
			if len(entries) != 1 || entries[0].Name() != tt.want {
				actualFileNames := []string{}
				for _, e := range entries {
					actualFileNames = append(actualFileNames, e.Name())
				}
				t.Errorf("Expected the only file %q after cleaning - got %v", tt.want, actualFileNames)
			}
		})
	}
	if err := os.RemoveAll(storePath); err != nil && err != os.ErrExist {
		t.Errorf("rm %q failed: %s", storePath, err)
	}
}
