package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gluek/timetracker/internal/components"
	"github.com/gluek/timetracker/internal/database"

	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/de"
	"github.com/spf13/viper"
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
	globalIDProject    int                   = 1
	activeDate         time.Time             = time.Now()
	activeMonthSummary time.Time             = time.Now()
	activeYearSummary  time.Time             = time.Now()
	calendar           *cal.BusinessCalendar = cal.NewBusinessCalendar()
)

func HandlerInit() {
	calendar.AddHoliday(de.HolidaysNW...)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	component := components.HomePage(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		database.GetLocations(),
		activeDate,
		workTotalByDate(activeDate),
		workDeltaWeek(workTotalWeek(activeDate), activeDate), GetOvertimeHoursUntilDay(activeDate, activeDate),
		calendar.IsWorkday(activeDate))
	err := component.Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render homepage: %v", err)
	}
}

func RecordsPageHandler(w http.ResponseWriter, r *http.Request) {
	component := components.Records(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		database.GetLocations(),
		activeDate,
		workTotalByDate(activeDate),
		workDeltaWeek(workTotalWeek(activeDate), activeDate), GetOvertimeHoursUntilDay(activeDate, activeDate),
		calendar.IsWorkday(activeDate))
	err := component.Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render records: %v", err)
	}
}
func RecordsPageDateChangeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error change date records: %v", err)
		return
	}
	activeDate, err = time.Parse("2006-01-02", r.PathValue("date"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error change date: %v", err)
		return
	}
	component := components.Records(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		database.GetLocations(),
		activeDate,
		workTotalByDate(activeDate),
		workDeltaWeek(workTotalWeek(activeDate), activeDate), GetOvertimeHoursUntilDay(activeDate, activeDate),
		calendar.IsWorkday(activeDate))
	err = component.Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render records: %v", err)
		return
	}
}
func RecordsHandler(w http.ResponseWriter, r *http.Request) {
	component := components.RecordList(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		database.GetLocations(),
	)
	err := component.Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render records: %v", err)
		return
	}
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	err := components.Projects(database.GetProjects()).Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render projects: %v", err)
		return
	}
}

func MonthlySummaryHandler(w http.ResponseWriter, r *http.Request) {
	component := components.MonthlySummary(
		activeMonthSummary,
		GetProjectHoursMonth(activeMonthSummary),
		GetWorkDays(activeMonthSummary),
		GetProjectsHoursOverview(activeMonthSummary),
		database.GetLocationDaysForMonth(activeMonthSummary),
	)
	err := component.Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render monthly summary: %v", err)
		return
	}
}

func MonthlySummaryDownload(w http.ResponseWriter, r *http.Request) {
	projects := GetProjectsHoursOverview(activeMonthSummary)
	w.Header().Add("Content-Type", `text/csv; charset=utf-8`)
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_summary.csv", activeMonthSummary.Format("2006_01")))
	w.Write([]byte("Date;Hours;Projects;Workplaces\n"))
	for _, v := range projects {
		if v.Hours != "00.00" {
			w.Write([]byte(fmt.Sprintf("%s;%s;%s;%s\n", v.Date, v.Hours, v.Projects, v.Locations)))
		}
	}
}

func YearlySummaryHandler(w http.ResponseWriter, r *http.Request) {
	component := components.YearlySummary(
		activeYearSummary,
		GetProjectHoursYear(activeYearSummary),
		GetOvertimeHoursUntilDay(activeYearSummary, time.Now()),
		database.GetLocationDaysForYear(activeYearSummary),
	)
	err := component.Render(r.Context(), w)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error render yearly summary: %v", err)
		return
	}
}

func PlannerPageHandler(w http.ResponseWriter, r *http.Request) {
	vacations := database.GetRecordsForProjectAndYear(activeYearSummary, 2)
	entries := convertTimeframesForPlanner(vacations)
	component := components.PlannerPage(activeYearSummary, entries)
	err := component.Render(r.Context(), w)
	if err != nil {
		log.Printf("error render planner page: %v", err)
		return
	}
}

func ChangeDate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error change date: %v", err)
		return
	}
	if r.PathValue("button") == "left" {
		activeDate = activeDate.Add(-24 * time.Hour)
	} else if r.PathValue("button") == "right" {
		activeDate = activeDate.Add(24 * time.Hour)
	} else {
		activeDate, err = time.Parse("2006-01-02", r.FormValue("dateofrecord"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error change date: %v", err)
			return
		}
	}
	log.Printf("Date changed to %s\n", activeDate.Format("2006-01-02"))
	RecordsPageHandler(w, r)
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error create record: %v", err)
		return
	}
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:         0, // not used
		Date:       r.FormValue("dateofrecord"),
		Year:       year,
		Month:      month,
		Day:        day,
		Start:      r.FormValue("start"),
		End:        r.FormValue("end"),
		Duration:   "",
		ProjectID:  atoi(r.FormValue("project")),
		LocationID: atoi(r.FormValue("location")),
	}
	globalID += 1
	err = database.CreateRecord(timeframe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}

	log.Printf("Created Record - From: %s, To: %s, ProjectID: %d, LocationID: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID, timeframe.LocationID)
	RecordsPageHandler(w, r)
}

