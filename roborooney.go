package roborooney

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/arashout/mlpapi"
	"github.com/davecgh/go-spew/spew"
	"github.com/nlopes/slack"
)

const (
	robotName = "roborooney"
)

func NewRobo(pitches []mlpapi.Pitch, rules []func(mlpapi.Slot) bool) (robo *RoboRooney) {
	robo = &RoboRooney{}
	robo.mlpClient = mlpapi.New()
	robo.initialize()
	if len(pitches) == 0 {
		log.Fatal("Need atleast one pitch to check")
	}
	robo.pitches = pitches
	robo.rules = rules
	return robo
}

func (robo *RoboRooney) initialize() {
	robo.cred = &Credentials{}
	robo.cred.Read()

	robo.slackClient = slack.New(robo.cred.APIToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	robo.slackClient.SetDebug(false)
}

func (robo *RoboRooney) Connect() {
	robo.rtm = robo.slackClient.NewRTM()
	go robo.rtm.ManageConnection()

	t1 := time.Now()
	t2 := t1.AddDate(0, 0, 14)

	for msg := range robo.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if !isBot(ev.Msg) {
				if robo.isMentioned(&ev.Msg) {
					robo.sendMessage("You mentioned me!")
					for _, pitch := range robo.pitches {
						slots := robo.mlpClient.GetPitchSlots(pitch, t1, t2)
						filteredSlots := robo.mlpClient.FilterSlotsByRules(slots, robo.rules)
						for _, slot := range filteredSlots {
							robo.sendMessage(spew.Sdump(slot))
						}
					}
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
	robo.mlpClient.Close()
}

// Simple Wrapper functions
func (robo *RoboRooney) isMentioned(msg *slack.Msg) bool {
	if robo.cred.BotID != "" {
		return strings.Contains(msg.Text, robotName) || strings.Contains(msg.Text, fmt.Sprintf("<@%s>", robo.cred.BotID))
	}
	return strings.Contains(msg.Text, robotName)
}

func isBot(msg slack.Msg) bool {
	return msg.BotID != ""
}

func (robo *RoboRooney) sendMessage(s string) {
	robo.rtm.SendMessage(robo.rtm.NewOutgoingMessage(s, robo.cred.ChannelID))
}
