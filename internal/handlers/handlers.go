package handlers

import (
	"fmt"
	"local/timetracker/internal/components"
	"local/timetracker/internal/database"
	"log"
	"net/http"
	"time"

	"golang.design/x/clipboard"
)

var (
	err         error
	tfList      []database.Timeframe
	projectList = []database.Project{{
		ID:       0,
		Name:     "NotAssigned",
		Activity: "",
		Details:  "",
	}}
	globalID           int
	globalIDProject    int       = 1
	activeDate         string    = time.Now().Format("2006-01-02")
	activeMonthSummary time.Time = time.Now()
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	component := components.HomePage(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		activeDate,
		workTotalByDate(activeDate).String(),
		workDeltaWeek(workTotalWeek(activeDate), activeDate).String())
	component.Render(r.Context(), w)
}

func RecordsPageHandler(w http.ResponseWriter, r *http.Request) {
	component := components.Records(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		activeDate,
		workTotalByDate(activeDate).String(),
		workDeltaWeek(workTotalWeek(activeDate), activeDate).String())
	component.Render(r.Context(), w)
}

func RecordsHandler(w http.ResponseWriter, r *http.Request) {
	component := components.RecordList(
		database.GetRecordsForDate(activeDate),
		database.GetProjects())
	component.Render(r.Context(), w)
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	components.Projects(database.GetProjects()).Render(r.Context(), w)
}

func MonthlySummaryHandler(w http.ResponseWriter, r *http.Request) {
	components.MonthlySummary(activeMonthSummary, GetProjectHoursMonth(activeMonthSummary), GetWorkDays(activeMonthSummary)).Render(r.Context(), w)
}

func YearlySummaryHandler(w http.ResponseWriter, r *http.Request) {
	components.YearlySummary(time.Time{}, GetProjectHoursYear(time.Now()), 0).Render(r.Context(), w)
}

func ChangeDate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activeDate = r.FormValue("dateofrecord")
	log.Printf("Date changed to %s\n", activeDate)
	RecordsPageHandler(w, r)
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:        database.GetRecordsMaxID() + 1,
		Date:      r.FormValue("dateofrecord"),
		Year:      year,
		Month:     month,
		Day:       day,
		Start:     r.FormValue("start"),
		End:       r.FormValue("end"),
		Duration:  "",
		ProjectID: atoi(r.FormValue("project")),
	}
	globalID += 1
	err = database.CreateRecord(timeframe)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created Record %s %s ID: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID)
	RecordsPageHandler(w, r)
}

func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	timefr := database.Timeframe{
		ID:        atoi(id),
		Date:      r.FormValue("dateofrecord"),
		Year:      year,
		Month:     month,
		Day:       day,
		Start:     r.FormValue("start"),
		End:       r.FormValue("end"),
		Duration:  "",
		ProjectID: atoi(r.FormValue("project")),
	}
	err = database.UpdateRecord(timefr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Update Record with ID %s\n", id)
	RecordsPageHandler(w, r)
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteRecord(atoi(id))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Remove Record with ID: %v\n", id)
	RecordsPageHandler(w, r)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var project = database.Project{
		ID:       database.GetProjectsMaxID() + 1,
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	globalIDProject += 1
	err = database.CreateProject(project)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Create Project %s %s %s\n", project.Activity, project.Details, project.Name)
	ProjectsHandler(w, r)
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.PathValue("id")
	project := database.Project{
		ID:       atoi(id),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	err = database.UpdateProject(project)
	log.Printf("Update Project with ID %s\n", id)
	ProjectsHandler(w, r)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteProject(atoi(id))
	log.Printf("Remove Project with ID: %v\n", id)
	ProjectsHandler(w, r)
}

func MonthlySummaryChangeMonth(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activeMonthSummary, err = time.Parse("2006-Jan", fmt.Sprintf("%s-%s", r.FormValue("year"), r.FormValue("month")[:3]))
	log.Println("Month changed to:", activeMonthSummary.Format("2006-01"))
	if err != nil {
		log.Println(err)
	}
	MonthlySummaryHandler(w, r)
}

func MonthlySummaryToClipboard(w http.ResponseWriter, r *http.Request) {
	projects := GetProjectHoursMonth(activeMonthSummary)
	out := ""
	for _, project := range projects[4 : len(projects)-1] {
		out += fmt.Sprintf("%s\t%s\t%s\n", project.Activity, project.Details, project.Hours)
	}
	// Init returns an error if the package is not ready for use.
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	clipboard.Write(clipboard.FmtText, []byte(out))
}
