package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS todos (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        task TEXT,
        done BOOLEAN
    );`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, task, done FROM todos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Task, &t.Done); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	var t Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO todos(task, done) VALUES(?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Task, t.Done)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func removeTodo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/removeTodo/"):]
	stmt, err := db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	initDB()
	http.HandleFunc("/todos", getTodos)
	http.HandleFunc("/addTodo", addTodo)
	http.HandleFunc("/removeTodo/", removeTodo)

	// Serve static files
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
