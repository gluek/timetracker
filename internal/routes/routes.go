package routes

import (
	"net/http"

	"github.com/gluek/timetracker/internal/handlers"
)

func RegisterOtherRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/currentdate", handlers.ChangeDate)
	mux.HandleFunc("POST /api/currentdate/{button}", handlers.ChangeDate)
	mux.HandleFunc("GET /month", handlers.MonthlySummaryHandler)
	mux.HandleFunc("GET /year", handlers.YearlySummaryHandler)
	mux.HandleFunc("POST /api/monthlysummary", handlers.MonthlySummaryChangeMonth)
	mux.HandleFunc("GET /api/monthlysummary/export", handlers.MonthlySummaryDownload)
	mux.HandleFunc("POST /api/yearlysummary", handlers.YearlySummaryChangeYear)
	mux.HandleFunc("POST /api/clipboard", handlers.MonthlySummaryToClipboard)
	mux.HandleFunc("POST /api/quickbar/{name}", handlers.Quickbar)
	mux.HandleFunc("GET /vacation", handlers.PlannerPageHandler)
	mux.HandleFunc("POST /api/vacation/{date}", handlers.PlannerToggleVacationWhole)
	mux.HandleFunc("POST /api/vacationhalf/{date}", handlers.PlannerToggleVacationHalf)
	mux.HandleFunc("POST /api/vacation", handlers.PlannerChangeYear)
}

func RegisterRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /timeframes", handlers.RecordsPageHandler)
	mux.HandleFunc("GET /timeframes/{date}", handlers.RecordsPageDateChangeHandler)
	mux.HandleFunc("POST /api/timeframes", handlers.CreateRecord)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", handlers.DeleteRecord)
	mux.HandleFunc("PUT /api/timeframes/{id}", handlers.UpdateRecord)
}
func RegisterProjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /projects", handlers.ProjectsHandler)
	mux.HandleFunc("POST /api/projects", handlers.CreateProject)
	mux.HandleFunc("DELETE /api/projects/{id}/", handlers.DeleteProject)
	mux.HandleFunc("PUT /api/projects/{id}", handlers.UpdateProject)
}
func RegisterMockRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/timeframes", handlers.MockCreateRecord)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", handlers.MockDeleteRecord)
	mux.HandleFunc("PUT /api/timeframes/{id}", handlers.MockUpdateRecord)
}
func RegisterMockProjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/projects", handlers.MockCreateProject)
	mux.HandleFunc("DELETE /api/projects/{id}/", handlers.MockDeleteProject)
	mux.HandleFunc("PUT /api/projects/{id}", handlers.MockUpdateProject)
}
