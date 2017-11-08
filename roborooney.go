package roborooney

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arashout/mlpapi"
	"github.com/nlopes/slack"
)

const (
	robotName       = "roborooney"
	commandCheckout = "checkout"
	commandPoll     = "poll"
	commandList     = "list"
	commandRules    = "rules"
	commandPitches  = "pitches"
	commandHelp     = "help"
	textHelp        = `
	I'm RoboRooney, the football bot. You can mention me whenever you want to find pitches to play on.
	@roborooney : Bring up this dialogue again
	@roborooney list : Lists the available slots that satisfy the rules
	@roborooney rules : Lists the descriptions of the rules currently in effect
	@roborooney pitches : Lists the monitored pitches
	@roborooney poll : Start a poll with the available slots (Not working...)
	@roborooney checkout {pitch-slot ID} : Get the checkout link for a slot (pitch-slot ID is listed after each slot)
	`
)

var regexPitchSlotID = regexp.MustCompile(`\d{5}-\d{6}`)

// NewRobo creates a new initialized robo object that the client can interact with
func NewRobo(pitches []mlpapi.Pitch, rules []mlpapi.Rule, cred *Credentials) (robo *RoboRooney) {
	robo = &RoboRooney{}
	robo.mlpClient = mlpapi.New()
	robo.tracker = NewTracker()

	robo.initialize(cred)

	if len(pitches) == 0 {
		log.Fatal("Need atleast one pitch to check")
	}

	robo.pitches = pitches
	robo.rules = rules

	return robo
}

func (robo *RoboRooney) initialize(cred *Credentials) {
	robo.cred = cred
	robo.slackClient = slack.New(robo.cred.APIToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	robo.slackClient.SetDebug(false)
}

// Connect to Slack and start main loop
func (robo *RoboRooney) Connect() {
	log.Println("Creating a websocket connection with Slack")
	robo.rtm = robo.slackClient.NewRTM()
	go robo.rtm.ManageConnection()
	log.Println(robotName + " is ready to go.")

	for msg := range robo.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if !isBot(ev.Msg) && robo.isMentioned(&ev.Msg) {
				// Determine which command is passed as an argument, default to sending help message
				// Note: We send these messages only to the channel we received the event from
				if strings.Contains(ev.Msg.Text, commandList) {
					handleCommand(robo, commandList, ev.Msg.Text, ev.Msg.Channel)
				} else if strings.Contains(ev.Msg.Text, commandCheckout) {
					handleCommand(robo, commandCheckout, ev.Msg.Text, ev.Msg.Channel)
				} else if strings.Contains(ev.Msg.Text, commandPoll) {
					handleCommand(robo, commandPoll, ev.Msg.Text, ev.Msg.Channel)
				} else if strings.Contains(ev.Msg.Text, commandRules) {
					handleCommand(robo, commandRules, ev.Msg.Text, ev.Msg.Channel)
				} else if strings.Contains(ev.Msg.Text, commandPitches) {
					handleCommand(robo, commandPitches, ev.Msg.Text, ev.Msg.Channel)
				} else {
					handleCommand(robo, commandHelp, ev.Msg.Text, ev.Msg.Channel)
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

// Close robo
func (robo *RoboRooney) Close() {
	log.Println(robotName + " is shutting down.")
	robo.mlpClient.Close()
}

func (robo *RoboRooney) isMentioned(msg *slack.Msg) bool {
	if robo.cred.BotID != "" {
		return strings.Contains(msg.Text, robotName) || strings.Contains(msg.Text, fmt.Sprintf("<@%s>", robo.cred.BotID))
	}
	return strings.Contains(msg.Text, robotName)
}

func (robo *RoboRooney) sendMessage(s string, channelID string) {
	robo.rtm.SendMessage(robo.rtm.NewOutgoingMessage(s, channelID))
}

// UpdateTracker updates the list of available slots in the shared tracker struct given two time objects
func (robo *RoboRooney) UpdateTracker(t1 time.Time, t2 time.Time) {
	robo.tracker.Clear()

	for _, pitch := range robo.pitches {
		slots := robo.mlpClient.GetPitchSlots(pitch, t1, t2)
		filteredSlots := robo.mlpClient.FilterSlotsByRules(slots, robo.rules)
		for _, slot := range filteredSlots {
			robo.tracker.Insert(pitch, slot)
		}
	}
}
