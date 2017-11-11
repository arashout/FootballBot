package roborooney

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Tracker
type Tracker struct {
	pitchSlotMap map[string]pitchSlot
	mu           sync.RWMutex
}

func NewTracker() *Tracker {
	return &Tracker{
		pitchSlotMap: make(map[string]pitchSlot),
	}
}

func (tracker *Tracker) upsert(ps pitchSlot) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	_, ok := tracker.pitchSlotMap[ps.id]
	if ok {
		pitchslot.seen = true
		tracker.pitchSlotMap[ps.id] = ps
	} else {
		ps.seen = false
		tracker.pitchSlotMap[ps.id] = ps
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

func (tracker *Tracker) retrieve(pitchSlotID string) (pitchSlot, error) {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	if ps, ok := tracker.pitchSlotMap[pitchSlotID]; ok {
		return ps, nil
	}

	return pitchSlot{}, errors.New("pitch-slot-ID not found in tracker")
}

func (tracker *Tracker) retrieveAll() []pitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	pitchSlots := []pitchSlot{}
	for _, ps := range tracker.pitchSlotMap {
		pitchSlots = append(pitchSlots, ps)
	}

	sort.Slice(pitchSlots, func(i, j int) bool {
		return pitchSlots[i].slot.Attributes.Starts.Unix() < pitchSlots[j].slot.Attributes.Starts.Unix()
	})

	return pitchSlots
}

func (tracker *Tracker) retrieveUnseen() []pitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	unseenPitchSlots := []pitchSlot{}
	for _, ps := range tracker.pitchSlotMap {
		if !ps.seen {
			unseenPitchSlots = append(unseenPitchSlots, ps)
		}
	}

	sort.Slice(unseenPitchSlots, func(i, j int) bool {
		return unseenPitchSlots[i].slot.Attributes.Starts.Unix() < unseenPitchSlots[j].slot.Attributes.Starts.Unix()
	})

	return unseenPitchSlots
}

func (tracker *Tracker) getMap() map[string]pitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	copyMap := make(map[string]pitchSlot)

	for pitchSlotID, ps := range tracker.pitchSlotMap {
		copyMap[pitchSlotID] = ps
	}

	return copyMap
}

func calculatePitchSlotID(pitchID, slotID string) string {
	return fmt.Sprintf("%s-%s", pitchID, slotID)
}
