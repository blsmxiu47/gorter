package main

import (
	"fmt"
	"html/template"
	"io"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// type for our REST API respose(s)
type Response struct {
	Name string `json:"name"`
}

// struct to contain urls, which is simply a map containing 
// enhancedKeys as keys and original URLs as values
type URLMap struct {
    urls map[string]string
}

// struct for injecting variables into template HTML
type templateUpdate struct {
	enhanced bool
	enhancedURL string
}

// At this moment there are 1017 apparently.. 
// Somehow seems like more than I remember..
// This must be what it feels like to get old
const MAX_SPECIES_ID = 1017

// Function to GET a single pokemon species name from 
// the PokeAPI: https://pokeapi.co/docs/v2
func getSpeciesName(species_id int) string {
	client := &http.Client{}
	
	req, err := http.NewRequest("GET", fmt.Sprintf("https://pokeapi.co/api/v2/pokemon-species/%d"), nil)
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
	// TODO: remove print
	fmt.Printf("Name returned: %s", responseObject.Name)
	
	return responseObject.Name
}

// Function to insert anagramized pokemon species name into an 
// existing string and shuffle
func enhanceText(text string) string {
	// Get pseudo-random name from the API
	rand.New(rand.NewSource(time.Now().UnixNano()))
	name := []byte(getSpeciesName(rand.Intn(MAX_SPECIES_ID)))

	// Concat the text and name as byte slice (doesn't matter where 
	// name get inserted was this whole thing will get shuffled below)
	b := append(name, text[:]...)
	// Shuffle the letters
	// math//rand has a pseudo-random shuffling function
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})
	// and back to string
	enhancedText := string(b)

	return enhancedText
}


// Function to actually perform the URL shortening and spit out html
// containing the new URL to the page
// args: http response writer, http request
// returns: HTML string to be rendered, including link to the enhanced URL
func enhanceURL(inputURL string) string {
	// generate a unique enhanced key for the original URL
	shortKey := generateShortKey()
	// enhance the shortKey
	// TODO: implement enhanceShortKey()
	enhancedKey := enhanceText(shortKey)
	// assign item in urls map, enhancedKey: originalURL

	// Construct the full enhanced URL
	// TODO: replace with live version
	enhancedURL := fmt.Sprintf("http://localhost:8080/enhanced/%s", enhancedKey)

	// Render the HTML response with the enhanced URL
	responseHTML := fmt.Sprintf(`
        	<p>Better URL: <a href="%s">%s</a></p>
	`, enhancedURL, enhancedURL)

	return responseHTML
}

// URLMap method to handle redirection from enhanced URL to original
func (ue *URLMap) handleRedirect(w http.ResponseWriter, r *http.Request) {
	// get enhanced key portion of request url path
	enhancedKey := r.URL.Path[len("/enhanced/"):]
	// if empty string, error out
	if enhancedKey == "" {
		http.Error(w, "Enhanced key is missing", http.StatusBadRequest)
		return
	}

	// Retrieve the original URL from the `urls` map using the enhanced key
	originalURL, found := ue.urls[enhancedKey]
	// if not found in map, error out
	if !found {
		http.Error(w, "Enhanced key not found", http.StatusNotFound)
		return
	}

	// and redirect the user to the original URL
	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

// function to create a new unique enhanced key for the original URL
func generateShortKey() string {
	// alphanumerics
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// sets length of the key. longer makes uniqueness more likely
	// with 62 characters, N_possibilities := 62**keyLength
	// with keyLength of 8, there are > 2.18e14 possiblities 
	const keyLength = 8

	// New generates a new Rand struct.
	// NewSource returns a new psuedo-random Source seeded with the given value.
	// our seed is the current time in seconds in the Unix epoch
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// make allocates memory on the heap, slotting zero-values.
	// in this case a bytes slices (not array, mind you) of length keyLength
	enhancedKey := make([]byte, keyLength)
	// then for each element in the slice put in 
	// the character from charset at the index 
	// that is a pseudo-random number in the
	// interval [0, len(charset))
	for i := range enhancedKey {
		enhancedKey[i] = charset[rand.Intn(len(charset))]
	}
	// and return the filled byte slice enhancedKey as (joined) a string
	return string(enhancedKey)
}

func main() {
	// pointer to URLMap, allocating memory for urls
	urlMap := &URLMap{
		urls: make(map[string]string),
	}
	
	// base web app
	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}

		// grab URL from input form when entered by user
		inputURL := r.FormValue("url")

		// initiate URL enhancing
		enhancedURL := enhanceURL(inputURL)
		urlMap.urls[enhancedURL] = inputURL 

		// fill in variables to insert into the web page
		tmplUpdate := templateUpdate{
			enhanced: true,
			enhancedURL: enhancedURL,
		}
		tmpl.Execute(w, tmplUpdate)
	})
	
	// handle URL enhancing and redirection via methods defined above
	// http.HandleFunc("/enhance/", enhancer.HandleRedirect)

	// TODO: refactor for live app. using localhost now for testing
	fmt.Println("URL Enhancer is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
