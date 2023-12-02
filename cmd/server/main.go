package main

import (
	"fmt"
	"html/template"
	"local/timetracker/internal/database"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jchv/go-webview2"
)

var (
	pwd, _ = os.Getwd()
	err    error
)

type PageData struct {
	Title   string
	Entries []database.Timeframe
}

func main() {
	webServer()

	//webView()
}

func webServer() {
	database.Connect()
	defer database.Close()

	router := mux.NewRouter()

	http.Handle("/tailwind.css", http.FileServer(http.Dir(pwd+"/internal/templates/")))
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/layout", TemplateHandler)
	http.Handle("/api/", router)
	RegisterEntryRoutes(router)

	// Windows may be missing this
	mime.AddExtensionType(".js", "application/javascript")

	log.Fatal(http.ListenAndServe("127.0.0.1:34115", nil))
}

func webView() {
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true, // To display the development tools
		AutoFocus: true,
		DataPath:  filepath.Join(pwd, "/webview"),
		WindowOptions: webview2.WindowOptions{
			Title:  "Time Tracker",
			Width:  200,
			Height: 200,
			IconId: 2, // icon resource id
			Center: true,
		},
	})
	if w == nil {
		log.Fatalln("Failed to load webview.")
	}
	defer w.Destroy()

	w.SetSize(600, 600, webview2.HintNone)
	w.Navigate("http://localhost:34115/")
	w.Run()
}

func RegisterEntryRoutes(router *mux.Router) {
	router.HandleFunc("/api/timeframes", database.GetEntries).Methods("GET")

	router.HandleFunc("/api/timeframes", database.CreateEntry).Methods("POST")
	router.HandleFunc("/api/form", TestParse).Methods("POST")
	router.HandleFunc("/api/timeframes/{id}", database.GetEntryByID).Methods("GET")
	router.HandleFunc("/api/timeframes/{id}", database.UpdateEntry).Methods("PUT")
	router.HandleFunc("/api/timeframes/{id}", database.DeleteEntry).Methods("DELETE")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		filepath.Join(pwd, "/internal/templates/inputbox.tmpl"),
		filepath.Join(pwd, "/internal/templates/index.tmpl"),
		filepath.Join(pwd, "/internal/templates/nav.tmpl"),
	}

	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func TemplateHandler(w http.ResponseWriter, r *http.Request) {
	var entries []database.Timeframe

	statement, err := database.DB.Prepare("SELECT * FROM timeframes")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := statement.Query()

	for rows.Next() {
		timefr := database.Timeframe{}
		rows.Scan(&timefr.ID, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.Project)
		timefr.Start = convertTimeStr(timefr.Start)
		timefr.End = convertTimeStr(timefr.End)
		entries = append(entries, timefr)
	}

	paths := []string{
		filepath.Join(pwd, "/internal/templates/inputbox.tmpl"),
		filepath.Join(pwd, "/internal/templates/index.tmpl"),
		filepath.Join(pwd, "/internal/templates/nav.tmpl"),
		filepath.Join(pwd, "/internal/templates/listentries.tmpl"),
	}

	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		log.Fatal(err)
	}

	data := PageData{
		Title:   "My Title",
		Entries: entries,
	}
	tmpl.ExecuteTemplate(w, "test", data)
}

func TestParse(w http.ResponseWriter, r *http.Request) {
	err = r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	project := r.PostFormValue("project")
	start := r.PostFormValue("start")
	end := r.PostFormValue("end")
	date := r.PostFormValue("currentdate")

	fmt.Printf("%s, %s, %s - Project: %s!\n", date, start, end, project)

	paths := []string{
		filepath.Join(pwd, "/internal/templates/inputbox.tmpl"),
		filepath.Join(pwd, "/internal/templates/index.tmpl"),
		filepath.Join(pwd, "/internal/templates/nav.tmpl"),
	}

	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func convertTimeStr(time string) string {
	return strings.Replace(strings.Replace(time, "h", ":", 1), "m", "", 1)
}
