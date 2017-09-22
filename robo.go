package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

const (
	RobotName = "roborooney"
)

type RoboRooney struct {
	cred      *Credentials
	apiClient *slack.Client
	rtm       *slack.RTM
}

func NewRobo() (robo *RoboRooney) {
	robo = &RoboRooney{}
	robo.initialize()
	return robo
}

func (robo *RoboRooney) initialize() {
	robo.cred = &Credentials{}
	robo.cred.Read()

	robo.apiClient = slack.New(robo.cred.APIToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	robo.apiClient.SetDebug(false)
}

func (robo *RoboRooney) Connect() {
	robo.rtm = robo.apiClient.NewRTM()
	go robo.rtm.ManageConnection()

	for msg := range robo.rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if !isBot(ev.Msg) {
				if robo.isMentioned(&ev.Msg) {
					robo.sendMessage("You mentioned me!")
				}
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return
		}
	}
}

func (robo *RoboRooney) Close() {
	// Nothing to clean-up so far I think
}

// Simple Wrapper functions
func (robo *RoboRooney) isMentioned(msg *slack.Msg) bool {
	if robo.cred.BotID != "" {
		return strings.Contains(msg.Text, RobotName) || strings.Contains(msg.Text, fmt.Sprintf("<@%s>", robo.cred.BotID))
	}
	return strings.Contains(msg.Text, RobotName)
}

func isBot(msg slack.Msg) bool {
	return msg.BotID != ""
}

func (robo *RoboRooney) sendMessage(s string) {
	robo.rtm.SendMessage(robo.rtm.NewOutgoingMessage(s, robo.cred.ChannelID))
}
