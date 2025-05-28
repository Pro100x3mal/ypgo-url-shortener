package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"sync"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const shortURLLength = 8
const scheme = "http"

var errEmptyOriginalURL = errors.New("original URL is empty")
var errInvalidOriginalURL = errors.New("invalid URL")
var errEmptyShortURL = errors.New("short URL is empty")
var errNotExistShortURL = errors.New("short URL does not exist")

type URLStore struct {
	mu              sync.RWMutex
	originalToShort map[string]string
	shortToOriginal map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		originalToShort: make(map[string]string),
		shortToOriginal: make(map[string]string),
	}
}

var store = NewURLStore()

func (s *URLStore) Save(originalURL string) (string, error) {
	if len(originalURL) == 0 {
		return "", errEmptyOriginalURL
	}

	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", errInvalidOriginalURL
	}

	s.mu.RLock()
	shortURL, exists := s.originalToShort[originalURL]
	s.mu.RUnlock()

	if exists {
		return shortURL, nil
	}

	for {
		shortURL, err = genRandomString(shortURLLength)
		if err != nil {
			return "", err
		}
		s.mu.RLock()
		_, exists = s.shortToOriginal[shortURL]
		s.mu.RUnlock()

		if !exists {
			break
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.originalToShort[originalURL] = shortURL
	s.shortToOriginal[shortURL] = originalURL

	return shortURL, nil
}

func (s *URLStore) Get(shortURL string) (string, error) {
	if len(shortURL) == 0 {
		return "", errEmptyShortURL
	}
	s.mu.RLock()
	originalURL, exists := s.shortToOriginal[shortURL]
	s.mu.RUnlock()

	if exists {
		return originalURL, nil
	}

	return "", errNotExistShortURL
}

func saveShortURLHandler(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id, err := store.Save(string(originalURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   id,
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL.String()))
}

func getOrigURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	originalURL, err := store.Get(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func genRandomString(n int) (string, error) {
	randString := make([]byte, n)
	for i := range randString {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random index for short URL: %w", err)
		}
		randString[i] = charset[num.Int64()]
	}
	return string(randString), nil
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", saveShortURLHandler)
	mux.HandleFunc("GET /", getOrigURLHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
