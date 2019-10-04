package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Command runs commands for SMIB
type Command struct {
	commandDir string
}

// New creates a new Client, commandDir must be the path to a directory containing smib commands
func New(commandDir string) *Command {
	c := Command{
		commandDir: commandDir,
	}
	return &c
}

// Run takes a command and if it exists in the command diractory and is valid, runs it and
// streams the output. The caller must close the output ReadCloser if err was nil.
// User is the slack syntax for mentioning the user, userDisplay is the user's short display name.
func (c *Command) Run(command, user, userDisplay, channel, args string) (io.ReadCloser, error) {
	files, err := ioutil.ReadDir(c.commandDir)
	if err != nil {
		return nil, fmt.Errorf("error listing command directory '%s': %s", c.commandDir, err)
	}

	commands := []string{}
	var file os.FileInfo
	for _, file = range files {
		if file.IsDir() {
			continue
		}

		fileCmd := strings.SplitN(file.Name(), ".", 2)[0]

		if fileCmd == command {
			commands = []string{file.Name()}
			break
		}

		if strings.HasPrefix(fileCmd, command) {
			commands = append(commands, file.Name())
		}
	}
	if len(commands) == 0 {
		return nil, NotFoundError(fmt.Sprintf("command '%s' not found", command))
	}
	if len(commands) > 1 {
		return nil, NotUniqueError{
			text:     fmt.Sprintf("command '%s' was not unique", command),
			Commands: commands,
		}
	}

	sender := channel
	if sender == "null" {
		sender = user
	}

	log.Print(fmt.Sprintf("Command '%s' run in '%s' by '%s' with args '%s'", commands[0], channel, userDisplay, args))
	cmd := exec.Command(
		filepath.Join(c.commandDir, commands[0]),
		user,
		channel,
		sender,
		args,
		command,
		userDisplay,
	)
	cmd.Dir = c.commandDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command '%s': %s", commands[0], err)
	}

	done := make(chan struct{})
	go func() {
		<-done
		err := cmd.Wait()
		if err != nil {
			log.Print(fmt.Sprintf("Command %s failed: %s", commands[0], err))
		}
	}()

	return output{
		done:   done,
		reader: stdout,
	}, nil
}

// output is an io.ReadCloser that lets us only wait for the command to finish after the caller
// has finished reading. cmd.Wait() will close the Srdout io.ReadCloser for us.
type output struct {
	done   chan struct{}
	reader io.Reader
}

func (o output) Read(b []byte) (int, error) {
	return o.reader.Read(b)
}
func (o output) Close() error {
	close(o.done)
	return nil
}

type NotFoundError string

func (n NotFoundError) Error() string { return string(n) }

type NotUniqueError struct {
	text     string
	Commands []string
}

func (n NotUniqueError) Error() string { return n.text }

// GetCommands returns the conflicting commands as a printable string
func (n NotUniqueError) GetCommands() string {
	out := ""
	for _, cmd := range n.Commands {
		out = out + strings.SplitN(cmd, ".", 2)[0] + " "
	}
	return strings.TrimSuffix(out, " ")
}
