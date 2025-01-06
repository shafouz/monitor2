package crawler

import (
	"fmt"
	database "monitor2/src/db"
	models "monitor2/src/db/models"
	"net/http"
	"strconv"
	"html/template"
)

func RunEndpoint(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")

	if len(url) == 0 {
		fmt.Fprintf(w, "Missing 'url' param")
		return
	}

	endpoint, err := database.DB.GetEndpointByUrl(url)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	err = RunSingle(&endpoint)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	err = database.DB.UpdateEndpointByUrl(endpoint, true)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "Done: %+v\n", url)
}

func CreateEndpoint(w http.ResponseWriter, r *http.Request) {
  err := r.ParseForm()
  if err != nil { fmt.Fprint(w, err) }

  url := r.PostFormValue("url")
  // validate that it is a valid html selector
	selector := r.PostFormValue("selector")
	profile := r.PostFormValue("profile")

	if len(profile) == 0 {
		fmt.Fprintf(w, "Missing 'profile' param")
		return
	}

	if !(profile == "js" || profile == "html") {
		fmt.Fprintf(w, "Current profiles: js, html")
		return
	}

	if profile == "html" && len(selector) == 0 {
		fmt.Fprintf(w, "Missing 'selector' param")
		return
	}

	if len(url) == 0 {
		fmt.Fprintf(w, "Missing 'url' param")
		return
	}

	endpoint := models.Endpoint{
		Url:      url,
		Selector: selector,
		Profile:  profile,
	}

	err = database.DB.CreateEndpoint(endpoint)
	if err != nil {
		fmt.Fprintf(w, "on create: %+v\n", err)
		return
	}

	err = RunSingle(&endpoint)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	err = database.DB.UpdateEndpointByUrl(endpoint, true)
	if err != nil {
		fmt.Fprintf(w, "on update: %+v\n", err)
		return
	}
	fmt.Fprintf(w, "Done: %+v\n", url)
}

func UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	scheduleHoursRaw := r.PostFormValue("schedule_hours")
	selector := r.PostFormValue("selector")
	profile := r.PostFormValue("profile")
	deletedRaw := r.PostFormValue("deleted")

	if url == "" {
		fmt.Fprint(w, "All fields are required")
		return
	}

	scheduleHours, err := strconv.Atoi(scheduleHoursRaw)
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

	endpoint := models.Endpoint{
		Url:                  url,
		ScheduleHours:        scheduleHours,
		Selector:             selector,
		Profile:              profile,
		Deleted:              deleted,
	}

	err = database.DB.UpdateEndpointByUrl(endpoint, false)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	http.Redirect(w, r, "/crawl", 303)
}

func DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	if len(url) == 0 {
		fmt.Fprintf(w, "Missing 'url' param")
		return
	}

	endpoint := models.Endpoint{
		Url: url,
	}

	rows_affected, err := database.DB.DeleteEndpoint(endpoint)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	fmt.Fprintf(w, "Rows affected: %+v\n", rows_affected)
}

func Endpoints(w http.ResponseWriter, r *http.Request) {
	repos, err := database.DB.GetAllEndpoints()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	template, err := template.ParseFiles("static/templates/endpoints.html")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	err = template.ExecuteTemplate(w, "endpoints.html", repos)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}
