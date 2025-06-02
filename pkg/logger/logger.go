package logger

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

type JustProjectNameFormatter struct {
	log.TextFormatter
}

func (f *JustProjectNameFormatter) Format(entry *log.Entry) ([]byte, error) {
	if entry.HasCaller() {
		_, file := filepath.Split(entry.Caller.File)
		funcName := entry.Caller.Function
		parts := strings.Split(funcName, ".")
		if len(parts) > 1 {
			funcName = parts[len(parts)-2]
		}
		entry.Caller.File = "    " + file
		entry.Caller.Function = funcName
	}

	f.TextFormatter.PadLevelText = true
	f.TextFormatter.QuoteEmptyFields = true
	f.TextFormatter.DisableQuote = true
	f.TextFormatter.ForceColors = true
	f.TextFormatter.EnvironmentOverrideColors = true
	f.TextFormatter.DisableTimestamp = false
	f.TextFormatter.TimestampFormat = "2006-01-02T15:04:05-07:00"
	f.TextFormatter.DisableSorting = true

	serialized, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	str := string(serialized)
	str = strings.ReplaceAll(str, "                               ", " ")

	return []byte(str), nil
}

func Init(env string) {
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)

	log.SetFormatter(&JustProjectNameFormatter{
		TextFormatter: log.TextFormatter{
			FullTimestamp: true,
			DisableColors: false,
		},
	})

	if env == "production" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
}

func statusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\033[32m" // green
	case status >= 300 && status < 400:
		return "\033[33m" // yellow
	case status >= 400 && status < 500:
		return "\033[31m" // red
	default:
		return "\033[31m" // red
	}
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		status := ww.Status()
		color := statusColor(status)

		log.Infof("%s %s - %s%d\033[0m in %s%s\033[0m from %s",
			r.Method,
			r.URL.Path,
			color,
			status,
			color,
			duration.Round(time.Millisecond),
			r.UserAgent(),
		)
	})
}
