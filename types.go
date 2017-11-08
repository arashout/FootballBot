package roborooney

import (
	"github.com/arashout/mlpapi"
)

// Credentials ...
type Credentials struct {
	APIToken              string
	BotID                 string
	NotificationChannelID string
	TickerInterval        int // In minutes
}

// PitchSlot is a struct used in tracker for keeping track of all the already queryed slots for retrieval
type PitchSlot struct {
	pitch mlpapi.Pitch
	slot  mlpapi.Slot
}
