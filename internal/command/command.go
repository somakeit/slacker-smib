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
// streams the output
func (c *Command) Run(command, user, channel, args string) (io.Reader, error) {
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

	log.Print(fmt.Sprintf("Command '%s' run in '%s' by '%s' with args '%s'", commands[0], channel, user, args))
	cmd := exec.Command(filepath.Join(c.commandDir, commands[0]), user, channel, user, args)
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
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Print(fmt.Sprintf("Command %s failed: %s", commands[0], err))
		}
	}()

	return stdout, nil
}

type NotFoundError string

func (n NotFoundError) Error() string { return string(n) }

type NotUniqeError struct {
	text     string
	Commands []string
}

func (n NotUniqeError) Error() string { return n.text }
