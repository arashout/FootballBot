package roborooney

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arashout/mlpapi"
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
		calculatePitchSlotID(pitch.ID, slot.ID),
	)
}

func createPitchSlot(_pitch mlpapi.Pitch, _slot mlpapi.Slot) pitchSlot {
	// Use the Pitch ID and Slot ID to create a unique identifer
	pitchSlotID := calculatePitchSlotID(_pitch.ID, _slot.ID)
	return pitchSlot{
		id:    pitchSlotID,
		pitch: _pitch,
		slot:  _slot,
		seen:  false,
	}
}

func getTimeRange() (time.Time, time.Time) {
	// Look for slots between now and 2 weeks ahead, which is the limit of MyLocalPitch API anyway
	t1 := time.Now()
	return t1, t1.AddDate(0, 0, 14)
}

func sendPOSTJSON(url string, jsonString string) {
	jsonBytes := []byte(jsonString)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(body))
}
