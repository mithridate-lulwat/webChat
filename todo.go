package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const timeFormat = "2006-01-02"

type Todo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Creation    string `json:"created_at,omitempty"`
	Due         string `json:"due_date,omitempty"`
}

const filename = "tasklist.txt"

type TodoList []Todo

func (t *TodoList) save() error {
	data, err := json.MarshalIndent(t, "", "	")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	return ioutil.WriteFile(filename, []byte(data), 0600)
}

func loadTodoList() TodoList {
	var todos TodoList
	// if the file doesn't exist, the tasklist is empty
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return todos
	} else {
		// otherwise we process the contained data
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Reading tasklist failed : #{err}")
		}
		if err = json.Unmarshal(data, &todos); err != nil {
			log.Fatalf("JSON unmarshaling failed: %s", err)
		}
		return todos
	}

}

func addTodo(list TodoList, t Todo) TodoList {
	list = append(list, t)
	return list
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	// it is only used to display the form, treatment is made in the saveHandler
	todos := loadTodoList()
	renderTemplate(w, "add", &todos)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// display the current tasklist
	todos := loadTodoList()
	renderTemplate(w, "view", &todos)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	nowString := time.Now().Format(timeFormat)
	now, _ := time.Parse(timeFormat, nowString)
	creation := now.Format(timeFormat)
	description := r.FormValue("description")
	dueString := r.FormValue("due")
	dueDate, _ := time.Parse(timeFormat, dueString)
	if dueDate.Before(now) { // Check if the dueDate makes sense
		http.Redirect(w, r, "/add/", http.StatusFound)
	} else {
		todo := Todo{title, description, creation, dueString}
		todos := loadTodoList()
		todos = addTodo(todos, todo)
		todos.save()
		http.Redirect(w, r, "/view/", http.StatusFound)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, t *TodoList) {
	templates := template.Must(template.ParseFiles("view.html", "add.html"))
	err := templates.ExecuteTemplate(w, tmpl+".html", t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
