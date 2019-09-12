package smib

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/nlopes/slack"
	"github.com/somakeit/slacker-smib/internal/command"
)

// SMIB is the bot
type SMIB struct {
	slack *slack.RTM
	cmd   *command.Command
}

// New returns a new SMIB, client must be a pointer to a valid slack.Client and commandRunner
// must be a valid Smob command runner.
func New(client *slack.Client, commandRunner *command.Command) *SMIB {
	s := SMIB{
		slack: client.NewRTM(),
		cmd:   commandRunner,
	}
	return &s
}

// ListenAndRobot starts the
func (s *SMIB) ListenAndRobot() error {
	go s.slack.ManageConnection()

	for event := range s.slack.IncomingEvents {
		var err error
		switch event := event.Data.(type) {
		case *slack.MessageEvent:
			err = s.handleMessage(event)
		}
		if err != nil {
			log.Print("Faled to handle message: ", err)
		}
	}

	return errors.New("IncomingEvents channel was closed")
}

func (s *SMIB) handleMessage(message *slack.MessageEvent) error {
	if message.Text[0] == '?' {
		s.slack.SendMessage(s.slack.NewTypingMessage(message.Channel))

		user, err := s.slack.GetUserInfo(message.User)
		if err != nil {
			return fmt.Errorf("failed to get user info: %s", err)
		}

		parts := strings.SplitN(message.Text, " ", 2)
		cmd := strings.TrimPrefix(parts[0], "?")
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		channel, err := s.slack.GetChannelInfo(message.Channel)
		if err != nil {
			// TODO handle DMs
			return fmt.Errorf("failed to get channel info: %s", err)
		}

		output, err := s.cmd.Run(
			cmd,
			user.Name,
			channel.Name,
			args,
		)
		// TODO handle command not found etc.
		if err != nil {
			return err
		}

		s.slack.SendMessage(s.slack.NewOutgoingMessage(string(output), message.Channel))
	}
	return nil
}
