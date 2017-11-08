package roborooney

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/arashout/mlpapi"
)

// Tracker
type Tracker struct {
	pitchSlotMap map[string]PitchSlot
	mu           sync.RWMutex
}

func NewTracker() *Tracker {
	return &Tracker{
		pitchSlotMap: make(map[string]PitchSlot),
	}
}

func (tracker *Tracker) Insert(_pitch mlpapi.Pitch, _slot mlpapi.Slot) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Use the Pitch ID and Slot ID to create a unique identifer
	pitchSlotID := calculatePitchSlotId(_pitch.ID, _slot.ID)
	tracker.pitchSlotMap[pitchSlotID] = PitchSlot{
		pitch: _pitch,
		slot:  _slot,
	}
}

func (tracker *Tracker) Retrieve(pitchSlotID string) (PitchSlot, error) {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	if pitchSlot, ok := tracker.pitchSlotMap[pitchSlotID]; ok {
		return pitchSlot, nil
	}
	return PitchSlot{}, errors.New("pitch-slot-ID not found in tracker")
}

func (tracker *Tracker) RetrieveAll() []PitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	pitchSlots := []PitchSlot{}
	for _, pitchSlot := range tracker.pitchSlotMap {
		pitchSlots = append(pitchSlots, pitchSlot)
	}
	sort.Slice(pitchSlots, func(i, j int) bool {
		return pitchSlots[i].slot.Attributes.Starts.Unix() < pitchSlots[j].slot.Attributes.Starts.Unix()
	})
	return pitchSlots
}

func (tracker *Tracker) Clear() {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	tracker.pitchSlotMap = make(map[string]PitchSlot)
}

func calculatePitchSlotId(pitchID, slotID string) string {
	return fmt.Sprintf("%s-%s", pitchID, slotID)
}
