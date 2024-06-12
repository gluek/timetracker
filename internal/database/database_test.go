package database

import (
	"log"
	"testing"
	"time"
)

func TestGetLocationsDays(t *testing.T) {
	Connect()
	locations := GetLocationDaysForMonth(time.Now())
	for _, v := range locations {
		log.Println("Location:", v)
	}
	Close()
}
