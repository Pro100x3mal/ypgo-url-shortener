package main

import (
	"crypto/rand"
	"errors"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const shortURLLength = 8
const scheme = "http"

var errEmptyOriginalURL = errors.New("original url is empty")
var errEmptyShortURL = errors.New("short url is empty")
var errNotExistShortURL = errors.New("this short url is not exist")

var urlStore = make(map[string]string)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postHandler(w, r)
	case http.MethodGet:
		getHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id, err := saveShortURL(string(originalURL))
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

func getHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := r.URL.Path[1:]
	originalURL, err := getOriginalURL(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func saveShortURL(originalURL string) (string, error) {
	if len(originalURL) == 0 {
		return "", errEmptyOriginalURL
	}
	shortURL, exist := urlStore[originalURL]
	if !exist {
		shortURL = genRandomString(shortURLLength)
		urlStore[originalURL] = shortURL
		return shortURL, nil
	}
	return shortURL, nil
}

func getOriginalURL(id string) (string, error) {
	if len(id) == 0 {
		return "", errEmptyShortURL
	}
	for originalURL, shortURL := range urlStore {
		if shortURL == id {
			return originalURL, nil
		}
	}
	return "", errNotExistShortURL
}

func genRandomString(n int) string {
	randString := make([]byte, n)
	for i := range randString {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			log.Fatal(err)
		}
		randString[i] = charset[num.Int64()]
	}
	return string(randString)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}

}
