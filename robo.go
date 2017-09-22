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
	rtm := robo.apiClient.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev.Msg.Text)
			if isBot(ev.Msg) {
				fmt.Print("This is a robot message")
			}
			fmt.Println(ev.Msg.BotID)
			if robo.isMentioned(ev.Msg.Text) {
				rtm.SendMessage(rtm.NewOutgoingMessage("You mentioned me!", robo.cred.ChannelID))
			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage("That wasn't my name"+ev.Msg.Text, robo.cred.ChannelID))
			}

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func (robo *RoboRooney) Close() {
	// Nothing to clean-up so far I think
}

// Simple Wrapper functions
func (robo *RoboRooney) isMentioned(s string) bool {
	return strings.Contains(s, RobotName)
}

func isBot(msg slack.Msg) bool {
	return msg.BotID == ""
}
