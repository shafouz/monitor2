package diffs

import (
	"fmt"
	"html/template"
	"net/http"

	database "monitor2/src/db"

	"github.com/gorilla/mux"
)

func Diffs(w http.ResponseWriter, r *http.Request) {
	diffs, err := database.DB.GetAllDiffs()

	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	template, err := template.ParseFiles("static/templates/diffs.html")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	err = template.ExecuteTemplate(w, "diffs.html", diffs)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}

func Diff(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	diff, err := database.DB.GetDiff(id)

	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	template, err := template.ParseFiles("static/templates/diff.html")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	github_url := fmt.Sprintf("%s/commit/%s", diff.Url, diff.Commit)
	err = template.ExecuteTemplate(w, "diff.html", map[string]string{"body": diff.Body, "github_url": github_url})
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}
