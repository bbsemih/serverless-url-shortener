package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func main() {
	fmt.Println("---------------------")
	port, addr := os.Getenv("PORT"), os.Getenv("LISTEN_ADDR")
	if port == "" {
		port = "8080"
	}
	if addr == "" {
		addr = "localhost"
	}

	googleSheetsId := os.Getenv("GSHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")

	srv := &server{
		googleSheetsId: googleSheetsId,
		sheetName:      sheetName,
	}

	http.HandleFunc("/", srv.redirect)
	//combines host and port into a network address
	listenAddr := net.JoinHostPort(addr, port)
	log.Printf("Starting server at %s", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	log.Fatal(err)
}

type server struct {
	googleSheetsId string
	sheetName      string
}

func (s *server) redirect(w http.ResponseWriter, r *http.Request) {
	if s.googleSheetsId == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "No google sheets id provided!")
		return
	} else if s.sheetName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "No sheet name provided!")
		return
	}

	ctx := r.Context()

	srv, err := sheets.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	log.Println("querying the sheet!")
	readRange := "!A:B"
	resp, err := srv.Spreadsheets.Values.Get(s.googleSheetsId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	shortcuts := urlMap(resp.Values)
	log.Printf("parsed %d shortcuts", len(shortcuts))

	requestedPath := r.URL.Path
	redirTo := checkRedirect(shortcuts, requestedPath)
	if redirTo == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no shortcut found for %q", requestedPath)
		return
	}
	log.Printf("redirecting %q to %q", requestedPath, redirTo.String())
	http.Redirect(w, r, redirTo.String(), http.StatusMovedPermanently)
}

func checkRedirect(m map[string]*url.URL, path string) *url.URL {
	path = strings.TrimPrefix(path, "/")
	layers := strings.Split(path, "/")
	for len(layers) > 0 {
		query := strings.Join(layers, "/")
		v, ok := m[query]
		if ok {
			return v
		}
		layers = layers[:len(layers)-1]
	}
	return nil
}

//46

func urlMap(in [][]interface{}) map[string]*url.URL {
	out := make(map[string]*url.URL)
	for _, row := range in {
		if len(row) < 2 {
			continue
		}
		k, ok := row[0].(string)
		if !ok || k == "" {
			continue
		}
		v, ok := row[1].(string)
		if !ok || v == "" {
			continue
		}
		k = strings.ToLower(k)

		u, err := url.Parse(v)
		if err != nil {
			log.Printf("warn: %s=%s url invalid", k, v)
			continue
		}

		_, exists := out[k]
		if exists {
			log.Printf("warn: shortcut %q already declared, overwriting", k)
		}
		out[k] = u
	}
	return out
}
