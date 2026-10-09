package id

import (

	"github.com/google/uuid"
)


func GenerateEventId(prefix string) string {
	

	id,err := uuid.NewV7()

	if err != nil {
		return ""
	}

	eventId := prefix+"_"+id.String()


	return eventId
}