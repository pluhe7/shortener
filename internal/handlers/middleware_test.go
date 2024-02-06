package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/pluhe7/shortener/internal/models"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/context"
)

func TestGzipCompressorMiddleware(t *testing.T) {
	srv := app.NewServer(&config.Config{
		Address: ":8080",
		BaseURL: "http://localhost:8080",
	})

	requestBody := `{"url":"https://yandex.ru"}`
	responseBodyRegexp := `{"result":"` + testConfig.BaseURL + `/([A-Za-z]{8})"}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)

		gzipWriter := gzip.NewWriter(buf)
		_, err := gzipWriter.Write([]byte(requestBody))
		require.NoError(t, err)

		err = gzipWriter.Close()
		require.NoError(t, err)

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", buf)
		request.Header.Set(echo.HeaderContentEncoding, "gzip")
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		responseRecorder := httptest.NewRecorder()
		c := srv.Echo.NewContext(request, responseRecorder)
		cc := &context.Context{
			Context: c,
			Server:  srv,
		}

		err = CompressorMiddleware(func(c echo.Context) error {
			return APIShortenHandler(cc)
		})(cc)
		require.NoError(t, err)

		result := responseRecorder.Result()
		require.Equal(t, http.StatusCreated, result.StatusCode)

		resultBody, err := io.ReadAll(result.Body)
		defer result.Body.Close()
		require.NoError(t, err)

		assert.Regexp(t, responseBodyRegexp, string(resultBody))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer([]byte(requestBody)))
		request.Header.Set(echo.HeaderAcceptEncoding, "gzip")
		request.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)

		responseRecorder := httptest.NewRecorder()
		c := srv.Echo.NewContext(request, responseRecorder)
		cc := &context.Context{
			Context: c,
			Server:  srv,
		}

		err := CompressorMiddleware(func(c echo.Context) error {
			return APIShortenHandler(cc)
		})(cc)
		require.NoError(t, err)

		result := responseRecorder.Result()
		require.Equal(t, http.StatusCreated, result.StatusCode)

		gzipReader, err := gzip.NewReader(result.Body)
		defer result.Body.Close()
		require.NoError(t, err)

		resultBody, err := io.ReadAll(gzipReader)
		defer gzipReader.Close()
		require.NoError(t, err)

		assert.Regexp(t, responseBodyRegexp, string(resultBody))
	})
}

func TestAuthMiddleware(t *testing.T) {
	const cookieName = "Token"

	srv := app.NewServer(&config.Config{
		Address: ":8080",
		BaseURL: "http://localhost:8080",
	})

	urls := []string{"https://yandex.ru", "https://google.com/search?q=test&q=something", "https://www.twitch.tv/guit88man"}

	t.Run("auth", func(t *testing.T) {
		// получим токен
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		responseRecorder := httptest.NewRecorder()

		c := srv.Echo.NewContext(request, responseRecorder)
		cc := &context.Context{
			Context: c,
			Server:  srv,
		}

		err := AuthMiddleware(func(c echo.Context) error {
			return PingDatabaseHandler(cc)
		})(cc)
		require.NoError(t, err)

		token := strings.TrimPrefix(responseRecorder.Header().Get("Set-Cookie"), "Token=")

		for _, url := range urls {
			reqBodyReader := strings.NewReader(url)

			request = httptest.NewRequest(http.MethodPost, "/", reqBodyReader)
			request.AddCookie(&http.Cookie{
				Name:  cookieName,
				Value: token,
			})

			responseRecorder = httptest.NewRecorder()

			c = srv.Echo.NewContext(request, responseRecorder)
			cc.Context = c

			err = AuthMiddleware(func(c echo.Context) error {
				return ShortenHandler(cc)
			})(cc)
			require.NoError(t, err)
		}

		request = httptest.NewRequest(http.MethodGet, "/user/urls", nil)
		request.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: token,
		})

		responseRecorder = httptest.NewRecorder()

		c = srv.Echo.NewContext(request, responseRecorder)
		cc.Context = c

		err = AuthMiddleware(func(c echo.Context) error {
			return GetUserURLs(cc)
		})(cc)
		require.NoError(t, err)

		result := responseRecorder.Result()
		defer result.Body.Close()
		require.Equal(t, http.StatusOK, result.StatusCode)

		var resp []models.ShortURLRecord
		err = json.NewDecoder(result.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, len(urls), len(resp)) // у первого запроса не будет
	})
}
