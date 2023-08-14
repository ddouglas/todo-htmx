package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type app struct {
	Items []string
}

var (
	this = app{
		Items: make([]string, 0),
	}
	lock = new(sync.Mutex)
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", handleHome).Name("home")
	router.HandleFunc("/todos", handleAddTodo).Methods("POST")

	this.Items = append(this.Items, "Hello There")

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", 8080),
		WriteTimeout: time.Second * 2,
		ReadTimeout:  time.Second * 2,
	}

	log.Fatal(srv.ListenAndServe())

}

var files = template.Must(findAndParseTemplates("./views", nil))

func handleHome(w http.ResponseWriter, r *http.Request) {

	err := files.ExecuteTemplate(w, "pages/home.html", this)
	if err != nil {
		fmt.Println("yeet", err)
	}

}

func handleAddTodo(w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("item")
	lock.Lock()
	defer lock.Unlock()
	this.Items = append(this.Items, data)
	err := files.ExecuteTemplate(w, "list-item", data)
	if err != nil {
		fmt.Println("yeet", err)
	}

}

func findAndParseTemplates(rootDir string, funcMap template.FuncMap) (*template.Template, error) {

	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	root := template.New("")

	err := filepath.Walk(cleanRoot, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			if err != nil {
				return err
			}

			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			name := path[pfx:]
			_, err = root.New(name).Funcs(funcMap).Parse(string(b))
			return err
		}

		return nil
	})

	return root, err

}
