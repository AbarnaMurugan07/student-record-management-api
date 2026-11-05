package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Student struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Year       int    `json:"year"`
}

var (
	students = make(map[int]Student)
	nextID   = 1
	mutex    sync.RWMutex
)

func getStudents(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	var s Student

	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	s.ID = nextID
	nextID++
	students[s.ID] = s
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func updateStudent(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))

	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := students[id]; !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	s.ID = id
	students[id] = s

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func deleteStudent(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := students[id]; !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	delete(students, id)
	fmt.Fprintln(w, "Student deleted successfully")
}

func getStudent(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))

	mutex.RLock()
	defer mutex.RUnlock()

	s, exists := students[id]
	if !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /students", getStudents)
	mux.HandleFunc("POST /students", createStudent)
	mux.HandleFunc("GET /students/{id}", getStudent)
	mux.HandleFunc("PUT /students/{id}", updateStudent)
	mux.HandleFunc("DELETE /students/{id}", deleteStudent)

	fmt.Println("Student Record Management API")
	fmt.Println("Server running at http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", loggingMiddleware(mux)))
}
