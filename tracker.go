package roborooney

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Tracker
type Tracker struct {
	psMap map[string]pitchSlot
	mu    sync.RWMutex
}

func NewTracker() *Tracker {
	return &Tracker{
		psMap: make(map[string]pitchSlot),
	}
}

func (tracker *Tracker) upsert(ps pitchSlot) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	_, ok := tracker.psMap[ps.id]
	if ok {
		ps.seen = true
		tracker.psMap[ps.id] = ps
	} else {
		ps.seen = false
		tracker.psMap[ps.id] = ps
	}

}

func (tracker *Tracker) remove(pitchSlotID string) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	_, ok := tracker.psMap[pitchSlotID]
	if ok {
		delete(tracker.psMap, pitchSlotID)
	}
}

func (tracker *Tracker) retrieve(pitchSlotID string) (pitchSlot, error) {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	if ps, ok := tracker.psMap[pitchSlotID]; ok {
		return ps, nil
	}

	return pitchSlot{}, errors.New("pitch-slot-ID not found in tracker")
}

func (tracker *Tracker) retrieveAll() []pitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	psSlice := []pitchSlot{}
	for _, ps := range tracker.psMap {
		psSlice = append(psSlice, ps)
	}

	sort.Slice(psSlice, func(i, j int) bool {
		return psSlice[i].slot.Attributes.Starts.Unix() < psSlice[j].slot.Attributes.Starts.Unix()
	})

	return psSlice
}

func (tracker *Tracker) retrieveUnseen() []pitchSlot {
	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	unseenPitchSlots := []pitchSlot{}
	for _, ps := range tracker.psMap {
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

	for pitchSlotID, ps := range tracker.psMap {
		copyMap[pitchSlotID] = ps
	}

	return copyMap
}

func calculatePitchSlotID(pitchID, slotID string) string {
	return fmt.Sprintf("%s-%s", pitchID, slotID)
}
