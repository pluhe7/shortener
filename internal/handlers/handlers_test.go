package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExpandHandler(t *testing.T) {
	type want struct {
		statusCode       int
		resp             string
		redirectLocation string
	}

	shortenReqBodyReader := strings.NewReader("https://yandex.ru")

	shortenRequest := httptest.NewRequest(http.MethodPost, "/", shortenReqBodyReader)
	shortenResponseRecorder := httptest.NewRecorder()

	ShortenHandler(shortenResponseRecorder, shortenRequest)

	shortenResult := shortenResponseRecorder.Result()

	preparedShortenURL, err := io.ReadAll(shortenResult.Body)
	defer shortenResult.Body.Close()
	require.NoError(t, err)

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "base test",
			url:  string(preparedShortenURL),
			want: want{
				statusCode:       http.StatusTemporaryRedirect,
				redirectLocation: "https://yandex.ru",
				resp:             "<a href=\"https://yandex.ru\">Temporary Redirect</a>.\n\n",
			},
		},
		{
			name: "no id in url",
			url:  "http://localhost:8080",
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "invalid url id",
			},
		},
		{
			name: "wrong id len",
			url:  "http://localhost:8080/tooLongId",
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "invalid url id",
			},
		},
		{
			name: "not existing id",
			url:  "http://localhost:8080/notEx9",
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "invalid url id",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expandRequest := httptest.NewRequest(http.MethodGet, test.url, nil)
			expandResponseRecorder := httptest.NewRecorder()

			ExpandHandler(expandResponseRecorder, expandRequest)

			expandResult := expandResponseRecorder.Result()
			expandResultBody, err := io.ReadAll(expandResult.Body)
			defer expandResult.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.redirectLocation, expandResult.Header.Get("Location"))
			require.Equal(t, test.want.resp, string(expandResultBody))
		})
	}
}

func TestShortenHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		respRegexp  string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "simple url",
			url:  "https://yandex.ru",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				respRegexp:  "http://localhost:8080/([A-Za-z]{8})",
			},
		},
		{
			name: "url with params",
			url:  "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				respRegexp:  "http://localhost:8080/([A-Za-z]{8})",
			},
		},
		{
			name: "empty url",
			url:  "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				respRegexp:  "shorten url error: url shouldn't be empty",
			},
		},
		{
			name: "not url",
			url:  "something that is not url",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				respRegexp:  "shorten url error: url validation error: (.*)",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBodyReader := strings.NewReader(test.url)

			request := httptest.NewRequest(http.MethodPost, "/", reqBodyReader)
			responseRecorder := httptest.NewRecorder()

			ShortenHandler(responseRecorder, request)

			result := responseRecorder.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Contains(t, test.want.contentType, result.Header.Get("Content-Type"))

			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			assert.Regexp(t, test.want.respRegexp, string(resultBody))
		})
	}
}
