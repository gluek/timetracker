package main

import (
	"embed"
	"fmt"
	"local/timetracker/internal/components"
	"local/timetracker/internal/database"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"time"
)

//go:embed internal/assets/css/input.css
//go:embed internal/assets/js/htmx.min.js
//go:embed internal/assets/js/echarts.js
var content embed.FS

var (
	pwd, _ = os.Getwd()
	err    error
)

type PageData struct {
	Title   string
	Entries []database.Timeframe
}

func randomFloats() []float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	floatSlice := []float64{}
	count := 6
	for i := 0; i < count; i++ {
		floatSlice = append(floatSlice, r.Float64())
	}
	return floatSlice
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	component := components.Page(randomFloats())
	component.Render(r.Context(), w)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	getHandler(w, r)
}

func main() {
	// Init session
	pwd, _ := os.Getwd()
	fmt.Printf("%s\n", pwd)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	mime.AddExtensionType(".js", "application/javascript")

	fmt.Println("Listening on localhost:34115")
	if err := http.ListenAndServe("localhost:34115", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}

func RegisterEntryRoutes(mux http.ServeMux) {
	mux.HandleFunc("GET /api/timeframes", database.GetEntries)
	mux.HandleFunc("POST /api/timeframes", database.CreateEntry)
	mux.HandleFunc("GET /api/timeframes/{id}", database.GetEntryByID)
	mux.HandleFunc("PUT /api/timeframes/{id}", database.UpdateEntry)
	mux.HandleFunc("DELETE /api/timeframes/{id}", database.DeleteEntry)
}
