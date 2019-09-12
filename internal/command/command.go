package command

import (
	"fmt"
	"io/ioutil"
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

// Run takes a command and if it exists in the command diractory and is valid, runs it and returns
// the output.
func (c *Command) Run(command, user, channel, args string) ([]byte, error) {
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

		if strings.HasPrefix(strings.SplitN(file.Name(), ".", 1)[0], command) {
			commands = append(commands, file.Name())
		}
	}
	if len(commands) == 0 {
		return nil, NotFoundError(fmt.Sprintf("command '%s' not found", command))
	}
	if len(commands) > 1 {
		return nil, NotUniqeError{
			text:     fmt.Sprintf("command '%s' was not unique", command),
			Commands: commands,
		}
	}

	cmd := exec.Command(filepath.Join(c.commandDir, commands[0]), user, channel, user, args)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %s", err)
	}
	defer cmd.Wait()

	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read command output: %s", err)
	}
	return output, nil
}

type NotFoundError string

func (n NotFoundError) Error() string { return string(n) }

type NotUniqeError struct {
	text     string
	Commands []string
}

func (n NotUniqeError) Error() string { return n.text }
