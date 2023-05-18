package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/api/sheets/v4"
)

func main() {
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

	if len(resp.Values) == 0 {
		log.Println("No data found.")
	} else {
		log.Printf("found %d rows of data", len(resp.Values))
		for _, row := range resp.Values {
			fmt.Printf("%#v\n", row)
		}
	}
}
