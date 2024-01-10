package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/models"
)

var testConfig = config.Config{
	Address: ":8080",
	BaseURL: "http://localhost:8080",
}

func TestExpandHandler(t *testing.T) {
	type want struct {
		statusCode       int
		resp             string
		redirectLocation string
	}

	shortenReqBodyReader := strings.NewReader("https://yandex.ru")

	shortenRequest := httptest.NewRequest(http.MethodPost, "/", shortenReqBodyReader)
	shortenResponseRecorder := httptest.NewRecorder()

	srv := app.NewServer(&testConfig)
	srvHandler := SrvHandler{srv}

	c := srv.Echo.NewContext(shortenRequest, shortenResponseRecorder)

	err := srvHandler.ShortenHandler(c)
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
			id:   strings.TrimPrefix(string(preparedShortenURL), testConfig.BaseURL+"/"),
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

			c = srv.Echo.NewContext(expandRequest, expandResponseRecorder)
			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(test.id)

			srvHandler.ExpandHandler(c)

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
				respRegexp:  testConfig.BaseURL + "/([A-Za-z]{8})",
			},
		},
		{
			name: "url with params",
			url:  "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMETextPlain,
				respRegexp:  testConfig.BaseURL + "/([A-Za-z]{8})",
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

	srv := app.NewServer(&testConfig)
	srvHandler := SrvHandler{srv}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBodyReader := strings.NewReader(test.url)

			request := httptest.NewRequest(http.MethodPost, "/", reqBodyReader)
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)

			srvHandler.ShortenHandler(c)

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

func TestAPIShortenHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		withError   bool
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
				contentType: echo.MIMEApplicationJSON,
				withError:   false,
				respRegexp:  testConfig.BaseURL + "/([A-Za-z]{8})",
			},
		},
		{
			name: "url with params",
			url:  "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMEApplicationJSON,
				withError:   false,
				respRegexp:  testConfig.BaseURL + "/([A-Za-z]{8})",
			},
		},
		{
			name: "empty url",
			url:  "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				withError:   true,
				respRegexp:  "shorten url error: url shouldn't be empty",
			},
		},
	}

	srv := app.NewServer(&testConfig)
	srvHandler := SrvHandler{srv}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := models.ShortenRequest{
				URL: test.url,
			}

			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(reqJSON))
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)

			srvHandler.APIShortenHandler(c)

			result := responseRecorder.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get(echo.HeaderContentType), test.want.contentType)

			if !test.want.withError {
				var resp models.ShortenResponse
				err = json.NewDecoder(result.Body).Decode(&resp)
				require.NoError(t, err)

				assert.Regexp(t, test.want.respRegexp, resp.Result)

			} else {
				resultBody, err := io.ReadAll(result.Body)
				defer result.Body.Close()
				require.NoError(t, err)

				assert.Regexp(t, test.want.respRegexp, string(resultBody))
			}
		})
	}
}

func TestAPIShortenHandlerWithSaveToFile(t *testing.T) {
	cfg := testConfig
	cfg.FileStoragePath = "./test.json"

	defer os.Remove(cfg.FileStoragePath)

	urls := []string{"https://yandex.ru", "https://google.com/search?q=test&q=something"}

	srv := app.NewServer(&cfg)
	srvHandler := SrvHandler{srv}

	for _, url := range urls {
		req := models.ShortenRequest{
			URL: url,
		}

		reqJSON, err := json.Marshal(req)
		require.NoError(t, err)

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(reqJSON))
		responseRecorder := httptest.NewRecorder()

		c := srv.Echo.NewContext(request, responseRecorder)

		srvHandler.APIShortenHandler(c)
	}

	var records []models.ShortURLRecord

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	require.NoError(t, err)
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	for fileScanner.Scan() {
		var record models.ShortURLRecord

		row := fileScanner.Text()
		err = json.Unmarshal([]byte(row), &record)
		if err != nil {
			continue
		}

		records = append(records, record)
	}

	assert.Len(t, records, 2)
}
