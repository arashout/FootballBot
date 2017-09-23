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
	pitches     []*mlpapi.Pitch
}

// Credentials ...
type Credentials struct {
	APIToken  string `json:"apiToken"`
	ChannelID string `json:"channelId"`
	BotID     string `json:"botId,omitempty"`
}
