package main

import (
    // "fmt"
    // "math/rand"
    "net/http"
    // "time"
)

// struct to contain urls, which is simply a map containing 
// shortened keys as keys and original URLs as values
type URLShortener struct {
    urls map[string]string
}

// function to actually perform the shortening
// TODO: further comments for LEARNING
func (us *URLShortener) HandleShorten(w http.ResponseWriter, r *http.Request) {
	// TODO
}
