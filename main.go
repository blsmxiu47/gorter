package main

import (
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

// struct to contain urls, which is simply a map containing 
// enhancedKeys as keys and original URLs as values
type URLEnhancer struct {
    urls map[string]string
}

// Method of URLEnhance struct to actually perform the enhancing.
// args: http response writer, http request
func (us *URLEnhancer) HandleEnhance(w http.ResponseWriter, r *http.Request) {
	// if request method is not POST, error out
	if r.Method != http.MethodPost {
		http.Error(w, "Inbalid request method", http.StatusMethodNotAllowed)
		return
	}

	// if url from request is missing, error out
	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// generate a unique enhanced key for the original URL
	enhancedKey := generateShortKey()
	// assign item in urls map, enhancedKey: originalURL
	us.urls[enhancedKey] = originalURL

	// Construct the full enhanced URL
	// TODO: replace with live version
	enhancedURL := fmt.Sprintf("http://localhost:8080/better/%s", enhancedKey)

	// Render the HTML response with the enhanced URL
	w.Header().Set("Content-Type", "text/html")
	responseHTML := fmt.Sprintf(`
		<h1>gorter URL Enhancer</h1>
	        <p>Original URL: %s</p>
        	<p>Better URL: <a href="%s">%s</a></p>
        	<form method="post" action="/enhanced">
            		<input type="text" name="url" placeholder="Enter a URL">
            		<input type="submit" value="Enhance">
        	</form>
	`, originalURL, enhancedURL, enhancedURL)
	fmt.Fprintf(w, responseHTML)
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
