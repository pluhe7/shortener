package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/models"
	"github.com/pluhe7/shortener/internal/storage"
	"github.com/pluhe7/shortener/internal/storage/mocks"
)

var testConfig = config.Config{
	Address: ":8080",
	BaseURL: "http://localhost:8080",
}

const idLen = 8

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
				respRegexp:  fmt.Sprintf("%s/([A-Za-z]{%d})", testConfig.BaseURL, idLen),
			},
		},
		{
			name: "url with params",
			url:  "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMETextPlain,
				respRegexp:  fmt.Sprintf("%s/([A-Za-z]{%d})", testConfig.BaseURL, idLen),
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

func TestAPIShortenHandlerMemoryStorage(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		resp        string
	}

	tests := []struct {
		name      string
		withError bool
		url       string
		want      want
	}{
		{
			name:      "simple url",
			withError: false,
			url:       "https://yandex.ru",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMEApplicationJSON,
				resp:        fmt.Sprintf("%s/([A-Za-z]{%d})", testConfig.BaseURL, idLen),
			},
		},
		{
			name:      "url with params",
			withError: false,
			url:       "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMEApplicationJSON,
				resp:        fmt.Sprintf("%s/([A-Za-z]{%d})", testConfig.BaseURL, idLen),
			},
		},
		{
			name:      "empty url",
			withError: true,
			url:       "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				resp:        "shorten url error: url shouldn't be empty",
			},
		},
	}

	srv := app.NewServer(&testConfig)

	memStorage, err := storage.NewMemoryStorage()
	require.NoError(t, err)
	srv.Storage = memStorage

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

			if !test.withError {
				var resp models.ShortenResponse
				err = json.NewDecoder(result.Body).Decode(&resp)
				require.NoError(t, err)

				assert.Regexp(t, test.want.resp, resp.Result)

			} else {
				resultBody, err := io.ReadAll(result.Body)
				defer result.Body.Close()
				require.NoError(t, err)

				assert.Equal(t, test.want.resp, string(resultBody))
			}
		})
	}
}

func TestAPIShortenHandlerFileStorage(t *testing.T) {
	cfg := testConfig
	cfg.FileStoragePath = "./test.json"

	defer os.Remove(cfg.FileStoragePath)

	type want struct {
		statusCode  int
		contentType string
		resp        string
	}

	tests := []struct {
		name      string
		withError bool
		url       string
		want      want
	}{
		{
			name:      "simple url",
			withError: false,
			url:       "https://yandex.ru",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMEApplicationJSON,
				resp:        fmt.Sprintf("%s/([A-Za-z]{%d})", testConfig.BaseURL, idLen),
			},
		},
		{
			name:      "url with params",
			withError: false,
			url:       "https://google.com/search?q=test&q=something",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMEApplicationJSON,
				resp:        fmt.Sprintf("%s/([A-Za-z]{%d})", testConfig.BaseURL, idLen),
			},
		},
		{
			name:      "empty url",
			withError: true,
			url:       "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				resp:        "shorten url error: url shouldn't be empty",
			},
		},
	}

	srv := app.NewServer(&cfg)
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

			if !test.withError {
				var resp models.ShortenResponse
				err = json.NewDecoder(result.Body).Decode(&resp)
				require.NoError(t, err)

				assert.Regexp(t, test.want.resp, resp.Result)

			} else {
				resultBody, err := io.ReadAll(result.Body)
				defer result.Body.Close()
				require.NoError(t, err)

				assert.Equal(t, test.want.resp, string(resultBody))
			}
		})
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

func TestPingDBHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name      string
		url       string
		withError bool
		want      want
	}{
		{
			name:      "success ping",
			withError: false,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "fail ping",
			withError: true,
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockStorage := mocks.NewMockStorage(mockController)

	srv := app.NewServer(&testConfig)
	srv.Storage = mockStorage
	srvHandler := SrvHandler{srv}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.withError {
				mockStorage.EXPECT().PingContext(gomock.Any()).Return(fmt.Errorf("some error"))
			} else {
				mockStorage.EXPECT().PingContext(gomock.Any()).Return(nil)
			}

			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)

			srvHandler.PingDatabaseHandler(c)

			result := responseRecorder.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
		})
	}
}

func TestAPIBatchShortenHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		resp        string
	}

	tests := []struct {
		name      string
		withError bool
		req       string
		want      want
	}{
		{
			name:      "success",
			withError: false,
			req: `[
				{
					"correlation_id": "yandex",
					"original_url": "https://yandex.ru"
				},
				{
					"correlation_id": "google",
					"original_url": "https://google.com/search?q=test&q=something"
				}
			]`,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: echo.MIMEApplicationJSON,
				resp: fmt.Sprintf(`[
					{
						"correlation_id": "yandex",
						"short_url": "%s/([A-Za-z]{%d})"
					},
					{
						"correlation_id": "google",
						"short_url": "%s/([A-Za-z]{%d})"
					}
				]`, testConfig.BaseURL, idLen, testConfig.BaseURL, idLen),
			},
		},
		{
			name:      "wrong request",
			withError: true,
			req:       "something",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: echo.MIMETextPlain,
				resp:        "decode request error",
			},
		},
		{
			name:      "save error",
			withError: true,
			req: `[
				{
					"correlation_id": "yandex",
					"original_url": "https://yandex.ru"
				},
				{
					"correlation_id": "google",
					"original_url": "https://google.com/search?q=test&q=something"
				}
			]`,
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: echo.MIMETextPlain,
				resp:        "some error",
			},
		},
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockStorage := mocks.NewMockStorage(mockController)

	srv := app.NewServer(&testConfig)
	srv.Storage = mockStorage
	srvHandler := SrvHandler{srv}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.withError {
				mockStorage.EXPECT().SaveBatch(gomock.Any()).Return(fmt.Errorf("some error")).AnyTimes()
			} else {
				mockStorage.EXPECT().SaveBatch(gomock.Any()).Return(nil)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte(test.req)))
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)

			srvHandler.APIBatchShortenHandler(c)

			result := responseRecorder.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get(echo.HeaderContentType), test.want.contentType)

			if !test.withError {
				var resp []models.ShortURLWithID
				err := json.NewDecoder(result.Body).Decode(&resp)
				require.NoError(t, err)

				var wantResp []models.ShortURLWithID
				json.Unmarshal([]byte(test.want.resp), &wantResp)

				require.Equal(t, len(wantResp), len(resp))

				sort.Slice(resp, func(i, j int) bool {
					return resp[i].CorrelationID < resp[j].CorrelationID
				})
				sort.Slice(wantResp, func(i, j int) bool {
					return wantResp[i].CorrelationID < wantResp[j].CorrelationID
				})

				for i := range wantResp {
					assert.Equal(t, wantResp[i].CorrelationID, resp[i].CorrelationID)
					assert.Regexp(t, wantResp[i].ShortURL, resp[i].ShortURL)
				}

			} else {
				resultBody, err := io.ReadAll(result.Body)
				defer result.Body.Close()
				require.NoError(t, err)

				assert.Contains(t, string(resultBody), test.want.resp)
			}
		})
	}
}
