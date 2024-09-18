package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	d "gt/data"
)

type errorType struct {
	ErrorCode int
	Message   string
}

var (
	artists []d.Artist      // A slice to store artist data
	artis   d.AutoGenerated // A variable to store auto-generated data
)

// handleError sends an error response to the client and logs the error
func handleError(w http.ResponseWriter, status int, msg string, err error) {
	http.Error(w, msg, status) // Send an error response
	if err != nil {
		fmt.Println(err) // Log the error
	}
}

func ErrorPages(w http.ResponseWriter, code int, message string) {
	// w.WriteHeader(code)
	t, err := template.ParseFiles("template/error.html")
	if err != nil {
		t.Execute(w, errorType{
			ErrorCode: 500,
			Message:   "internal server error",
		})
		return
	}
	t.Execute(w, errorType{ErrorCode: code, Message: message})
}

// SearchHandler handles search requests and returns search results in JSON format
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorPages(w, 500, "internal server error")
		return
	}
	domain := r.Header.Get("Sec-Fetch-Site")
	if domain != "same-origin" {
		ErrorPages(w, http.StatusNotFound, "access denied")
		return
	}

	// Check if there is a query string in the URL
	if r.URL.RawQuery == "" {
		ErrorPages(w, 404, "not found")
		return
	}
	var results []d.SearchResult
	q := strings.TrimSpace(r.URL.Query().Get("s"))
	if q == "" {
		// If query is empty, return an empty result
		if err := json.NewEncoder(w).Encode(results); err != nil {
			ErrorPages(w, 500, "internal server error")
		}
		return
	}
	query := strings.ToLower(q)

	// Loop through all artists to find matching names and add them to results
	for i, artist := range artists {
		if len(results) > 16 {
			// Stop if we have reached the limit of 16 results
			break
		}

		if i == 0 {
			// On the first iteration, check if artist names start with the query
			for _, ar := range artists {
				artistName2 := strings.ToLower(ar.Name)
				if strings.HasPrefix(artistName2, query) {
					results = append(results, d.SearchResult{
						Image: ar.Image,
						ID:    ar.ID,
						Name:  ar.Name,
						Type:  "artist/band",
					})
				}
			}
		}

		artistName := strings.ToLower(artist.Name)
		// Check if artist names contain the query
		if !strings.HasPrefix(artistName, query) && strings.Contains(artistName, query) {
			results = append(results, d.SearchResult{
				Image: artist.Image,
				ID:    artist.ID,
				Name:  artist.Name,
				Type:  "artist/band",
			})
		}

		// Check if the query matches the artist's first album name
		if strings.HasPrefix(strings.ToLower(artist.FirstAlbum), query) {
			results = append(results, d.SearchResult{
				ID:   artist.ID,
				Name: artist.FirstAlbum,
				Type: "FirstAlbum of " + artist.Name,
			})
		}

		C_Date := strconv.Itoa(artist.CreationDate)
		// Check if the query matches the artist's creation date
		if strings.HasPrefix(strings.ToLower(C_Date), query) {
			results = append(results, d.SearchResult{
				ID:   artist.ID,
				Name: C_Date,
				Type: "Creation Date of " + artist.Name,
			})
		}
	}

	// Loop through all artists again to find matching members and add them to results
	for i, artist := range artists {
		if len(results) > 16 {
			// Stop if we have reached the limit of 16 results
			break
		}

		if i == 0 {
			// On the first iteration, check if any member names start with the query
			for _, ar := range artists {
				for _, member := range ar.Members {
					artistName2 := strings.ToLower(member)
					if strings.HasPrefix(artistName2, query) {
						results = append(results, d.SearchResult{
							Image: ar.Image,
							ID:    ar.ID,
							Name:  member,
							Type:  "member of " + ar.Name,
						})
					}
				}
			}
		}

		// Check if any member names contain the query
		for _, member := range artist.Members {
			if !strings.HasPrefix(strings.ToLower(member), query) && strings.Contains(strings.ToLower(member), query) {
				results = append(results, d.SearchResult{
					ID:   artist.ID,
					Name: member,
					Type: "member of " + artist.Name,
				})
			}
		}
	}

	// Loop through all location indices to find matching locations and add them to results
	j := 0
	for _, loc := range artis.Index {
		name := artists[j].Name
		for _, lo := range loc.Locations {
			// Check if the query matches any location name
			if strings.Contains(strings.ToLower(lo), query) {
				if len(results) < 16 {
					results = append(results, d.SearchResult{
						ID:   loc.ID,
						Name: lo,
						Type: "location " + name,
					})
				}
			}
		}
		j++
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		ErrorPages(w, 500, "internal server error")
	}
}

