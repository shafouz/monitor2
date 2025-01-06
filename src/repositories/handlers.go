package repositories

import (
	"encoding/json"
	"fmt"
	"html/template"
	database "monitor2/src/db"
	"monitor2/src/db/models"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Repos(w http.ResponseWriter, r *http.Request) {
	repos, err := database.DB.GetAllRepos()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	template, err := template.ParseFiles("static/templates/repos.html")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	err = template.ExecuteTemplate(w, "repos.html", repos)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}

func CreateRepo(w http.ResponseWriter, r *http.Request) {
	repos_path := os.Getenv("REPOS_PATH")

	url := r.PostFormValue("url")
	if len(url) == 0 {
		fmt.Fprint(w, "url can't be empty.")
    return
	}

	if strings.Contains("..", url) {
		fmt.Fprint(w, "No path traversal pls")
    return
	}

	files := r.PostFormValue("files")
  if len(files) == 0 || files == "[]" {
		fmt.Fprint(w, "files can't be empty.")
		return
  }

  var files_json []string;
  err := json.Unmarshal([]byte(files), &files_json)
  if err != nil { 
		fmt.Fprint(w, "files need to be a valid json array.")
		return
  }

	directory := repos_path + "/" + getRepoDir(url)
	remote := r.PostFormValue("remote")
	if len(remote) == 0 {
		remote = "origin"
	}

	repo := models.Repository{
		Url:          url,
		WatchedFiles: []byte(files),
		Directory:    directory,
		Remote:       remote,
	}

	err = database.DB.CreateRepository(repo)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	go gitClone(url, directory)
	fmt.Fprintf(w, "Repo created")
}

func UpdateRepo(w http.ResponseWriter, r *http.Request) {
	idRaw := r.PostFormValue("id")
	url := r.PostFormValue("url")
	directory := r.PostFormValue("directory")
	watchedFiles := r.PostFormValue("watched_files")
	remote := r.PostFormValue("remote")
	scheduleHoursRaw := r.PostFormValue("schedule_hours")
	deletedRaw := r.PostFormValue("deleted")

	if idRaw == "" || url == "" || directory == "" || watchedFiles == "" || remote == "" || scheduleHoursRaw == "" {
		fmt.Fprint(w, "All fields are required")
		return
	}

  if len(watchedFiles) == 0 || watchedFiles == "[]" {
		fmt.Fprint(w, "files can't be empty.")
		return
  }

  var files_json []string;
  err := json.Unmarshal([]byte(watchedFiles), &files_json)
  if err != nil { 
		fmt.Fprint(w, "files need to be a valid json array.")
		return
  }

	scheduleHours, err := strconv.Atoi(scheduleHoursRaw)
	if err != nil {
		fmt.Fprint(w, "Invalid id value")
		return
	}

	id, err := strconv.Atoi(idRaw)
	if err != nil {
		fmt.Fprint(w, "Invalid schedule_hours value")
		return
	}

	deleted := false
	if deletedRaw != "" {
		deleted, err = strconv.ParseBool(deletedRaw)
		if err != nil {
			fmt.Fprint(w, "Invalid deleted value")
			return
		}
	}

	repository := models.Repository{
		Url:           url,
		Directory:     directory,
		WatchedFiles:  []byte(watchedFiles),
		Remote:        remote,
		ScheduleHours: scheduleHours,
		Deleted:       deleted,
	}

	err = database.DB.UpdateRepository(id, repository)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	http.Redirect(w, r, "/repos", 303)
}
