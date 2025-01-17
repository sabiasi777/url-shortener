package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
)

var (
	store = make(map[string]string) // In-memory storage
	mu    sync.Mutex                // To handle concurrent writes
)

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data:"+err.Error(), http.StatusBadRequest)
		return
	}

	originalURL := r.FormValue("URL")
	if originalURL == "" {
		http.Error(w, "URL field is missing", http.StatusBadRequest)
		return
	}

	fmt.Println("originalURL:", originalURL)

	hash := sha256.Sum256([]byte(originalURL))
	shortKey := base64.URLEncoding.EncodeToString(hash[:])[:8]

	// baseURL := "https://short-url.local"
	// shortenedURL := fmt.Sprintf("%s/%s", baseURL, shortKey)

	mu.Lock()
	store[shortKey] = originalURL
	mu.Unlock()

	fmt.Fprintf(w, "\nShort URL: http://shorturl:8080/%s\n", shortKey)
	// fmt.Fprintf(w, `{"shortened_url": "%s"}`, shortenedURL)

}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	shortKey := r.URL.Path[1:]

	mu.Lock()
	originalURL, exists := store[shortKey]
	mu.Unlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
