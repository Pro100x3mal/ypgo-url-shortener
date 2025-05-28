package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_saveShortURLHandler(t *testing.T) {
	originalURL := "https://example.com"
	type want struct {
		method      string
		contentType string
		statusCode  int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test save short URL",
			want: want{
				method:      http.MethodPost,
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.want.method, "/", strings.NewReader(originalURL))
			w := httptest.NewRecorder()
			saveShortURLHandler(w, r)

			assert.Equal(t, tt.want.statusCode, w.Code, "status code should be 201")
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"), "content type should be text/plain")
			assert.NotEmpty(t, w.Body.String(), "response body should not be empty")
			_, err := url.ParseRequestURI(w.Body.String())
			assert.NoError(t, err, "response body should be a valid URL")

		})
	}
}

func Test_getOrigURLHandler(t *testing.T) {
	originalURL := "https://example.com"
	type want struct {
		saveMethod     string
		getMethod      string
		saveStatusCode int
		getStatusCode  int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test get original URL",
			want: want{
				saveMethod:     http.MethodPost,
				getMethod:      http.MethodGet,
				saveStatusCode: http.StatusCreated,
				getStatusCode:  http.StatusTemporaryRedirect,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveReq := httptest.NewRequest(tt.want.saveMethod, "/", strings.NewReader(originalURL))
			saveW := httptest.NewRecorder()
			saveShortURLHandler(saveW, saveReq)

			assert.Equal(t, tt.want.saveStatusCode, saveW.Code, "status code should be 201")
			shortURL := saveW.Body.String()
			parsed, err := url.Parse(shortURL)
			assert.NoError(t, err, "short URL should be a valid URL")

			getReq := httptest.NewRequest(tt.want.getMethod, parsed.Path, nil)
			getW := httptest.NewRecorder()
			getOrigURLHandler(getW, getReq)

			assert.Equal(t, tt.want.getStatusCode, getW.Code, "status code should be 307")
			assert.Equal(t, originalURL, getW.Header().Get("Location"), "location header should be the original URL")
		})
	}
}
