package main

import (
    // "fmt"
    "math/rand"
    "net/http"
    "time"
)

// struct to contain urls, which is simply a map containing 
// shortened keys as keys and original URLs as values
type URLShortener struct {
    urls map[string]string
}

// Method of URLShortner struct to actually perform the shortening.
// args: http response writer, http request
func (us *URLShortener) HandleShorten(w http.ResponseWriter, r *http.Request) {
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

	// generate a unique shortened key for the original URL
	shortKey := generateShortKey()
	us.urls[shortKey] = originalURL
}

// function to create a new unique short key for the original URL
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
	shortKey := make([]byte, keyLength)
	// then for each element in the slice put in 
	// the character from charset at the index 
	// that is a pseudo-random number in the
	// interval [0, len(charset))
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	// and return the filled byte slice shortKey as (joined) a string
	return string(shortKey)
}
