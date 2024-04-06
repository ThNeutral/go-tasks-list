package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const (
	port = ":8081"

	indexPath   = "./templates/index.html"
	tasksPath   = "./templates/tasks.html"
	wrapperPath = "./templates/tasks-wrapper.html"
	storagePath = "./storage/tasks.json"
)

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

type Storage struct {
	Current []Task `json:"current"`
	Done    []Task `json:"done"`
}

func writeTaskToJSON(t Task) error {
	b, err := os.ReadFile(storagePath)
	if err != nil {
		return err
	}

	var storage Storage
	err = json.Unmarshal(b, &storage)
	if err != nil {
		return err
	}

	storage.Current = append(storage.Current, t)

	b, err = json.Marshal(storage)
	if err != nil {
		return err
	}

	err = os.Truncate(storagePath, 0)
	if err != nil {
		return err
	}

	err = os.WriteFile(storagePath, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getTasksFromJSON() (Storage, error) {
	b, err := os.ReadFile(storagePath)
	if err != nil {
		return Storage{}, err
	}

	var storage Storage
	err = json.Unmarshal(b, &storage)
	if err != nil {
		return Storage{}, err
	}

	return storage, nil
}

func deleteTaskFromJSON(id string) error {
	b, err := os.ReadFile(storagePath)
	if err != nil {
		return err
	}

	var storage Storage
	err = json.Unmarshal(b, &storage)
	if err != nil {
		return err
	}

	for i, t := range storage.Current {
		if t.Id == id {
			storage.Done = append(storage.Done, storage.Current[i])
			storage.Current = append(storage.Current[:i], storage.Current[i+1:]...)
		}
	}

	b, err = json.Marshal(storage)
	if err != nil {
		return err
	}

	err = os.Truncate(storagePath, 0)
	if err != nil {
		return err
	}

	err = os.WriteFile(storagePath, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	log.Println("Recieved request to /")
	tmpl, err := template.ParseFiles(indexPath, tasksPath)
	if err != nil {
		log.Println("Failed to parse files.", err.Error())
		return
	}

	st, err := getTasksFromJSON()
	if err != nil {
		log.Println("Failed to get task from JSON", err.Error())
		return
	}

	err = tmpl.Execute(w, st)
	if err != nil {
		log.Println("Failed to execute template.", err.Error())
		return
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	log.Println("Recieved request to /delete")
	id := r.URL.Query().Get("id")
	err := deleteTaskFromJSON(id)
	if err != nil {
		log.Println("Failed to delete task.", err.Error())
		return
	}

	tmpl, err := template.ParseFiles(wrapperPath, tasksPath)
	if err != nil {
		log.Println("Failed to parse files.", err.Error())
		return
	}

	st, err := getTasksFromJSON()
	if err != nil {
		log.Println("Failed to get task from JSON", err.Error())
		return
	}

	err = tmpl.Execute(w, st)
	if err != nil {
		log.Println("Failed to execute template.", err.Error())
		return
	}
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	log.Println("Recieved request to /create")
	t := Task{
		Id:          uuid.New().String(),
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	}

	err := writeTaskToJSON(t)
	if err != nil {
		log.Println("Failed to write task.", err.Error())
	}

	tmpl, err := template.ParseFiles(wrapperPath, tasksPath)
	if err != nil {
		log.Println("Failed to parse files.", err.Error())
		return
	}

	st, err := getTasksFromJSON()
	if err != nil {
		log.Println("Failed to get task from JSON", err.Error())
		return
	}

	err = tmpl.Execute(w, st)
	if err != nil {
		log.Println("Failed to execute template.", err.Error())
		return
	}
}

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/create", handleCreate)
	http.Handle("/resourses/", http.StripPrefix("/resourses", http.FileServer(http.Dir("./resourses"))))

	log.Println("Server listening on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Server failed.", err.Error())
	}
}
