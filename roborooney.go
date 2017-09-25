package roborooney

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arashout/mlpapi"
	"github.com/nlopes/slack"
)

const (
	robotName       = "roborooney"
	commandCheckout = "checkout"
	commandPoll     = "poll"
	commandHelp     = "help"
)

var regexPitchSlotID = regexp.MustCompile(`\d{5}-\d{6}`)

func NewRobo(pitches []mlpapi.Pitch, rules []mlpapi.SlotFilter) (robo *RoboRooney) {
	robo = &RoboRooney{}
	robo.mlpClient = mlpapi.New()
	robo.tracker = NewTracker()

	robo.initialize()
	if len(pitches) == 0 {
		log.Fatal("Need atleast one pitch to check")
	}
	robo.pitches = pitches
	robo.rules = rules
	return robo
}

func (robo *RoboRooney) initialize() {
	log.Println("Reading config.json for credentials")
	robo.cred = &Credentials{}
	robo.cred.Read()

	if robo.cred.BotID == "" {
		log.Println("BotID not set, at @roborooney will not work...")
	}

	robo.slackClient = slack.New(robo.cred.APIToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	robo.slackClient.SetDebug(false)
}

func (robo *RoboRooney) Connect() {
	log.Println("Creating a websocket connection with Slack")
	robo.rtm = robo.slackClient.NewRTM()
	go robo.rtm.ManageConnection()
	log.Println(robotName + " is ready to go.")

	// Look for slots between now and 2 weeks ahead
	t1 := time.Now()
	t2 := t1.AddDate(0, 0, 14)

	for msg := range robo.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if !isBot(ev.Msg) {
				if robo.isMentioned(&ev.Msg) {
					// TODO: Have a help command
					if strings.Contains(ev.Msg.Text, commandHelp) {
						robo.sendMessage("Not implemented yet")
					} else if strings.Contains(ev.Msg.Text, commandCheckout) {
						pitchSlotID := regexPitchSlotID.FindString(ev.Msg.Text)
						if pitchSlotID != "" {
							pitchSlot, err := robo.tracker.Retrieve(pitchSlotID)
							if err != nil {
								robo.sendMessage("Pitch-Slot ID not found. Try listing all available bookings again")
							} else {
								checkoutLink := mlpapi.GetSlotCheckoutLink(pitchSlot.pitch, pitchSlot.slot)
								robo.sendMessage(checkoutLink)
							}
						}
					} else if strings.Contains(ev.Msg.Text, commandPoll) {
						// TODO: Create a poll from available slots
						robo.sendMessage("Not implemented yet")
					} else {
						robo.sendMessage("Slots available for:")
						for _, pitch := range robo.pitches {
							slots := robo.mlpClient.GetPitchSlots(pitch, t1, t2)
							filteredSlots := robo.mlpClient.FilterSlotsByRules(slots, robo.rules)
							for _, slot := range filteredSlots {
								robo.tracker.Insert(pitch, slot)
								robo.sendMessage(formatSlotMessage(pitch, slot))
							}
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
	log.Println(robotName + " is shutting down.")
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

func formatSlotMessage(pitch mlpapi.Pitch, slot mlpapi.Slot) string {
	const layout = "Mon Jan 2 15:04:05"
	duration := slot.Attributes.Ends.Sub(slot.Attributes.Starts).Hours()
	stringDuration := strconv.FormatFloat(duration, 'f', -1, 64)
	return fmt.Sprintf(
		"%s\tDuration: %s Hour(s)\t@\t%s\tID:\t%s",
		slot.Attributes.Starts.Format(layout),
		stringDuration,
		pitch.Name,
		calculatePitchSlotId(pitch.ID, slot.ID),
	)
}
