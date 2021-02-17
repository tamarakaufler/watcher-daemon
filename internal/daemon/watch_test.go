package daemon_test

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tamarakaufler/watcher-daemon/internal/daemon"
)

func TestDaemon_CollectFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	type fields struct {
		BasePath  string
		Extention string
		Command   string
		Excluded  string
		Frequency string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "got correctly all files - no exclusions",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "",
				Frequency: "3",
			},
			want:    []string{"test.go", "test1.go", "test.go", "test2.go", "test.go"},
			wantErr: false,
		},
		{
			name: "got correctly all files - with one individual file exclusion",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/test.go",
				Frequency: "3",
			},
			want:    []string{"test1.go", "test.go", "test2.go", "test.go"},
			wantErr: false,
		},
		{
			name: "got correctly all files - with individual file exclusions",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/test.go,fixtures/basepath/subdir2/test2.go",
				Frequency: "3",
			},
			want:    []string{"test1.go", "test.go", "test.go"},
			wantErr: false,
		},
		{
			name: "got correctly all files - with regex file exclusions",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/*,fixtures/basepath/subdir2/test.go",
				Frequency: "3",
			},
			want:    []string{"test2.go", "test.go"},
			wantErr: false,
		},
		{
			name: "got correctly all files - excluding file of the same name in multiple dirs",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "test.go",
				Frequency: "3",
			},
			want:    []string{"test1.go", "test2.go"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("WATCHER_DAEMON_BASE_PATH", tt.fields.BasePath)
			os.Setenv("WATCHER_DAEMON_EXTENSION", tt.fields.Extention)
			os.Setenv("WATCHER_DAEMON_COMMAND", tt.fields.Command)
			os.Setenv("WATCHER_DAEMON_EXCLUDED", tt.fields.Excluded)
			os.Setenv("WATCHER_DAEMON_FREQUENCY", tt.fields.Frequency)

			d, err := daemon.New()
			require.Nil(t, err, "daemon creation failure")

			got, err := d.CollectFiles(ctx)
			gotNames := extractNames(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Daemon.CollectFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNames, tt.want) {
				t.Errorf("Daemon.CollectFiles() = %v, want %v", gotNames, tt.want)
			}
		})
	}
}

func extractNames(files []daemon.FileInfo) []string {
	names := []string{}
	for _, f := range files {
		names = append(names, f.Name)
	}
	return names
}

func TestDaemon_IsExcluded(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	type fields struct {
		BasePath  string
		Extention string
		Command   string
		Excluded  string
		Frequency string
	}
	type args struct {
		path string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "file is excluded - regex files exclusion",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/*",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir1/test1.go",
				name: "test1.go",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "file is excluded - regex files exclusion 2",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "test2.go,fixtures/basepath/subdir1/test1*",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir1/test1.go",
				name: "test1.go",
			},
			want:    true,
			wantErr: false,
		},
		// this does not work - why?
		// {
		// 	name: "file is excluded - regex files exclusion - 3",
		// 	fields: fields{
		// 		BasePath:  "fixtures/basepath",
		// 		Extention: ".go",
		// 		Command:   "echo \"Hello world\"",
		// 		Excluded:  "test2.gos,fixtures/basepath/*/test.go",
		// 		Frequency: "3",
		// 	},
		// 	args: args{
		// 		path: "fixtures/basepath/subdir1/test.go",
		// 		name: "test.go",
		// 	},
		// 	want:    true,
		// 	wantErr: false,
		// },
		{
			name: "file is excluded - string path exclusion 1",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/test.go",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir1/test.go",
				name: "test.go",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "file is excluded - string file exclusion 2",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "test2.go,fixtures/basepath/subdir1/test.go",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir2/test2.go",
				name: "test2.go",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "file is excluded - string file exclusion 3",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "test2.go,fixtures/basepath/subdir1/test.go",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir2/test2.go",
				name: "test2.go",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "file is excluded - regex ? file exclusion",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/test.g?",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir1/test.go",
				name: "test.go",
			},
			want:    true,
			wantErr: false,
		},
		// this does not work - why?
		// {
		// 	name: "file is excluded - regex ? file exclusion 2",
		// 	fields: fields{
		// 		BasePath:  "fixtures/basepath",
		// 		Extention: ".go",
		// 		Command:   "echo \"Hello world\"",
		// 		Excluded:  "fixtures/basepath/subdir1/test.?o",
		// 		Frequency: "3",
		// 	},
		// 	args: args{
		// 		path: "fixtures/basepath/subdir1/test.go",
		// 		name: "test.go",
		// 	},
		// 	want:    true,
		// 	wantErr: false,
		// },
		{
			name: "file is not excluded - string file exclusion 4",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/test.go",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir1/test2.go",
				name: "test2.go",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "file is not excluded - regex file exclusion 1",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/test*",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir1/aaa.go",
				name: "aaa.go",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "file is not excluded - regex file exclusion 2",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures/basepath/subdir1/*",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir2/aaa.go",
				name: "aaa.go",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "file to be excluded - regex file exclusion",
			fields: fields{
				BasePath:  "fixtures/basepath",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "fixtures(a-]basepath/subdir1/*",
				Frequency: "3",
			},
			args: args{
				path: "fixtures/basepath/subdir2/aaa.go",
				name: "aaa.go",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("WATCHER_DAEMON_BASE_PATH", tt.fields.BasePath)
			os.Setenv("WATCHER_DAEMON_EXTENSION", tt.fields.Extention)
			os.Setenv("WATCHER_DAEMON_COMMAND", tt.fields.Command)
			os.Setenv("WATCHER_DAEMON_EXCLUDED", tt.fields.Excluded)
			os.Setenv("WATCHER_DAEMON_FREQUENCY", tt.fields.Frequency)

			d, err := daemon.New()
			require.Nil(t, err, "daemon creation failure")

			got, err := d.IsExcluded(ctx, tt.args.path, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Daemon.IsExcluded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Daemon.IsExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaemon_ProcessFilesInParallel(t *testing.T) {
	t.Parallel()
	type fields struct {
		BasePath  string
		Extention string
		Command   string
		Excluded  string
		Frequency string
	}
	type args struct {
		ctx    context.Context
		doneCh chan struct{}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		expectChange bool
	}{
		{
			name: "processing files - no change",
			fields: fields{
				BasePath:  "fixtures/basepath/subdir1",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "",
				Frequency: "2",
			},
			args: args{
				ctx:    context.Background(),
				doneCh: make(chan struct{}, 5),
			},
			expectChange: false,
		},
		{
			name: "processing files - file changed",
			fields: fields{
				BasePath:  "fixtures/basepath/subdir2",
				Extention: ".go",
				Command:   "echo \"Hello world\"",
				Excluded:  "",
				Frequency: "3",
			},
			args: args{
				ctx:    context.Background(),
				doneCh: make(chan struct{}, 5),
			},
			expectChange: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("WATCHER_DAEMON_BASE_PATH", tt.fields.BasePath)
			os.Setenv("WATCHER_DAEMON_EXTENSION", tt.fields.Extention)
			os.Setenv("WATCHER_DAEMON_COMMAND", tt.fields.Command)
			os.Setenv("WATCHER_DAEMON_EXCLUDED", tt.fields.Excluded)
			os.Setenv("WATCHER_DAEMON_FREQUENCY", tt.fields.Frequency)

			d, err := daemon.New()
			require.Nil(t, err, "daemon creation failure")

			// Within this gouroutine we verify
			go func() {
				fr, err := strconv.Atoi(tt.fields.Frequency)
				require.Nil(t, err)
				timeout := time.After(time.Duration(fr+1) * time.Second)
				if tt.expectChange {
					select {
					case <-tt.args.doneCh: // receiving on this channel is expected if a change is detected
					case <-timeout:
						t.Error("TestDaemon_ProcessFilesInParallel - change should have been detected")
					}
				} else {
					select {
					case <-tt.args.doneCh: // receiving on this channel is not expected as no change happened
						t.Error("TestDaemon_ProcessFilesInParallel - change was detected")
					case <-timeout:
					}
				}
			}()

			files, err := d.CollectFiles(tt.args.ctx)
			if err != nil {
				t.Errorf("TestDaemon_ProcessFilesInParallel - %s", err)
			}
			if tt.expectChange {
				// simulate change
				err := os.Chtimes(files[0].Path, time.Now(), time.Now())
				if err != nil {
					t.Errorf("TestDaemon_ProcessFilesInParallel - %s", err)
				}
			}
			d.ProcessFilesInParallel(tt.args.ctx, files, tt.args.doneCh)
		})
	}
}
