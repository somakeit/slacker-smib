package smib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slacktest"
	"github.com/somakeit/slacker-smib/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCommand struct {
	mock.Mock
}

func (m *mockCommand) Run(cmd, user, channel, args string) (io.Reader, error) {
	mArgs := m.Called(cmd, user, channel, args)
	return mArgs.Get(0).(io.Reader), mArgs.Error(1)
}

func TestNew(t *testing.T) {
	client := slack.New("xoxb-whatever")
	cmd := &mockCommand{}
	smib := New(client, cmd)
	assert.IsType(t, &slack.RTM{}, smib.slack)
	assert.Same(t, cmd, smib.cmd)
}

func TestListenAndRobot(t *testing.T) {
	testServer := slacktest.NewTestServer()
	testServer.SetBotName("smib")
	testServer.Handle("/channels.info", func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "Xgeneral", r.Form["channel"][0], "Need to paramaterise this mock")
		resp, _ := json.Marshal(struct{ Channel slack.Channel }{slack.Channel{GroupConversation: slack.GroupConversation{Name: "general"}}})
		w.Write(resp)
	})
	testServer.Start()
	testRTM := testServer.GetTestRTMInstance()

	mockCmd := &mockCommand{}
	mockCmd.Test(t)
	empty := bytes.NewReader(nil)
	mockCmd.On("Run", "command", "spengler", "general", "arg arg").Return(empty, errors.New("woteva")).Once()
	defer mockCmd.AssertExpectations(t)

	smib := SMIB{
		slack: testRTM,
		cmd:   mockCmd,
	}

	done := make(chan struct{})
	var err error
	go func() {
		err = smib.ListenAndRobot()
		close(done)
	}()

	testRTM.IncomingEvents <- slack.RTMEvent{
		Type: "typing",
		Data: &slack.MessageEvent{
			Msg: slack.Msg{
				Text:    "?command arg arg",
				Channel: "Xgeneral",
			},
		},
	}

	// allow time to connect an process the message before closing the channel
	time.Sleep(time.Millisecond * 100)

	close(testRTM.IncomingEvents)
	testServer.Stop()
	<-done
	assert.EqualError(t, err, "IncomingEvents channel was closed")
}

func TestSMIB_handleMessage(t *testing.T) {
	tests := []struct {
		name         string
		message      *slack.MessageEvent
		primeCommand func(*testing.T, *mockCommand)
		wantMessage  string
		wantErr      string
	}{
		{
			name: "not a command",
			message: &slack.MessageEvent{
				Msg: slack.Msg{
					Text:    "some random chatter",
					User:    "Xspengler",
					Channel: "Xgeneral",
				},
			},
		},
		{
			name: "zero length message",
			message: &slack.MessageEvent{
				Msg: slack.Msg{
					Text:    "",
					User:    "Xspengler",
					Channel: "Xgeneral",
				},
			},
		},
		{
			name: "a command",
			message: &slack.MessageEvent{
				Msg: slack.Msg{
					Text:    "?command y0",
					User:    "Xspengler",
					Channel: "Xgeneral",
				},
			},
			primeCommand: func(t *testing.T, m *mockCommand) {
				cmdReader := bytes.NewReader([]byte("computer says yes"))
				t.Log("here")
				m.On("Run", "command", "spengler", "general", "y0").Return(cmdReader, nil).Once()
			},
			wantMessage: "computer says yes",
		},
		{
			name: "unknown command",
			message: &slack.MessageEvent{
				Msg: slack.Msg{
					Text:    "?badcommand",
					User:    "Xspengler",
					Channel: "Xgeneral",
				},
			},
			primeCommand: func(t *testing.T, m *mockCommand) {
				empty := bytes.NewReader(nil)
				m.On("Run", "badcommand", "spengler", "general", "").Return(empty, command.NotFoundError("")).Once()
			},
			wantMessage: "Sorry spengler, I don't have a badcommand command.",
		},
		{
			name: "nonunique command",
			message: &slack.MessageEvent{
				Msg: slack.Msg{
					Text:    "?c",
					User:    "Xspengler",
					Channel: "Xgeneral",
				},
			},
			primeCommand: func(t *testing.T, m *mockCommand) {
				empty := bytes.NewReader(nil)
				m.On("Run", "c", "spengler", "general", "").Return(
					empty,
					command.NotUniqueError{
						Commands: []string{"commands", "countdown"},
					},
				).Once()
			},
			wantMessage: "Sorry spengler, that wasn't unique, try one of: commands countdown",
		},
		{
			name: "error running command",
			message: &slack.MessageEvent{
				Msg: slack.Msg{
					Text:    "?crash",
					User:    "Xspengler",
					Channel: "Xgeneral",
				},
			},
			primeCommand: func(t *testing.T, m *mockCommand) {
				empty := bytes.NewReader(nil)
				m.On("Run", "crash", "spengler", "general", "").Return(empty, errors.New("oops")).Once()
			},
			wantMessage: "Sorry spengler, crash is on fire.",
			wantErr:     "oops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := slacktest.NewTestServer()
			testServer.SetBotName("smib")
			testServer.Handle("/channels.info", func(w http.ResponseWriter, r *http.Request) {
				assert.NoError(t, r.ParseForm())
				assert.Equal(t, "Xgeneral", r.Form["channel"][0], "Need to paramaterise this mock")
				resp, _ := json.Marshal(struct{ Channel slack.Channel }{slack.Channel{GroupConversation: slack.GroupConversation{Name: "general"}}})
				w.Write(resp)
			})
			testServer.Start()
			testRTM := testServer.GetTestRTMInstance()
			go testRTM.ManageConnection()

			mockCmd := &mockCommand{}
			mockCmd.Test(t)
			if tt.primeCommand != nil {
				tt.primeCommand(t, mockCmd)
			}
			defer mockCmd.AssertExpectations(t)

			smib := SMIB{
				slack: testRTM,
				cmd:   mockCmd,
			}

			err := smib.handleMessage(tt.message)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.wantErr)
			}

			// handleMessage has returned but the testServer needs time to receive its messages
			time.Sleep(time.Millisecond * 10)
			testServer.Stop()

			t.Log(testServer.GetSeenInboundMessages())
			if tt.wantMessage != "" {
				assert.True(t, sawTypingMessage(testServer), "Typing message not sent")
				assert.True(t, testServer.SawMessage(tt.wantMessage), "Message '%s' not seen", tt.wantMessage)
			} else {
				assert.Empty(t, testServer.GetSeenInboundMessages())
			}
		})
	}
}

func sawTypingMessage(server *slacktest.Server) bool {
	for _, msg := range server.GetSeenInboundMessages() {
		typing := slack.UserTypingEvent{}
		if err := json.Unmarshal([]byte(msg), &typing); err != nil {
			continue
		}
		if typing.Type != "typing" {
			continue
		}
		return true
	}
	return false
}
