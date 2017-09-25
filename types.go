package roborooney

import (
	"github.com/arashout/mlpapi"
	"github.com/nlopes/slack"
)

type RoboRooney struct {
	cred        *Credentials
	slackClient *slack.Client
	mlpClient   *mlpapi.MLPClient
	rtm         *slack.RTM
	tracker     *Tracker
	pitches     []mlpapi.Pitch
	rules       []mlpapi.SlotFilter
}

// Credentials ...
type Credentials struct {
	APIToken  string `json:"apiToken"`
	ChannelID string `json:"channelId"`
	BotID     string `json:"botId,omitempty"`
}

// Tracker
type Tracker struct {
	pitchSlotMap map[string]PitchSlot
}

// PitchSlot is a struct used in tracker for keeping track of all the already queryed slots for retrieval
type PitchSlot struct {
	pitch mlpapi.Pitch
	slot  mlpapi.Slot
}
