package roborooney

import (
	"fmt"
	"log"
	"time"

	"github.com/arashout/mlpapi"
)

func getTimeRange() (time.Time, time.Time) {
	// Look for slots between now and 2 weeks ahead, which is the limit of MyLocalPitch API anyway
	t1 := time.Now()
	return t1, t1.AddDate(0, 0, 14)
}

// NOTE: That msgText isn't always necessary so you can simply pass an empty string
func handleCommand(robo *RoboRooney, command string, channelID string, msgText string) {
	t1, t2 := getTimeRange()
	// Update the tracker so we have most up to date listings
	robo.UpdateTracker(t1, t2)

	var result string
	switch command {
	case commandList:
		result = handlerListCommand(robo)
	case commandCheckout:
		result = handlerCheckoutCommand(robo, msgText)
	case commandPoll:
		result = handlerPollCommand(robo)
	case commandRules:
		result = handlerRulesCommand(robo)
	case commandPitches:
		result = handlerPitchesCommand(robo)
	case commandHelp:
		result = textHelp
	default:
		log.Println("Command not recognized!")
	}
	robo.sendMessage(result, channelID)

}

func handlerListCommand(robo *RoboRooney) string {
	textListSlots := ""
	pitchSlots := robo.tracker.RetrieveAll()
	for _, pitchSlot := range pitchSlots {
		textSlot := fmt.Sprintf("%s\n", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		textListSlots += textSlot
	}
	return textListSlots
}

func handlerCheckoutCommand(robo *RoboRooney, msgTest string) string {
	pitchSlotID := regexPitchSlotID.FindString(msgTest)
	if pitchSlotID != "" {
		pitchSlot, err := robo.tracker.Retrieve(pitchSlotID)
		if err != nil {
			return "Pitch-Slot ID not found. Try listing all available bookings again"
		}
		return mlpapi.GetSlotCheckoutLink(pitchSlot.pitch, pitchSlot.slot)
	}
	return "No Pitch-Slot ID found in message. Make sure it is formatted correctly."
}

func handlerPollCommand(robo *RoboRooney) string {
	pitchSlots := robo.tracker.RetrieveAll()

	if len(pitchSlots) == 0 {
		return "No slots available for polling\nTry checking availablity first."
	}

	textPoll := "/poll 'Which time(s) works best?' "

	for _, pitchSlot := range pitchSlots {
		optionString := fmt.Sprintf(" \"%s\" ", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		textPoll += optionString
	}

	return textPoll
}

func handlerRulesCommand(robo *RoboRooney) string {
	textRules := ""
	for _, rule := range robo.rules {
		textRules += "-" + rule.Description + "\n"
	}
	return textRules
}

func handlerPitchesCommand(robo *RoboRooney) string {
	textPitches := ""
	for _, pitch := range robo.pitches {
		textPitches += "-" + pitch.Name + "\n"
	}
	return textPitches
}
