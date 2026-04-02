package bootstrap

import (
	"os"
	"reflect"
	"testing"
)

func TestDefaultToServeCommand(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no args defaults to serve",
			args: []string{"app"},
			want: []string{"app", "serve"},
		},
		{
			name: "flag only invocation inserts serve before flags",
			args: []string{"app", "--dir", "pb_data", "--http=0.0.0.0:8090"},
			want: []string{"app", "serve", "--dir", "pb_data", "--http=0.0.0.0:8090"},
		},
		{
			name: "explicit subcommand is preserved",
			args: []string{"app", "migrate", "collections"},
			want: []string{"app", "migrate", "collections"},
		},
		{
			name: "help flag is preserved",
			args: []string{"app", "--help"},
			want: []string{"app", "--help"},
		},
		{
			name: "existing serve command is preserved",
			args: []string{"app", "serve", "--http=127.0.0.1:8090"},
			want: []string{"app", "serve", "--http=127.0.0.1:8090"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append([]string(nil), tt.args...)

			defaultToServeCommand()

			if !reflect.DeepEqual(os.Args, tt.want) {
				t.Fatalf("os.Args = %v, want %v", os.Args, tt.want)
			}
		})
	}
}
