package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

const (
	FilesRoot       = "./files"
	FilesServerPath = "/files"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

// Initialize the SQLite database
func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "chat.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message TEXT
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}
	return db
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse() // parse the flags

	// Check if the directory exists
	if _, err := os.Stat(FilesRoot); os.IsNotExist(err) {
		// If the directory doesn't exist, create it
		err := os.Mkdir(FilesRoot, 0755)
		if err != nil {
			log.Fatal("Error creating directory:", err)
			return
		}
		fmt.Println("Files directory created successfully.")
	} else {
		fmt.Println("Found existing files directory.")
	}

	db := initDB()
	defer db.Close()

	r := newRoom(db)

	http.Handle(FilesServerPath+"/", http.StripPrefix(FilesServerPath, http.FileServer(http.Dir(FilesRoot))))
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
