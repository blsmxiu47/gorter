package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Name string `json:"name"`
}

type URLMap struct {
	urls map[string]string
}

type templateUpdate struct {
	Enhanced    bool
	EnhancedURL string
}

// At this moment there are 1017 apparently..
// Somehow seems like more than I remember..
// This must be what it feels like to get old
const MAX_SPECIES_ID = 1017

// App host name (and port, if applicable)
const MY_HOST = "localhost:8080"

// App home page rendering
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/enhance", http.StatusSeeOther)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/form.html"))
	w.Header().Set("Content-Type", "text/html")
	tmplUpdate := templateUpdate{
		Enhanced:    false,
		EnhancedURL: "",
	}
	err := tmpl.Execute(w, tmplUpdate)
	if err != nil {
		fmt.Print(err.Error())
	}
}

// Actually perform the URL enhancing and spit out html
// containing the new URL to the page
func (um *URLMap) handleEnhance(w http.ResponseWriter, r *http.Request) {
	inputURL := r.FormValue("url")
	if inputURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// generate a unique enhanced key for the original URL
	shortKey := generateShortKey()
	enhancedKey := enhanceText(shortKey)
	um.urls[enhancedKey] = inputURL

	// Construct the full enhanced URL
	enhancedURL := fmt.Sprintf("http://%s/enhanced/%s", MY_HOST, enhancedKey)

	// Render the HTML response with the enhanced URL
	tmpl := template.Must(template.ParseFiles("templates/form.html"))
	w.Header().Set("Content-Type", "text/html")
	// Fill in variables to insert into the web page
	tmplUpdate := templateUpdate{
		Enhanced:    true,
		EnhancedURL: enhancedURL,
	}
	err := tmpl.Execute(w, tmplUpdate)
	if err != nil {
		fmt.Print(err.Error())
	}
}

// URLMap method to handle redirection from enhanced URL to original
func (um *URLMap) handleRedirect(w http.ResponseWriter, r *http.Request) {
	enhancedKey := r.URL.Path[len("/enhanced/"):]
	if enhancedKey == "" {
		http.Error(w, "Enhanced key is missing", http.StatusBadRequest)
		return
	}
	originalURL, found := um.urls[enhancedKey]
	if !found {
		http.Error(w, "Enhanced key not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

// GET a single pokemon species name from
// the PokeAPI: https://pokeapi.co/docs/v2
func getSpeciesName(species_id int) string {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://pokeapi.co/api/v2/pokemon-species/%d", species_id), nil)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var responseObject Response
	json.Unmarshal(body, &responseObject)

	return responseObject.Name
}

// Insert anagramized pokemon species name into an
// existing string and shuffle
func enhanceText(text string) string {
	// Get pseudo-random name from the API
	rand.New(rand.NewSource(time.Now().UnixNano()))
	name := []byte(getSpeciesName(rand.Intn(MAX_SPECIES_ID)))

	// Concat the text and name as byte slice (doesn't matter where
	// name get inserted was this whole thing will get shuffled below)
	b := append(name, text[:]...)
	// Shuffle the letters
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})
	enhancedText := string(b)

	return enhancedText
}

// Create a new unique short key for the original URL
func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// sets length of the key. longer makes uniqueness more likely
	// with 62 characters, N_possibilities := 62**keyLength
	// with keyLength of 8, there are > 2.18e14 possiblities
	const keyLength = 8

	rand.New(rand.NewSource(time.Now().UnixNano()))
	// make allocates memory on the heap, slotting zero-values.
	// in this case a bytes slices (not array, mind you) of length keyLength
	shortKey := make([]byte, keyLength)
	// then for each element in the slice put in the character from charset at the index
	// that is a pseudo-random number in the interval [0, len(charset)]
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func main() {
	mux := http.NewServeMux()

	urls := &URLMap{
		urls: make(map[string]string),
	}

	// serve static files (aka css, at the moment)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/enhance", urls.handleEnhance)
	mux.HandleFunc("/enhanced/", urls.handleRedirect)

	// TODO: refactor for live app. using localhost now for testing
	fmt.Printf("URL Enhancer is running on http://%s\n", MY_HOST)
	http.ListenAndServe(":8080", mux)
}
