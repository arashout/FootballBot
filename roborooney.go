package roborooney

import (
	"bytes"
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
	textHelp        = `
	I'm RoboRooney, the football bot. You can mention me whenever you want to find pitches to play on.
	@roborooney : List available slots at nearby pitches
	@roborooney help : Bring up this dialague again
	@roborooney poll : Start a poll with the available slots
	@roborooney checkout {pitch-slot ID} : Get the checkout link for a slot (pitch-slot ID is listed after each slot)
	`
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

	robo.sendMessage(textHelp)

	// Look for slots between now and 2 weeks ahead
	t1 := time.Now()
	t2 := t1.AddDate(0, 0, 14)

	for msg := range robo.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			if !isBot(ev.Msg) && robo.isMentioned(&ev.Msg) {
				if strings.Contains(ev.Msg.Text, commandHelp) {
					robo.sendMessage(textHelp)
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
					robo.UpdateTracker(t1, t2)
					robo.createPoll(robo.tracker.RetrieveAll())
				} else {
					// Update the tracker and list all available slots as one message
					var messageBuffer bytes.Buffer

					robo.UpdateTracker(t1, t2)
					pitchSlots := robo.tracker.RetrieveAll()
					for _, pitchSlot := range pitchSlots {
						textSlot := fmt.Sprintf("%s\n", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
						messageBuffer.WriteString(textSlot)
					}
					robo.sendMessage(messageBuffer.String())
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

func (robo *RoboRooney) isMentioned(msg *slack.Msg) bool {
	if robo.cred.BotID != "" {
		return strings.Contains(msg.Text, robotName) || strings.Contains(msg.Text, fmt.Sprintf("<@%s>", robo.cred.BotID))
	}
	return strings.Contains(msg.Text, robotName)
}

func (robo *RoboRooney) sendMessage(s string) {
	robo.rtm.SendMessage(robo.rtm.NewOutgoingMessage(s, robo.cred.ChannelID))
}

func (robo *RoboRooney) createPoll(pitchSlots []PitchSlot) {
	if len(pitchSlots) == 0 {
		robo.sendMessage("No slots available for polling\nTry checking availablity first.")
	}
	// TODO: Check writing errors
	var pollBuffer bytes.Buffer
	pollBuffer.WriteString("/poll 'Which time(s) works best?' ")

	for _, pitchSlot := range pitchSlots {
		optionString := fmt.Sprintf(" '%s' ", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		pollBuffer.WriteString(optionString)
	}

	robo.sendMessage(pollBuffer.String())
}

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

func formatSlotMessage(pitch mlpapi.Pitch, slot mlpapi.Slot) string {
	const layout = "Mon Jan 2\t15:04:05"
	duration := slot.Attributes.Ends.Sub(slot.Attributes.Starts).Hours()
	stringDuration := strconv.FormatFloat(duration, 'f', -1, 64)
	return fmt.Sprintf(
		"%s\t%s Hour(s)\t@\t%s\tID:\t%s",
		slot.Attributes.Starts.Format(layout),
		stringDuration,
		pitch.Name,
		calculatePitchSlotId(pitch.ID, slot.ID),
	)
}

func isBot(msg slack.Msg) bool {
	return msg.BotID != ""
}
