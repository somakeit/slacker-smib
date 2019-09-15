package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	expected := &Command{
		commandDir: "/some/dir",
	}
	actual := New("/some/dir")
	assert.Equal(t, expected, actual)
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}

func TestCommand_Run(t *testing.T) {
	tests := []struct {
		name       string
		commandDir string
		command    string
		user       string
		channel    string
		args       string
		want       []byte
		wantErr    error
	}{
		{
			name:       "invalid command dir",
			commandDir: "notadir",
			command:    "lol",
			user:       "bob",
			channel:    "general",
			want:       []byte{},
			wantErr:    errors.New("error listing command directory 'notadir':"),
		},
		{
			name:       "command not found",
			commandDir: mustAbs("fixtures"),
			command:    "notacmd",
			user:       "bob",
			channel:    "general",
			want:       []byte{},
			wantErr:    NotFoundError("command 'notacmd' not found"),
		},
		{
			name:       "command not unique",
			commandDir: mustAbs("fixtures"),
			command:    "command",
			user:       "bob",
			channel:    "general",
			want:       []byte{},
			wantErr: NotUniqueError{
				text:     "command 'command' was not unique",
				Commands: []string{"commandone.sh", "commandtwo.sh"},
			},
		},
		{
			name:       "run a command",
			commandDir: mustAbs("fixtures"),
			command:    "commandone",
			user:       "bob",
			channel:    "general",
			want:       []byte("command one\n"),
			wantErr:    nil,
		},
		{
			name:       "run a unique command",
			commandDir: mustAbs("fixtures"),
			command:    "commandt",
			user:       "bob",
			channel:    "general",
			want:       []byte("command two\n"),
			wantErr:    nil,
		},
		{
			name:       "run the readme",
			commandDir: mustAbs("fixtures"),
			command:    "README",
			user:       "bob",
			channel:    "general",
			want:       []byte{},
			wantErr:    fmt.Errorf("failed to start command 'README.md':"),
		},
		{
			name:       "run a failing command",
			commandDir: mustAbs("fixtures"),
			command:    "fail",
			user:       "bob",
			channel:    "general",
			want:       []byte("i bad\n"),
			wantErr:    nil,
		},
		{
			name:       "run a command with args",
			commandDir: mustAbs("fixtures"),
			command:    "debug",
			user:       "bob",
			channel:    "general",
			args:       "some args",
			want:       []byte("bob, User: [bob] Channel: [general] Sender: [bob] Args: [some args]\n"),
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{
				commandDir: tt.commandDir,
			}

			r, outErr := c.Run(tt.command, tt.user, tt.channel, tt.args)

			switch wantErr := tt.wantErr.(type) {
			case nil:
				assert.NoError(t, outErr)
			case NotFoundError:
				assert.IsType(t, wantErr, outErr)
				assert.Contains(t, outErr.Error(), tt.wantErr.Error())
			case NotUniqueError:
				assert.IsType(t, wantErr, outErr)
				assert.Equal(t, wantErr.Commands, outErr.(NotUniqueError).Commands)
				assert.Contains(t, outErr.Error(), tt.wantErr.Error())
			default:
				assert.Contains(t, outErr.Error(), tt.wantErr.Error())
			}

			output := []byte{}
			if r != nil {
				var err error
				output, err = ioutil.ReadAll(r)
				require.NoError(t, err)
			}
			if outErr == nil {
				r.Close()
			}
			assert.Equal(t, tt.want, output)
		})
	}
}

func TestNotUniqueError_GetCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		want     string
	}{
		{
			name:     "Two similar commands",
			commands: []string{"cointoss.sh", "countdown.sh"},
			want:     "cointoss countdown",
		},
		{
			name:     "No commands?",
			commands: []string{},
			want:     "",
		},
		{
			name:     "One command?",
			commands: []string{"lol"},
			want:     "lol",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NotUniqueError{
				text:     tt.name,
				Commands: tt.commands,
			}
			if got := n.GetCommands(); got != tt.want {
				t.Errorf("NotUniqueError.Commands() = %v, want %v", got, tt.want)
			}
		})
	}
}
