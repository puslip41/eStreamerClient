package message

import (
	"fmt"
	"log"
)

var recordStorage map[string]fmt.Stringer = make(map[string]fmt.Stringer)

func getRecordKey(deviceID, eventID uint32 ) string {
	return fmt.Sprintf("%d#%d", deviceID, eventID)
}

func PutIntrusionEvent(deviceID, eventID uint32, stringer fmt.Stringer) string {
	key := getRecordKey(deviceID, eventID)

	if recordStorage[key] != nil {
		log.Println("cannot insert intrusion event record:", deviceID, eventID, stringer)
	} else {
		recordStorage[key] = stringer
	}

	return key
}

func GetIntrusionEvent(deviceID, eventID uint32) fmt.Stringer {
	key := getRecordKey(deviceID, eventID)

	v := recordStorage[key]

	if v != nil {
		delete(recordStorage, key)
	}

	return v
}

func GetCountInRecordStroage() int {
	return len(recordStorage)
}