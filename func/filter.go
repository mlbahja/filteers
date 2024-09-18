package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	d "gt/data"
)

func Filter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handleError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	if r.URL.Path != "/filter" {
		handleError(w, http.StatusNotFound, "page not found", nil)
		return
	}
	var results []d.Filter
	// Loop through all artists to find matching names and add them to results
	for i, artist := range artists {
		if Check_filter(r, artist.CreationDate, artist.Members, artist.FirstAlbum, i) {
			results = append(results, d.Filter{
				Image: artist.Image,
				ID:    artist.ID,
				Name:  artist.Name,
			})
		}
	}
	tmp, err := template.ParseFiles("template/Filter.html")
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal Server Error 500", err)
		return
	}

	if results == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := tmp.Execute(w, results); err != nil {
		handleError(w, http.StatusInternalServerError, "Internal Server Error 500", err)
	}
}

func Check_filter(r *http.Request, CreationDate int, Members []string, first_album string, i int) bool {
	creation_date_min, _ := strconv.Atoi(r.FormValue("creation_date_min"))
	creation_date_max, _ := strconv.Atoi(r.FormValue("creation_date_max"))
	first_album_date_min, _ := strconv.Atoi(r.FormValue("first_album_min"))
	first_album_date_max, _ := strconv.Atoi(r.FormValue("first_album_max"))

	// Parse member counts
	var selectedMembers []int
	for j := 1; j <= 8; j++ {
		if r.FormValue("num_members_"+strconv.Itoa(j)) != "" {
			selectedMembers = append(selectedMembers, j)
		}
	}

	locUK := r.FormValue("city")

	// Filter by creation date
	if r.FormValue("creation_date_min") != "" || r.FormValue("creation_date_max") != "" {
		if !(CreationDate >= creation_date_min && CreationDate <= creation_date_max) {
			return false
		}
	}

	// Filter by first album date
	if r.FormValue("first_album_min") != "" || r.FormValue("first_album_max") != "" {
		first_album_date, err := strconv.Atoi(first_album[6:]) // Assuming date at the end of the string
		if err != nil || !(first_album_date >= first_album_date_min && first_album_date <= first_album_date_max) {
			return false
		}
	}

	// Filter by number of members
	if len(selectedMembers) > 0 {
		if !Is_here(selectedMembers, len(Members)) {
			return false
		}
	}

	// Filter by location
	if locUK != "" {
		found := false
		for _, location := range artis.Index[i].Locations {
			if strings.HasSuffix(location, locUK) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// This function checks if the number of members matches the selected ones
func Is_here(selected []int, actual int) bool {
	for _, s := range selected {
		if s == actual {
			return true
		}
	}
	return false
}
