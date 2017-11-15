package roborooney

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/arashout/mlpapi"
)

// HandleEvent is an endpoint for handling event subscriptions
func (robo *RoboRooney) HandleEvent(w http.ResponseWriter, r *http.Request) {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	var req requestFromSlack
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	log.Println(req.ChallengeValue)
	// Write challenge back in plain text
	fmt.Fprintln(w, req.ChallengeValue)

}
func (robo *RoboRooney) HandleSlash(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}
	//Verify token - To ensure request coming from Slack
	verificationCode := r.Form.Get("token")
	if verificationCode != robo.cred.VerificationToken {
		http.Error(w, "Wrong verification token sent", http.StatusUnauthorized)
		return
	}

	// Get text called with slash command
	textSlash := r.Form.Get("text")

	var textResult string
	if strings.Contains(textSlash, commandList) {
		textResult = robo.handlerListCommand()
	} else if strings.Contains(textSlash, commandUnseen) {
		textResult = robo.handlerUnseenCommand(false)
	} else if strings.Contains(textSlash, commandRules) {
		textResult = robo.handlerRulesCommand()
	} else if strings.Contains(textSlash, commandPitches) {
		textResult = robo.handlerPitchesCommand()
	} else if strings.Contains(textSlash, commandCheckout) {
		textResult = robo.handlerCheckoutCommand(textSlash)
	} else {
		fmt.Fprintln(w, textHelp)
		return
	}

	fmt.Fprintln(w, textResult)
}

func (robo *RoboRooney) handlerListCommand() string {
	robo.updateTracker()

	textListSlots := ""
	pitchSlots := robo.tracker.retrieveAll()
	for _, pitchSlot := range pitchSlots {
		textSlot := fmt.Sprintf("%s\n", formatSlotMessage(pitchSlot.pitch, pitchSlot.slot))
		textListSlots += textSlot
	}
	return textListSlots
}

func (robo *RoboRooney) handlerUnseenCommand(fromTicker bool) string {
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

func (robo *RoboRooney) handlerCheckoutCommand(msgText string) string {
	robo.updateTracker()

	pitchSlotID := regexPitchSlotID.FindString(msgText)
	if pitchSlotID != "" {
		pitchSlot, err := robo.tracker.retrieve(pitchSlotID)
		if err != nil {
			return "Pitch-Slot ID not found. Try listing all available bookings again"
		}
		return mlpapi.GetSlotCheckoutLink(pitchSlot.pitch, pitchSlot.slot)
	}
	return "No Pitch-Slot ID found in message. Make sure it is formatted correctly."
}

func (robo *RoboRooney) handlerRulesCommand() string {
	textRules := ""
	for _, rule := range robo.rules {
		textRules += "-" + rule.Description + "\n"
	}
	return textRules
}

func (robo *RoboRooney) handlerPitchesCommand() string {
	textPitches := ""
	for _, pitch := range robo.pitches {
		textPitches += "-" + pitch.Name + "\n"
	}
	return textPitches
}
