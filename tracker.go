package roborooney

import (
	"errors"
	"fmt"
	"sort"
	"sync"
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

func (tracker *Tracker) upsert(pitchSlot PitchSlot) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	_, ok := tracker.pitchSlotMap[pitchSlot.id]
	if ok {
		pitchSlot.seen = true
		tracker.pitchSlotMap[pitchSlot.id] = pitchSlot
	} else {
		pitchSlot.seen = false
		tracker.pitchSlotMap[pitchSlot.id] = pitchSlot
	}

}

func (tracker *Tracker) remove(pitchSlotID string) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	_, ok := tracker.pitchSlotMap[pitchSlotID]
	if ok {
		delete(tracker.pitchSlotMap, pitchSlotID)
	}
}

func (tracker *Tracker) retrieve(pitchSlotID string) (PitchSlot, error) {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	if pitchSlot, ok := tracker.pitchSlotMap[pitchSlotID]; ok {
		return pitchSlot, nil
	}

	return PitchSlot{}, errors.New("pitch-slot-ID not found in tracker")
}

func (tracker *Tracker) retrieveAll() []PitchSlot {
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

func (tracker *Tracker) retrieveUnseen() []PitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	unseenPitchSlots := []PitchSlot{}
	for _, pitchSlot := range tracker.pitchSlotMap {
		if !pitchSlot.seen {
			unseenPitchSlots = append(unseenPitchSlots, pitchSlot)
		}
	}

	sort.Slice(unseenPitchSlots, func(i, j int) bool {
		return unseenPitchSlots[i].slot.Attributes.Starts.Unix() < unseenPitchSlots[j].slot.Attributes.Starts.Unix()
	})

	return unseenPitchSlots
}

func (tracker *Tracker) getMap() map[string]PitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	copyMap := make(map[string]PitchSlot)

	for pitchSlotID, pitchSlot := range tracker.pitchSlotMap {
		copyMap[pitchSlotID] = pitchSlot
	}

	return copyMap
}

func calculatePitchSlotID(pitchID, slotID string) string {
	return fmt.Sprintf("%s-%s", pitchID, slotID)
}
