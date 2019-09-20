package smib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/nlopes/slack"
	"github.com/somakeit/slacker-smib/internal/command"
)

type commandRunner interface {
	Run(cmd, user, channel, args string) (io.ReadCloser, error)
}

// SMIB is the bot
type SMIB struct {
	slack *slack.RTM
	cmd   commandRunner
}

// New returns a new SMIB, client must be a pointer to a valid slack.Client and commandRunner
// must be a valid Smob command runner.
func New(client *slack.Client, cmd commandRunner) *SMIB {
	s := SMIB{
		slack: client.NewRTM(),
		cmd:   cmd,
	}
	return &s
}

// ListenAndRobot starts the
func (s *SMIB) ListenAndRobot() error {
	go s.slack.ManageConnection()

	for event := range s.slack.IncomingEvents {
		switch data := event.Data.(type) {
		case *slack.MessageEvent:
			go func() {
				if err := s.handleMessage(data); err != nil {
					log.Print("Faled to handle message: ", err)
				}
			}()
		case *slack.ConnectedEvent:
			log.Println("SMIB connected")
		}
	}

	return errors.New("IncomingEvents channel was closed")
}

func (s *SMIB) handleMessage(message *slack.MessageEvent) error {
	if len(message.Text) < 1 || message.Text[0] != '?' {
		return nil
	}

	parts := strings.SplitN(message.Text, " ", 2)
	cmd := strings.TrimPrefix(parts[0], "?")
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	if len(cmd) < 1 {
		return nil
	}

	s.slack.SendMessage(s.slack.NewTypingMessage(message.Channel))

	user, err := s.slack.GetUserInfo(message.User)
	if err != nil {
		return fmt.Errorf("failed to get user info: %s", err)
	}

	channel, err := s.slack.GetChannelInfo(message.Channel)
	if err != nil {
		// TODO handle DMs
		return fmt.Errorf("failed to get channel info: %s", err)
	}

	var msgOpts []slack.RTMsgOption
	if message.ThreadTimestamp != "" {
		msgOpts = append(msgOpts, slack.RTMsgOptionTS(message.ThreadTimestamp))
	}

	output, err := s.cmd.Run(
		cmd,
		user.Name,
		channel.Name,
		args,
	)
	switch err := err.(type) {
	case nil:
		break
	case command.NotFoundError:
		s.slack.SendMessage(s.slack.NewOutgoingMessage(
			fmt.Sprintf("Sorry %s, I don't have a %s command.", user.Name, cmd),
			message.Channel,
			msgOpts...,
		))
		return nil
	case command.NotUniqueError:
		s.slack.SendMessage(s.slack.NewOutgoingMessage(
			fmt.Sprintf("Sorry %s, that wasn't unique, try one of: %s", user.Name, err.GetCommands()),
			message.Channel,
			msgOpts...,
		))
		return nil
	default:
		s.slack.SendMessage(s.slack.NewOutgoingMessage(
			fmt.Sprintf("Sorry %s, %s is on fire.", user.Name, cmd),
			message.Channel,
			msgOpts...,
		))
		return err
	}

	reader := bufio.NewReader(output)
	for {
		out, err := reader.ReadString('\n')
		if len(out) > 0 {
			s.slack.SendMessage(s.slack.NewOutgoingMessage(
				out,
				message.Channel,
				msgOpts...,
			))
		}
		switch err {
		case nil:
			continue
		case io.EOF:
			output.Close()
			return nil
		default:
			s.slack.SendMessage(s.slack.NewOutgoingMessage(
				fmt.Sprintf("Sorry %s, %s exploded or something.", user.Name, cmd),
				message.Channel,
				msgOpts...,
			))
			return fmt.Errorf("failed to read output from command: %s", err)
		}
	}
}
