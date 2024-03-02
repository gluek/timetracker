package handlers

import (
	"fmt"
	"testing"
	"time"
)

func TestGetWorkDays(t *testing.T) {
	fmt.Println("Year: 2024")
	for i := range 12 {
		day, _ := time.Parse("2006-01", fmt.Sprintf("2024-%02d", i+1))
		fmt.Printf("    Work Days %s: %d\n", day.Month().String(), GetWorkDays(day))
	}
}
