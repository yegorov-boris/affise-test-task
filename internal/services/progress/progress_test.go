package progress

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func TestState_Cancel(t *testing.T) {
	s := new(State)
	id, ctx := s.Start()
	tests := []struct {
		name string
		id   uint64
		want bool
	}{
		{
			name: "should return false when ID is not found",
			id:   0,
			want: false,
		},
		{
			name: "should return true and call cancel func when ID is found",
			id:   id,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Cancel(tt.id); got != tt.want {
				t.Errorf("Cancel() = %v, want %v", got, tt.want)
			}
			if !tt.want && ctx.Err() != nil {
				t.Error("cancel function should not be called when ID is not found")
			}
			if tt.want && ctx.Err() != context.Canceled {
				t.Error("cancel function should be called when ID is found")
			}
		})
	}
}

func TestNew(t *testing.T) {
	id1 := rand.Uint64()
	id2 := rand.Uint64()
	storePath := "./store/"
	tests := []struct {
		name      string
		storePath string
		wantID    uint64
		wantErr   bool
	}{
		{
			name:      "should start from 1 when the store directory is empty",
			storePath: storePath,
			wantID:    1,
			wantErr:   false,
		},
		{
			name:      "should start from max ID + 1 when the store directory is not empty",
			storePath: storePath,
			wantID:    max(id1, id2) + 1,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		//set up
		if err := os.RemoveAll(storePath); err != nil && err != os.ErrExist {
			t.Errorf("rm %q failed: %s", storePath, err)
		}
		if err := os.Mkdir(storePath, 0777); err != nil {
			t.Errorf("mkdir %q failed: %s", storePath, err)
		}
		if tt.wantID > 1 {
			for _, id := range []uint64{id1, id2} {
				name := fmt.Sprintf("%s%d.json", storePath, id)
				_, err := os.Create(name)
				if err != nil {
					t.Errorf("failed to create %q: %s", name, err)
				}
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			state, err := New(tt.storePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			id, _ := state.Start()
			if id != tt.wantID {
				t.Errorf("Start() got = %v, want %v", id, tt.wantID)
			}
		})
	}
	if err := os.RemoveAll(storePath); err != nil && err != os.ErrExist {
		t.Errorf("rm %q failed: %s", storePath, err)
	}
}
