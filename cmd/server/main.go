package main

import (
	"fmt"
	"local/timetracker/internal/database"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	database.Connect()
	database.Migrate()
	router := mux.NewRouter()

	http.Handle("/", http.FileServer(http.Dir("D:\\Dev\\Go\\TimeTracker\\internal\\templates\\")))
	http.Handle("/api/", router)
	RegisterEntryRoutes(router)

	// Windows may be missing this
	mime.AddExtensionType(".js", "application/javascript")
	//mime.AddExtensionType(".css", "text/css")

	log.Fatal(http.ListenAndServe("127.0.0.1:34115", nil))
}

func RegisterEntryRoutes(router *mux.Router) {
	router.HandleFunc("/api/hello", h1).Methods("GET")
	router.HandleFunc("/api/timeframes", database.CreateEntry).Methods("POST")
	router.HandleFunc("/api/timeframes", database.GetEntries).Methods("GET")
	router.HandleFunc("/api/timeframes/{id}", database.GetEntryByID).Methods("GET")
	router.HandleFunc("/api/timeframes/{id}", database.UpdateEntry).Methods("PUT")
	router.HandleFunc("/api/timeframes/{id}", database.DeleteEntry).Methods("DELETE")
}

func h1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Test!")
}