func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error update record: %v", err)
		return
	}
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	timefr := database.Timeframe{
		ID:         atoi(id),
		Date:       r.FormValue("dateofrecord"),
		Year:       year,
		Month:      month,
		Day:        day,
		Start:      r.FormValue("start"),
		End:        r.FormValue("end"),
		Duration:   "",
		ProjectID:  atoi(r.FormValue("project")),
		LocationID: atoi(r.FormValue("location")),
	}
	err = database.UpdateRecord(timefr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}
	log.Printf("Update Record with ID %s\n", id)
	RecordsPageHandler(w, r)
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteRecord(atoi(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}
	log.Printf("Remove Record with ID: %v\n", id)
	RecordsPageHandler(w, r)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error create project: %v", err)
		return
	}
	var project = database.Project{
		ID:       0, // not used
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
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error update project: %v", err)
		return
	}
	id := r.PathValue("id")
	project := database.Project{
		ID:       atoi(id),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	err = database.UpdateProject(project)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error update project: %v", err)
		return
	}
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
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error change month summary: %v", err)
		return
	}
	activeMonthSummary, err = time.Parse("2006-Jan", fmt.Sprintf("%s-%s", r.FormValue("year"), r.FormValue("month")[:3]))
	log.Println("Month changed to:", activeMonthSummary.Format("2006-01"))
	if err != nil {
		log.Println(err)
	}
	MonthlySummaryHandler(w, r)
}

func YearlySummaryChangeYear(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error change year summary: %v", err)
		return
	}
	activeYearSummary, err = time.Parse("2006", r.FormValue("year"))
	log.Println("Year changed to:", activeYearSummary.Format("2006"))
	if err != nil {
		log.Println(err)
	}
	YearlySummaryHandler(w, r)
}

func MonthlySummaryToClipboard(w http.ResponseWriter, r *http.Request) {
	projects := GetProjectHoursMonth(activeMonthSummary)
	out := ""
	for _, project := range projects[4 : len(projects)-1] {
		out += fmt.Sprintf("%s\t%s\t%s\n", project.Activity, project.Details, project.Hours)
	}
	separator := viper.GetString("decimal_separator")
	out = strings.ReplaceAll(out, ".", separator)
	// Init returns an error if the package is not ready for use.
	err := clipboard.Init()
	if err != nil {
		log.Fatal(err)
	}
	clipboard.Write(clipboard.FmtText, []byte(out))
}

func PlannerChangeYear(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error change year summary: %v", err)
		return
	}
	activeYearSummary, err = time.Parse("2006", r.FormValue("year"))
	log.Println("Year changed to:", activeYearSummary.Format("2006"))
	if err != nil {
		log.Println(err)
	}
	PlannerPageHandler(w, r)
}

func PlannerToggleVacation(w http.ResponseWriter, r *http.Request, divider float32) {
	clickedDate := r.PathValue("date")
	datetimeDate, _ := time.Parse("2006-01-02", clickedDate)
	vacationRecords := database.GetRecordsForProjectAndDate(datetimeDate, 2)
	if len(vacationRecords) > 0 {
		log.Println("Deleted Vacation Records:", len(vacationRecords))
		for _, record := range vacationRecords {
			database.DeleteRecord(record.ID)
		}
	} else {
		start, err := time.Parse("15:04", "08:00")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error parse time ToggleVacation: %v", err)
			return
		}
		end := start.Add(durationOneWorkday() / time.Duration(divider))
		var timeframe = database.Timeframe{
			ID:         0, // not used
			Date:       datetimeDate.Format("2006-01-02"),
			Year:       datetimeDate.Year(),
			Month:      int(datetimeDate.Month()),
			Day:        datetimeDate.Day(),
			Start:      start.Format("15:04"),
			End:        end.Format("15:04"),
			Duration:   "",
			ProjectID:  2,
			LocationID: 1,
		}
		err = database.CreateRecord(timeframe)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
		}
		log.Printf("Created Record - From: %s, To: %s, ProjectID: %d, LocationID: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID, timeframe.LocationID)
	}
	data := convertTimeframesForPlanner(database.GetRecordsForProjectAndYear(activeYearSummary, 2))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func PlannerToggleVacationWhole(w http.ResponseWriter, r *http.Request) {
	PlannerToggleVacation(w, r, 1)
}

func PlannerToggleVacationHalf(w http.ResponseWriter, r *http.Request) {
	PlannerToggleVacation(w, r, 2)
}

func Quickbar(w http.ResponseWriter, r *http.Request) {
	quickName := r.PathValue("name")
	switch quickName {
	case "vacation":
		start, err := time.Parse("15:04", "08:00")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error parse time Quickbar vacation: %v", err)
			return
		}
		end := start.Add(durationOneWorkday())
		var timeframe = database.Timeframe{
			ID:         0, // not used
			Date:       activeDate.Format("2006-01-02"),
			Year:       activeDate.Year(),
			Month:      int(activeDate.Month()),
			Day:        activeDate.Day(),
			Start:      start.Format("15:04"),
			End:        end.Format("15:04"),
			Duration:   "",
			ProjectID:  2,
			LocationID: 1,
		}
		err = database.CreateRecord(timeframe)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
		}

		log.Printf("Created Record - From: %s, To: %s, ProjectID: %d, LocationID: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID, timeframe.LocationID)
		RecordsPageHandler(w, r)
	}
}
