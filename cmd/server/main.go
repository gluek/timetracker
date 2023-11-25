package main

import (
	"html/template"
	"local/timetracker/internal/database"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jchv/go-webview2"
)

var (
	pwd, _ = os.Getwd()
)

type PageData struct {
	Title   string
	Entries []database.Timeframe
}

func main() {
	go webServer()

	webView()
}

func webServer() {
	database.Connect()
	defer database.Close()

	router := mux.NewRouter()

	http.Handle("/", http.FileServer(http.Dir(pwd+"/internal/templates/")))
	http.Handle("/api/", router)
	//http.HandleFunc("/layout", TemplateHandler)
	RegisterEntryRoutes(router)

	// Windows may be missing this
	mime.AddExtensionType(".js", "application/javascript")

	log.Fatal(http.ListenAndServe("127.0.0.1:34115", nil))
}

func webView() {
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true, // To display the development tools
		AutoFocus: true,
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
	router.HandleFunc("/api/timeframes/{id}", database.GetEntryByID).Methods("GET")
	router.HandleFunc("/api/timeframes/{id}", database.UpdateEntry).Methods("PUT")
	router.HandleFunc("/api/timeframes/{id}", database.DeleteEntry).Methods("DELETE")
}

func TemplateHandler(w http.ResponseWriter, r *http.Request) {
	var entries []database.Timeframe
	//database.Instance.Find(&entries)
	tmpl := template.Must(template.ParseFiles(pwd + "/internal/templates/templtest.html"))
	data := PageData{
		Title:   "My Title",
		Entries: entries,
	}
	tmpl.Execute(w, data)
}
