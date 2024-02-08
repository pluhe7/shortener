package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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
	"github.com/pluhe7/shortener/internal/context"
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
	c := srv.Echo.NewContext(shortenRequest, shortenResponseRecorder)

	cc := &context.Context{
		Context: c,
		Server:  srv,
	}

	err := ShortenHandler(cc)
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

			cc.Context = c

			ExpandHandler(cc)

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBodyReader := strings.NewReader(test.url)

			request := httptest.NewRequest(http.MethodPost, "/", reqBodyReader)
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			ShortenHandler(cc)

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
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			APIShortenHandler(cc)

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
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			APIShortenHandler(cc)

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.withError {
				mockStorage.EXPECT().PingContext(gomock.Any()).Return(errors.New("some error"))
			} else {
				mockStorage.EXPECT().PingContext(gomock.Any()).Return(nil)
			}

			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			PingDatabaseHandler(cc)

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.withError {
				mockStorage.EXPECT().SaveBatch(gomock.Any()).Return(errors.New("some error")).AnyTimes()
			} else {
				mockStorage.EXPECT().SaveBatch(gomock.Any()).Return(nil)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte(test.req)))
			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			APIBatchShortenHandler(cc)

			result := responseRecorder.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get(echo.HeaderContentType), test.want.contentType)

			if !test.withError {
				var resp []models.ShortURLWithID
				err := json.NewDecoder(result.Body).Decode(&resp)
				require.NoError(t, err)

				var wantResp []models.ShortURLWithID
				err = json.Unmarshal([]byte(test.want.resp), &wantResp)
				require.NoError(t, err)

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

func TestGetUserURLsHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		resp        string
	}

	tests := []struct {
		name      string
		userID    string
		withError bool
		records   []models.ShortURLRecord
		want      want
	}{
		{
			name:      "success",
			userID:    "someUser",
			withError: false,
			records: []models.ShortURLRecord{
				{
					ShortURL:    testConfig.BaseURL + "/asdzxc",
					OriginalURL: "https://yandex.ru",
					UserID:      "someUser",
				},
				{
					ShortURL:    testConfig.BaseURL + "/qwerty",
					OriginalURL: "https://google.com/search?q=test&q=something",
					UserID:      "someUser",
				},
				// эта запись не включена в ответ, т.к. она принадлежит другому юзеру
				{
					ShortURL:    testConfig.BaseURL + "/goodoq",
					OriginalURL: "https://www.twitch.tv/goodoq",
					UserID:      "otherUser",
				},
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: echo.MIMEApplicationJSON,
				resp: fmt.Sprintf(`[
					{
						"original_url": "https://yandex.ru",
						"short_url": "%s/asdzxc",
						"user_id": "someUser"
					},
					{
						"original_url": "https://google.com/search?q=test&q=something",
						"short_url": "%s/qwerty",
						"user_id": "someUser"
					}
				]`, testConfig.BaseURL, testConfig.BaseURL),
			},
		},
		{
			name:      "empty user id",
			userID:    "",
			withError: true,
			want: want{
				statusCode:  http.StatusUnauthorized,
				contentType: "",
				resp:        "",
			},
		},
		{
			name:      "empty answer",
			userID:    "someUser2",
			withError: true,
			records:   []models.ShortURLRecord{},
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "",
				resp:        "",
			},
		},
	}

	srv := app.NewServer(&testConfig)

	memStorage, err := storage.NewMemoryStorage()
	require.NoError(t, err)

	srv.Storage = memStorage

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
			request.AddCookie(&http.Cookie{
				Name:  "Token",
				Value: "some_cookie",
			})

			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			cc.Server.SessionUserID = test.userID

			for _, record := range test.records {
				err = cc.Server.Storage.Save(record)
				require.NoError(t, err)
			}

			GetUserURLs(cc)

			result := responseRecorder.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get(echo.HeaderContentType), test.want.contentType)

			if !test.withError {
				compareRecordsResponses(test.want.resp, result.Body, t)

			} else {
				resultBody, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				defer result.Body.Close()

				assert.Contains(t, string(resultBody), test.want.resp)
			}
		})
	}
}

