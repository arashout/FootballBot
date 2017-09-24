package roborooney

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arashout/mlpapi"
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
						robo.sendMessage("Slots available for:")
						robo.sendMessage(pitch.VenuePath)
						for _, slot := range filteredSlots {
							robo.sendMessage(formatSlotMessage(slot, pitch, true))
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

func formatSlotMessage(slot mlpapi.Slot, pitch mlpapi.Pitch, withLink bool) string {
	const layout = "Mon Jan 2 15:04:05"
	duration := slot.Attributes.Ends.Sub(slot.Attributes.Starts).Hours()
	stringDuration := strconv.FormatFloat(duration, 'f', -1, 64)
	if withLink {
		return fmt.Sprintf(
			"%s\tDuration: %s Hour(s)\tAt %s\nLink:\t%s",
			slot.Attributes.Starts.Format(layout),
			stringDuration,
			pitch.VenuePath,
			mlpapi.GetSlotCheckoutLink(slot, pitch),
		)
	}

	return fmt.Sprintf(
		"%s\tDuration: %s Hour(s)\tAt %s",
		slot.Attributes.Starts.Format(layout),
		stringDuration,
		pitch.VenuePath,
	)
}
