package roborooney

import (
	"github.com/arashout/mlpapi"
)

type credentials struct {
	VerificationToken  string
	IncomingWebhookURL string
	TickerInterval     int // In minutes
}

type requestFromSlack struct {
	Type           string `json:"url_verification"`
	Token          string `json:"token"`
	ChallengeValue string `json:"challenge"`
}

// PitchSlot is a struct used in tracker for keeping track of all the already queryed slots for retrieval
type pitchSlot struct {
	id    string
	seen  bool
	pitch mlpapi.Pitch
	slot  mlpapi.Slot
}