func TestDeleteUserURLsHandler(t *testing.T) {
	type want struct {
		statusCode     int
		getRecordsResp string
	}

	tests := []struct {
		name             string
		deleteUserID     string
		getRecordsUserID string
		req              string
		want             want
	}{
		{
			name:             "delete own records",
			deleteUserID:     "someUser",
			getRecordsUserID: "someUser",
			req:              "[\"asdzxc\",\"qwerty\"]",
			want: want{
				statusCode: http.StatusAccepted,
				getRecordsResp: fmt.Sprintf(`[
					{
						"original_url": "https://yandex.ru",
						"short_url": "%s/asdzxc",
						"user_id": "someUser",
						"is_deleted": "true"
					},
					{
						"original_url": "https://google.com/search?q=test&q=something",
						"short_url": "%s/qwerty",
						"user_id": "someUser",
						"is_deleted": "true"
					},
					{
						"original_url": "https://www.twitch.tv/goodoq",
						"short_url": "%s/goodoq",
						"user_id": "someUser",
						"is_deleted": "false"
					}
				]`, testConfig.BaseURL, testConfig.BaseURL, testConfig.BaseURL),
			},
		},
		{
			name:             "delete not existing record",
			deleteUserID:     "someUser",
			getRecordsUserID: "someUser",
			req:              "[\"gggggg\"]",
			want: want{
				statusCode: http.StatusAccepted,
				getRecordsResp: fmt.Sprintf(`[
					{
						"original_url": "https://yandex.ru",
						"short_url": "%s/asdzxc",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://google.com/search?q=test&q=something",
						"short_url": "%s/qwerty",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://www.twitch.tv/goodoq",
						"short_url": "%s/goodoq",
						"user_id": "someUser",
						"is_deleted": "false"
					}
				]`, testConfig.BaseURL, testConfig.BaseURL, testConfig.BaseURL),
			},
		},
		{
			name:             "delete not own record",
			deleteUserID:     "someUser2",
			getRecordsUserID: "someUser",
			req:              "[\"asdzxc\"]",
			want: want{
				statusCode: http.StatusAccepted,
				getRecordsResp: fmt.Sprintf(`[
					{
						"original_url": "https://yandex.ru",
						"short_url": "%s/asdzxc",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://google.com/search?q=test&q=something",
						"short_url": "%s/qwerty",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://www.twitch.tv/goodoq",
						"short_url": "%s/goodoq",
						"user_id": "someUser",
						"is_deleted": "false"
					}
				]`, testConfig.BaseURL, testConfig.BaseURL, testConfig.BaseURL),
			},
		},
		{
			name:             "empty user id",
			deleteUserID:     "",
			getRecordsUserID: "someUser",
			req:              "[\"asdzxc\"]",
			want: want{
				statusCode: http.StatusUnauthorized,
				getRecordsResp: fmt.Sprintf(`[
					{
						"original_url": "https://yandex.ru",
						"short_url": "%s/asdzxc",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://google.com/search?q=test&q=something",
						"short_url": "%s/qwerty",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://www.twitch.tv/goodoq",
						"short_url": "%s/goodoq",
						"user_id": "someUser",
						"is_deleted": "false"
					}
				]`, testConfig.BaseURL, testConfig.BaseURL, testConfig.BaseURL),
			},
		},
		{
			name:             "bad request",
			deleteUserID:     "someUser",
			getRecordsUserID: "someUser",
			req:              "123wrong_body",
			want: want{
				statusCode: http.StatusBadRequest,
				getRecordsResp: fmt.Sprintf(`[
					{
						"original_url": "https://yandex.ru",
						"short_url": "%s/asdzxc",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://google.com/search?q=test&q=something",
						"short_url": "%s/qwerty",
						"user_id": "someUser",
						"is_deleted": "false"
					},
					{
						"original_url": "https://www.twitch.tv/goodoq",
						"short_url": "%s/goodoq",
						"user_id": "someUser",
						"is_deleted": "false"
					}
				]`, testConfig.BaseURL, testConfig.BaseURL, testConfig.BaseURL),
			},
		},
	}

	srv := app.NewServer(&testConfig)

	memStorage, err := storage.NewMemoryStorage()
	require.NoError(t, err)

	srv.Storage = memStorage

	records := []models.ShortURLRecord{
		{
			ShortURL:    testConfig.BaseURL + "/asdzxc",
			OriginalURL: "https://yandex.ru",
			UserID:      "someUser",
		},
		{
			ShortURL:    testConfig.BaseURL + "/qwerty",
			OriginalURL: "https://google.com/search?q=test&q=something",
			UserID:      "someUser",
		},
		{
			ShortURL:    testConfig.BaseURL + "/goodoq",
			OriginalURL: "https://www.twitch.tv/goodoq",
			UserID:      "someUser",
		},
		{
			ShortURL:    testConfig.BaseURL + "/cakels",
			OriginalURL: "https://www.twitch.tv/c_a_k_e",
			UserID:      "otherUser",
		},
	}
	for _, record := range records {
		err = srv.Storage.Save(record)
		require.NoError(t, err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/user/urls", strings.NewReader(test.req))
			request.AddCookie(&http.Cookie{
				Name:  "Token",
				Value: "some_cookie",
			})

			responseRecorder := httptest.NewRecorder()

			c := srv.Echo.NewContext(request, responseRecorder)
			cc := &context.Context{
				Context: c,
				Server:  srv,
			}

			cc.Server.SessionUserID = test.deleteUserID

			DeleteUserURLs(cc)

			result := responseRecorder.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.statusCode, result.StatusCode)

			if test.want.getRecordsResp != "" {
				request = httptest.NewRequest(http.MethodDelete, "/user/urls", strings.NewReader(test.req))
				request.AddCookie(&http.Cookie{
					Name:  "Token",
					Value: "some_cookie",
				})

				responseRecorder = httptest.NewRecorder()

				c = srv.Echo.NewContext(request, responseRecorder)
				cc = &context.Context{
					Context: c,
					Server:  srv,
				}

				cc.Server.SessionUserID = test.getRecordsUserID

				GetUserURLs(cc)

				result = responseRecorder.Result()
				defer result.Body.Close()

				compareRecordsResponses(test.want.getRecordsResp, result.Body, t)

			}
		})
	}
}

func compareRecordsResponses(wantRespStr string, actualRespBody io.ReadCloser, t *testing.T) {
	var resp []models.ShortURLRecord
	err := json.NewDecoder(actualRespBody).Decode(&resp)
	require.NoError(t, err)

	var wantResp []models.ShortURLRecord
	err = json.Unmarshal([]byte(wantRespStr), &wantResp)
	require.NoError(t, err)

	require.Equal(t, len(wantResp), len(resp))

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].OriginalURL < resp[j].OriginalURL
	})
	sort.Slice(wantResp, func(i, j int) bool {
		return wantResp[i].OriginalURL < wantResp[j].OriginalURL
	})

	for i := range wantResp {
		assert.Equal(t, wantResp[i].OriginalURL, resp[i].OriginalURL)
		assert.Regexp(t, wantResp[i].ShortURL, resp[i].ShortURL)
	}
}
