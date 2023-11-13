package main

import (
	"local/timetracker/database"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	database.Connect()
	database.Migrate()

	router := mux.NewRouter().StrictSlash(true)

	RegisterEntryRoutes(router)

	// Windows may be missing this
	mime.AddExtensionType(".js", "application/javascript")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func RegisterEntryRoutes(router *mux.Router) {
	router.Handle("/", http.FileServer(http.Dir("frontend\\build")))
	router.HandleFunc("/api/timeframes", database.CreateEntry).Methods("POST")
	router.HandleFunc("/api/timeframes", database.GetEntries).Methods("GET")
	router.HandleFunc("/api/timeframes/{id}", database.GetEntryByID).Methods("GET")
	router.HandleFunc("/api/timeframes/{id}", database.UpdateEntry).Methods("PUT")
	router.HandleFunc("/api/timeframes/{id}", database.DeleteEntry).Methods("DELETE")
}
