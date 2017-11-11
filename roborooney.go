package roborooney

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/arashout/mlpapi"
)

const (
	robotName       = "roborooney"
	commandCheckout = "checkout"
	commandPoll     = "poll"
	commandList     = "list"
	commandUnseen   = "unseen"
	commandRules    = "rules"
	commandPitches  = "pitches"
	commandHelp     = "help"
	textHelp        = `
	I'm RoboRooney, the football bot. You can mention me whenever you want to find pitches to play on.
	@roborooney : Bring up this dialogue again
	@roborooney list : Lists the available slots that satisfy the rules
	@roborooney unseen : List the unseen slots available that satisfy the rules
	@roborooney rules : Lists the descriptions of the rules currently in effect
	@roborooney pitches : Lists the monitored pitches
	@roborooney poll : Start a poll with the available slots (Not working...)
	@roborooney checkout {pitch-slot ID} : Get the checkout link for a slot (pitch-slot ID is listed after each slot)
	`
)

var regexPitchSlotID = regexp.MustCompile(`\d{5}-\d{6}`)

type RoboRooney struct {
	cred      credentials
	mlpClient *mlpapi.MLPClient
	tracker   *Tracker
	ticker    *time.Ticker
	pitches   []mlpapi.Pitch
	rules     []mlpapi.Rule
}

// NewRobo creates a new initialized robo object that the client can interact with
func NewRobo(pitches []mlpapi.Pitch, rules []mlpapi.Rule) (robo *RoboRooney) {
	robo = &RoboRooney{}
	robo.cred = readCredentials()

	robo.mlpClient = mlpapi.New()
	robo.tracker = NewTracker()
	robo.ticker = time.NewTicker(time.Minute * time.Duration(robo.cred.TickerInterval))

	if len(pitches) == 0 {
		log.Fatal("Need atleast one pitch to check")
	}

	robo.pitches = pitches
	robo.rules = rules

	return robo
}

func readCredentials() credentials {
	log.Print("Reading credentials from enviroment:\n")
	tickerInterval, err := strconv.Atoi(os.Getenv("TICKER_INTERVAL"))
	if err != nil {
		log.Fatal("Unable to parse ticker interval: " + os.Getenv("TICKER_INTERVAL"))
	}

	cred := credentials{
		VertificationToken: os.Getenv("VERTIFICATION_TOKEN"),
		IncomingChannelID:  os.Getenv("INCOMING_CHANNEL_ID"),
		TickerInterval:     tickerInterval,
	}

	return cred
}

// Close robo
func (robo *RoboRooney) Close() {
	log.Println(robotName + " is shutting down.")
	robo.mlpClient.Close()
}

func (robo *RoboRooney) getFilteredPitchSlots(t1 time.Time, t2 time.Time) map[string]PitchSlot {
	pitchSlotMap := make(map[string]PitchSlot)

	for _, pitch := range robo.pitches {
		slots := robo.mlpClient.GetPitchSlots(pitch, t1, t2)
		filteredSlots := robo.mlpClient.FilterSlotsByRules(slots, robo.rules)

		for _, slot := range filteredSlots {
			pitchSlot := createPitchSlot(pitch, slot)
			pitchSlotMap[pitchSlot.id] = pitchSlot
		}

	}

	return pitchSlotMap
}

func (robo *RoboRooney) updateTracker() {
	t1, t2 := getTimeRange()

	newPitchSlotMap := robo.getFilteredPitchSlots(t1, t2)

	// Remove expired listings and mark pitchslots that are in both maps
	for _, oldPitchSlot := range robo.tracker.getMap() {
		_, ok := newPitchSlotMap[oldPitchSlot.id]
		if ok {
			robo.tracker.upsert(oldPitchSlot)
		} else {
			robo.tracker.remove(oldPitchSlot.id)
		}
	}

	for _, newPitchSlot := range newPitchSlotMap {
		robo.tracker.upsert(newPitchSlot)
	}
}
