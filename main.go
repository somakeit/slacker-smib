package main

import (
	"flag"
	"log"

	"github.com/nlopes/slack"
	"github.com/somakeit/slacker-smib/internal/command"
	"github.com/somakeit/slacker-smib/internal/smib"
)

func main() {
	var (
		token      string
		commandDir string
	)

	flag.StringVar(&token, "token", "", "Smib's slack token")
	flag.StringVar(&commandDir, "commands", "", "Directory containing Smib's commands")
	flag.Parse()

	client := slack.New(token)

	cmd := command.New(commandDir)

	bot := smib.New(client, cmd)
	log.Print("Starting SMIB")
	log.Fatal(bot.ListenAndRobot())
}