// Home handles requests to the home page
func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorPages(w, 404, "page not found")
		return
	}
	if r.Method != http.MethodGet {
		ErrorPages(w, 405, "merthod not alowd")
		return
	}

	// Fetch artist data from the API
	if err := fetchAndDecode("https://groupietrackers.herokuapp.com/api/artists", &artists); err != nil {
		ErrorPages(w, 500, "internal server error")

		return
	}

	// Fetch location data from the API
	if err := fetchAndDecode("https://groupietrackers.herokuapp.com/api/locations", &artis); err != nil {
		ErrorPages(w, 500, "internal server error")

		return
	}

	// Parse the home page template
	tmp, err := template.ParseFiles("template/home_page.html")
	if err != nil {
		ErrorPages(w, 500, "internal server error")

		return
	}
	h := removeDuplicates(artis)

	viewData := d.ViewDat{
		Da:  artists,
		Loc: h,
	}
	// Execute the template with artist data
	if err := tmp.Execute(w, viewData); err != nil {
		ErrorPages(w, 500, "internal server error")
	}
}

func removeDuplicates(s d.AutoGenerated) []string {
	// Create a map to track unique strings
	uniqueMap := make(map[string]bool)
	var result []string

	// Loop over the input slice
	for _, locations := range s.Index {
		for _, str := range locations.Locations {
			// If the string is not already in the map, add it to the result
			if !uniqueMap[str] {
				uniqueMap[str] = true
				result = append(result, str)
			}
		}
	}
	return result
}

// Search handles search requests and displays results
func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorPages(w, 405, "method not alowd")

		return
	}
	var results []d.SearchResult
	q := strings.TrimSpace(r.URL.Query().Get("s"))
	if q == "" {
		// Redirect to home page if query is empty
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	query := strings.ToLower(q)

	// Loop through all artists to find matching names and add them to results
	for i, artist := range artists {
		if len(results) > 16 {
			// Stop if we have reached the limit of 16 results
			break
		}

		if i == 0 {
			// On the first iteration, check if artist names start with the query
			for _, ar := range artists {
				artistName2 := strings.ToLower(ar.Name)
				if strings.HasPrefix(artistName2, query) {
					results = append(results, d.SearchResult{
						Image: ar.Image,
						ID:    ar.ID,
						Name:  ar.Name,
						Type:  "artist/band",
					})
				}
			}
		}

		artistName := strings.ToLower(artist.Name)
		// Check if artist names contain the query
		if !strings.HasPrefix(artistName, query) && strings.Contains(artistName, query) {
			results = append(results, d.SearchResult{
				Image: artist.Image,
				ID:    artist.ID,
				Name:  artist.Name,
				Type:  "artist/band",
			})
		}

		// Check if the query matches the artist's first album name
		if strings.HasPrefix(strings.ToLower(artist.FirstAlbum), query) {
			results = append(results, d.SearchResult{
				Image: artist.Image,
				ID:    artist.ID,
				Name:  artist.FirstAlbum,
				Type:  "FirstAlbum of " + artist.Name,
			})
		}

		C_Date := strconv.Itoa(artist.CreationDate)
		// Check if the query matches the artist's creation date
		if strings.HasPrefix(strings.ToLower(C_Date), query) {
			results = append(results, d.SearchResult{
				Image: artist.Image,
				ID:    artist.ID,
				Name:  C_Date,
				Type:  "Creation Date of " + artist.Name,
			})
		}
	}

	// Loop through all artists again to find matching members and add them to results
	for i, artist := range artists {
		if len(results) > 16 {
			// Stop if we have reached the limit of 16 results
			break
		}
		if i == 0 {
			// On the first iteration, check if any member names start with the query
			for _, ar := range artists {
				for _, member := range ar.Members {
					artistName2 := strings.ToLower(member)
					if strings.HasPrefix(artistName2, query) {
						results = append(results, d.SearchResult{
							Image: ar.Image,
							ID:    ar.ID,
							Name:  member,
							Type:  "member of " + ar.Name,
						})
					}
				}
			}
		}

		// Check if any member names contain the query
		for _, member := range artist.Members {
			if !strings.HasPrefix(strings.ToLower(member), query) && strings.Contains(strings.ToLower(member), query) {
				results = append(results, d.SearchResult{
					Image: artist.Image,
					ID:    artist.ID,
					Name:  member,
					Type:  "member of " + artist.Name,
				})
			}
		}
	}

	// Loop through all location indices to find matching locations and add them to results
	j := 0
	for _, loc := range artis.Index {
		name := artists[j].Name
		for _, lo := range loc.Locations {
			// Check if the query matches any location name
			if strings.Contains(strings.ToLower(lo), query) {
				if len(results) < 16 {
					results = append(results, d.SearchResult{
						Image: artists[j].Image,
						ID:    loc.ID,
						Name:  lo,
						Type:  "location " + name,
					})
				}
			}
		}
		j++
	}

	// Attempt to parse the search results template
	tmp, err := template.ParseFiles("template/search.html")
	if err != nil {
		ErrorPages(w, 500, "internal server error")

		return
	}

	// Check if results are empty and handle accordingly
	if results == nil {
		// Attempt to parse the notfound template file
		tmp1, err := template.ParseFiles("template/notfound.html")
		if err != nil {
			// If template parsing fails, handle the error
			ErrorPages(w, 500, "internal server error")

			return
		}

		// Execute the notfound template
		err = tmp1.Execute(w, nil)
		if err != nil {
			ErrorPages(w, 500, "internal server error")
		}
		return
	}

	// Execute the search results template with the results
	if err := tmp.Execute(w, results); err != nil {
		ErrorPages(w, 500, "internal server error")
	}
}
