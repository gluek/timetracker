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
	"strconv"
	"strings"
	"time"
)

//go:embed internal/assets/css/input.css
//go:embed internal/assets/js/htmx.min.js
//go:embed internal/assets/js/echarts.js
var content embed.FS

var (
	tfList   []database.Timeframe
	globalID int
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
	component := components.Page(tfList)
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
	RegisterMockEntryRoutes(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	mime.AddExtensionType(".js", "application/javascript")

	fmt.Println("Listening on localhost:34115")
	if err := http.ListenAndServe("localhost:34115", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}

func RegisterEntryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/timeframes", database.GetEntries)
	mux.HandleFunc("POST /api/timeframes", database.CreateEntry)
	mux.HandleFunc("GET /api/timeframes/{id}", database.GetEntryByID)
	mux.HandleFunc("PUT /api/timeframes/{id}", database.UpdateEntry)
	mux.HandleFunc("DELETE /api/timeframes/{id}", database.DeleteEntry)
}

func RegisterMockEntryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/timeframes", MockCreateEntry)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", MockDeleteEntry)
	mux.HandleFunc("PUT /api/timeframes/{id}", MockUpdateEntry)
	//mux.HandleFunc("GET /api/timeframes/{id}", database.GetEntryByID)
}

func atoi(s string) int {
	value, _ := strconv.Atoi(s)
	return value
}

func parseDate(date string) (int, int, int) {

	splitString := strings.Split(date, "-")
	year, month, day := splitString[0], splitString[1], splitString[2]
	return atoi(year), atoi(month), atoi(day)
}

func findID(id string) int {
	for index, timeframe := range tfList {
		if timeframe.ID == id {
			return index
		}
	}
	return -1
}

func MockCreateEntry(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:       strconv.Itoa(globalID),
		Date:     r.FormValue("dateofrecord"),
		Year:     year,
		Month:    month,
		Day:      day,
		Start:    r.FormValue("start"),
		End:      r.FormValue("end"),
		Duration: "",
		Project:  r.FormValue("project"),
	}
	globalID += 1
	tfList = append(tfList, timeframe)
	fmt.Printf("%s %s %s Len of tfList: %d\n", timeframe.Start, timeframe.End, timeframe.Project, len(tfList))
	components.Records(tfList).Render(r.Context(), w)
}

func MockUpdateEntry(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	index := findID(id)
	tfList[index] = database.Timeframe{
		ID:       id,
		Date:     r.FormValue("dateofrecord"),
		Year:     year,
		Month:    month,
		Day:      day,
		Start:    r.FormValue("start"),
		End:      r.FormValue("end"),
		Duration: "",
		Project:  r.FormValue("project"),
	}
	fmt.Printf("Entry with ID %s updated\n", id)
	components.Records(tfList).Render(r.Context(), w)
}

func MockDeleteEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findID(id)
	tfList = append(tfList[:index], tfList[index+1:]...)
	fmt.Printf("Remove ID: %v\n", id)
	components.Records(tfList).Render(r.Context(), w)
}
