package roborooney

import (
	"github.com/arashout/mlpapi"
)

// Credentials ...
type credentials struct {
	VerificationToken  string
	IncomingWebhookURL string
	TickerInterval     int // In minutes
}

// PitchSlot is a struct used in tracker for keeping track of all the already queryed slots for retrieval
type pitchSlot struct {
	id    string
	seen  bool
	pitch mlpapi.Pitch
	slot  mlpapi.Slot
}
