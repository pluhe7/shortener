package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/pluhe7/shortener/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var serverAddr = "http://localhost:8080/"

func TestExpandHandler(t *testing.T) {
	type want struct {
		statusCode       int
		resp             string
		redirectLocation string
	}

	shortenReqBodyReader := strings.NewReader("https://yandex.ru")

	shortenRequest := httptest.NewRequest(http.MethodPost, "/", shortenReqBodyReader)
	shortenResponseRecorder := httptest.NewRecorder()

	config.InitConfig()
	e := echo.New()
	c := e.NewContext(shortenRequest, shortenResponseRecorder)

	err := ShortenHandler(c)
	require.NoError(t, err)

	shortenResult := shortenResponseRecorder.Result()

	preparedShortenURL, err := io.ReadAll(shortenResult.Body)
	defer shortenResult.Body.Close()
	require.NoError(t, err)

	tests := []struct {
		name string
		id   string
		want want
	}{
		{
			name: "base test",
			id:   strings.TrimPrefix(string(preparedShortenURL), serverAddr),
			want: want{
				statusCode:       http.StatusTemporaryRedirect,
				redirectLocation: "https://yandex.ru",
				resp:             "",
			},
		},
		{
			name: "no id in url",
			id:   "",
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "expand url error: invalid url id",
			},
		},
		{
			name: "wrong id len",
			id:   "tooLongId",
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "expand url error: invalid url id",
			},
		},
		{
			name: "not existing id",
			id:   "notEx9",
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "expand url error: invalid url id",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expandRequest := httptest.NewRequest(http.MethodGet, "/"+test.id, nil)
			expandResponseRecorder := httptest.NewRecorder()

			c = e.NewContext(expandRequest, expandResponseRecorder)
			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(test.id)

			ExpandHandler(c)

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
				contentType: echo.MIMETextPlain,
				respRegexp:  serverAddr + "([A-Za-z]{8})",
			},
		},
		{
			name: "url with params",
			url:  "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMETextPlain,
				respRegexp:  serverAddr + "([A-Za-z]{8})",
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
	}

	config.InitConfig()
	e := echo.New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBodyReader := strings.NewReader(test.url)

			request := httptest.NewRequest(http.MethodPost, "/", reqBodyReader)
			responseRecorder := httptest.NewRecorder()

			c := e.NewContext(request, responseRecorder)

			ShortenHandler(c)

			result := responseRecorder.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get(echo.HeaderContentType), test.want.contentType)

			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			assert.Regexp(t, test.want.respRegexp, string(resultBody))
		})
	}
}
