package roborooney

import (
	"errors"
	"fmt"

	"github.com/arashout/mlpapi"
)

func NewTracker() *Tracker {
	return &Tracker{
		pitchSlotMap: make(map[string]PitchSlot),
	}
}

func (tracker *Tracker) Insert(_pitch mlpapi.Pitch, _slot mlpapi.Slot) {
	// Use the Pitch ID and Slot ID to create a unique identifer
	pitchSlotID := calculatePitchSlotId(_pitch.ID, _slot.ID)
	tracker.pitchSlotMap[pitchSlotID] = PitchSlot{
		pitch: _pitch,
		slot:  _slot,
	}
}

func (tracker *Tracker) Retrieve(pitchSlotID string) (PitchSlot, error) {
	if pitchSlot, ok := tracker.pitchSlotMap[pitchSlotID]; ok {
		return pitchSlot, nil
	}
	return PitchSlot{}, errors.New("pitch-slot-ID not found in tracker")
}

func (tracker *Tracker) RetrieveAll() []PitchSlot {
	pitchSlots := []PitchSlot{}
	for _, pitchSlot := range tracker.pitchSlotMap {
		pitchSlots = append(pitchSlots, pitchSlot)
	}
	return pitchSlots
}

func (tracker *Tracker) Clear() {
	tracker.pitchSlotMap = make(map[string]PitchSlot)
}

func calculatePitchSlotId(pitchID, slotID string) string {
	return fmt.Sprintf("%s-%s", pitchID, slotID)
}
