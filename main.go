package main

import (
	"flag"
	"log"

	"github.com/nlopes/slack"
	"github.com/somakeit/slacker-smib/internal/smib"
)

func main() {
	var (
		token string
	)

	flag.StringVar(&token, "token", "", "Smib's slack token")
	flag.Parse()

	client := slack.New(token)

	bot := smib.New(client)
	log.Print("Starting SMIB")
	log.Fatal(bot.ListenAndRobot())
}
