package smib

import (
	"errors"
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

// SMIB is the bot
type SMIB struct {
	slack *slack.RTM
}

// New returns a new SMIB, client must be a pointer to a valid slack.Client
func New(client *slack.Client) *SMIB {
	s := SMIB{
		slack: client.NewRTM(),
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

		s.slack.SendMessage(s.slack.NewOutgoingMessage(fmt.Sprint("hi ", user.Name), message.Channel))
	}
	return nil
}
