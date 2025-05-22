package main

import (
	"driftGo/api"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

type JustProjectNameFormatter struct {
	log.TextFormatter
}

func (f *JustProjectNameFormatter) Format(entry *log.Entry) ([]byte, error) {
	if entry.HasCaller() {
		entry.Caller.File = "    driftGo"
	}
	return f.TextFormatter.Format(entry)
}

func main() {

	log.SetReportCaller(true)
	var r *chi.Mux = chi.NewRouter()

	log.SetFormatter(&JustProjectNameFormatter{
		TextFormatter: log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "01/02 15:04:05",
		},
	})

	api.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println(`
     __    _ _____  _____   
 ___/ /___(_) _/ /_/ ___/__ 
/ _  / __/ / _/ __/ (_ / _ \
\_,_/_/ /_/_/ \__/\___/\___/                                  
   `)

	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
