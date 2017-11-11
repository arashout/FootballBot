package roborooney

import (
	"github.com/arashout/mlpapi"
)

// Credentials ...
type Credentials struct {
	VertificationToken string
	IncomingChannelID  string
	TickerInterval     int // In minutes
}

// PitchSlot is a struct used in tracker for keeping track of all the already queryed slots for retrieval
type PitchSlot struct {
	id    string
	seen  bool
	pitch mlpapi.Pitch
	slot  mlpapi.Slot
}
