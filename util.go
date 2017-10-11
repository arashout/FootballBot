package roborooney

import (
	"fmt"
	"strconv"

	"github.com/arashout/mlpapi"
	"github.com/nlopes/slack"
)

func formatSlotMessage(pitch mlpapi.Pitch, slot mlpapi.Slot) string {
	const layout = "Mon Jan 2\t15:04"
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
