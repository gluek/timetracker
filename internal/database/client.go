package database

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB
var err error

type Timeframe struct {
	ID       string `json:"id"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Duration string `json:"duration"`
	Project  string `json:"project"`
}

type Config struct {
	DailyHours time.Duration
}

func Connect() {
	Instance, err = gorm.Open(sqlite.Open("./internal/database/timetrack.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		panic("Cannot connect to DB")
	}
	log.Println("Connected to Database...")
}

func Migrate() {
	Instance.AutoMigrate(&Timeframe{})
	log.Println("Database Migration Completed...")
}

func checkIfEntryExists(entryID string) bool {
	var entry Timeframe
	Instance.First(&entry, entryID)
	return entry.ID != ""
}

func CreateEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var entry Timeframe
	json.NewDecoder(r.Body).Decode(&entry)
	Instance.Create(&entry)
	json.NewEncoder(w).Encode(entry)
}

func GetEntryByID(w http.ResponseWriter, r *http.Request) {
	entryID := mux.Vars(r)["id"]
	if !checkIfEntryExists(entryID) {
		json.NewEncoder(w).Encode("Entry Not Found!")
		return
	}
	var entry Timeframe
	Instance.First(&entry, entryID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func GetEntries(w http.ResponseWriter, r *http.Request) {
	var entries []Timeframe
	Instance.Find(&entries)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

func UpdateEntry(w http.ResponseWriter, r *http.Request) {
	entryID := mux.Vars(r)["id"]
	if !checkIfEntryExists(entryID) {
		json.NewEncoder(w).Encode("Entry Not Found!")
		return
	}
	var entry Timeframe
	Instance.First(&entry, entryID)
	json.NewDecoder(r.Body).Decode(&entry)
	Instance.Save(&entry)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func DeleteEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	entryID := mux.Vars(r)["id"]
	if !checkIfEntryExists(entryID) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Entry Not Found!")
		return
	}
	var entry Timeframe
	Instance.Delete(&entry, entryID)
	json.NewEncoder(w).Encode("Entry Deleted Succesfully!")
}
