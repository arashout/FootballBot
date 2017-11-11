package roborooney

import (
	"fmt"
	"log"

	"github.com/arashout/mlpapi"
)

// NOTE: That if msgText is empty string that means call came from ticker
func handleCommand(robo *RoboRooney, command string, channelID string, msgText string) {
	var result string
	switch command {
	case commandList:
		result = handlerListCommand(robo)
	case commandUnseen:
		result = handlerUnseenCommand(robo, msgText == "")
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
	if result != "" {
		robo.sendMessage(result, channelID)
	}

}

func handlerListCommand(robo *RoboRooney) string {
	robo.updateTracker()

	textListSlots := ""
	pitchSlots := robo.tracker.retrieveAll()
	for _, pitchSlot := range pitchSlots {
		textSlot := fmt.Sprintf("%s\n", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		textListSlots += textSlot
	}
	return textListSlots
}

func handlerUnseenCommand(robo *RoboRooney, fromTicker bool) string {
	robo.updateTracker()

	textListSlots := ""
	pitchSlots := robo.tracker.retrieveUnseen()

	if len(pitchSlots) == 0 && !fromTicker {
		return "No new slots are available"
	}

	for _, pitchSlot := range pitchSlots {
		textSlot := fmt.Sprintf("%s\n", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		textListSlots += textSlot
	}
	return textListSlots
}

func handlerCheckoutCommand(robo *RoboRooney, msgTest string) string {
	robo.updateTracker()

	pitchSlotID := regexPitchSlotID.FindString(msgTest)
	if pitchSlotID != "" {
		pitchSlot, err := robo.tracker.retrieve(pitchSlotID)
		if err != nil {
			return "Pitch-Slot ID not found. Try listing all available bookings again"
		}
		return mlpapi.GetSlotCheckoutLink(pitchSlot.pitch, pitchSlot.slot)
	}
	return "No Pitch-Slot ID found in message. Make sure it is formatted correctly."
}

func handlerPollCommand(robo *RoboRooney) string {
	pitchSlots := robo.tracker.retrieveAll()

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
