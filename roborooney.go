package roborooney

import (
	"bytes"
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

	// Look for slots between now and 2 weeks ahead, which is the limit of MyLocalPitch API anyway
	t1 := time.Now()
	t2 := t1.AddDate(0, 0, 14)

	for msg := range robo.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if !isBot(ev.Msg) && robo.isMentioned(&ev.Msg) {
				// Post response to the message the event is from
				currentChannelID := ev.Msg.Channel

				if strings.Contains(ev.Msg.Text, commandList) {
					// Update the tracker and list all available slots as one message
					var messageBuffer bytes.Buffer

					robo.UpdateTracker(t1, t2)
					pitchSlots := robo.tracker.RetrieveAll()
					for _, pitchSlot := range pitchSlots {
						textSlot := fmt.Sprintf("%s\n", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
						messageBuffer.WriteString(textSlot)
					}
					robo.sendMessage(messageBuffer.String(), currentChannelID)

				} else if strings.Contains(ev.Msg.Text, commandCheckout) {
					pitchSlotID := regexPitchSlotID.FindString(ev.Msg.Text)
					if pitchSlotID != "" {
						pitchSlot, err := robo.tracker.Retrieve(pitchSlotID)
						if err != nil {
							robo.sendMessage("Pitch-Slot ID not found. Try listing all available bookings again", currentChannelID)
						} else {
							checkoutLink := mlpapi.GetSlotCheckoutLink(pitchSlot.pitch, pitchSlot.slot)
							robo.sendMessage(checkoutLink, currentChannelID)
						}
					}
				} else if strings.Contains(ev.Msg.Text, commandPoll) {
					robo.UpdateTracker(t1, t2)
					robo.createPoll(robo.tracker.RetrieveAll(), currentChannelID)
				} else if strings.Contains(ev.Msg.Text, commandRules) {
					// TODO: Message buffers are definetely over kill and I should find a cleaner way
					var messageBuffer bytes.Buffer
					for _, rule := range robo.rules {
						messageBuffer.WriteString("-" + rule.Description + "\n")
					}
					robo.sendMessage(messageBuffer.String(), currentChannelID)
				} else if strings.Contains(ev.Msg.Text, commandPitches) {
					var messageBuffer bytes.Buffer
					for _, pitch := range robo.pitches {
						messageBuffer.WriteString("-" + pitch.Name + "\n")
					}
					robo.sendMessage(messageBuffer.String(), currentChannelID)
				} else {
					robo.sendMessage(textHelp, currentChannelID)
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

func (robo *RoboRooney) createPoll(pitchSlots []PitchSlot, channelID string) {
	if len(pitchSlots) == 0 {
		robo.sendMessage("No slots available for polling\nTry checking availablity first.", channelID)
	}
	// TODO: Check writing errors
	var pollBuffer bytes.Buffer
	pollBuffer.WriteString("/poll 'Which time(s) works best?' ")

	for _, pitchSlot := range pitchSlots {
		optionString := fmt.Sprintf(" \"%s\" ", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		pollBuffer.WriteString(optionString)
	}

	robo.sendMessage(pollBuffer.String(), channelID)
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
