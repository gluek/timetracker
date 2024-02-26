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
	tfList      []database.Timeframe
	projectList = []database.Project{{
		ID:       "0",
		Name:     "",
		Activity: "",
		Details:  "",
	}}
	globalID        int
	globalIDProject int = 1
)

type PageData struct {
	Title   string
	Entries []database.Timeframe
}

func main() {
	// Init session
	pwd, _ := os.Getwd()
	fmt.Printf("%s\n", pwd)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/projects", projectsPageHandler)
	RegisterMockRecordRoutes(mux)
	RegisterMockProjectRoutes(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	mime.AddExtensionType(".js", "application/javascript")

	fmt.Println("Listening on http://localhost:34115")
	if err := http.ListenAndServe("localhost:34115", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
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

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	component := components.HomePage(tfList)
	component.Render(r.Context(), w)
}

func projectsPageHandler(w http.ResponseWriter, r *http.Request) {
	component := components.ProjectsPage(projectList)
	component.Render(r.Context(), w)
}

func recordsHandler(w http.ResponseWriter, r *http.Request) {
	components.Records(tfList).Render(r.Context(), w)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	components.Projects(projectList).Render(r.Context(), w)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	homepageHandler(w, r)
}

func RegisterEntryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/timeframes", database.GetEntries)
	mux.HandleFunc("POST /api/timeframes", database.CreateEntry)
	mux.HandleFunc("GET /api/timeframes/{id}", database.GetEntryByID)
	mux.HandleFunc("PUT /api/timeframes/{id}", database.UpdateEntry)
	mux.HandleFunc("DELETE /api/timeframes/{id}", database.DeleteEntry)
}

func RegisterMockRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/timeframes", MockCreateRecord)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", MockDeleteRecord)
	mux.HandleFunc("PUT /api/timeframes/{id}", MockUpdateRecord)
	//mux.HandleFunc("GET /api/timeframes/{id}", database.GetEntryByID)
}
func RegisterMockProjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/projects", MockCreateProject)
	mux.HandleFunc("DELETE /api/projects/{id}/", MockDeleteProject)
	mux.HandleFunc("PUT /api/projects/{id}", MockUpdateProject)
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

func findIDTimeframe(id string) int {
	for index, timeframe := range tfList {
		if timeframe.ID == id {
			return index
		}
	}
	return -1
}

func findIDProject(id string) int {
	for index, project := range projectList {
		if project.ID == id {
			return index
		}
	}
	return -1
}

func MockCreateRecord(w http.ResponseWriter, r *http.Request) {
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
	recordsHandler(w, r)
}

func MockUpdateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	index := findIDTimeframe(id)
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
	fmt.Printf("Record with ID %s updated\n", id)
	recordsHandler(w, r)
}

func MockDeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findIDTimeframe(id)
	tfList = append(tfList[:index], tfList[index+1:]...)
	fmt.Printf("Remove Record with ID: %v\n", id)
	recordsHandler(w, r)
}

func MockCreateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var project = database.Project{
		ID:       strconv.Itoa(globalIDProject),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	globalIDProject += 1
	projectList = append(projectList, project)
	fmt.Printf("%s %s %s Len of projectList: %d\n", project.Activity, project.Details, project.Name, len(projectList)-1)
	projectsHandler(w, r)
}

func MockUpdateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.PathValue("id")
	index := findIDProject(id)
	projectList[index] = database.Project{
		ID:       strconv.Itoa(globalIDProject),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	fmt.Printf("Project with ID %s updated\n", id)
	projectsHandler(w, r)
}

func MockDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findIDProject(id)
	projectList = append(projectList[:index], projectList[index+1:]...)
	fmt.Printf("Remove Project with ID: %v\n", id)
	projectsHandler(w, r)
}
