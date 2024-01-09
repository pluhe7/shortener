package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/handlers"
)

var testConfig = config.Config{
	Address: ":8080",
	BaseURL: "http://localhost:8080",
}

func TestGzipCompressorMiddleware(t *testing.T) {
	config.SetConfig(testConfig)

	requestBody := `{"url":"https://yandex.ru"}`
	responseBodyRegexp := `{"result":"` + testConfig.BaseURL + `/([A-Za-z]{8})"}`

	e := echo.New()

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
		c := e.NewContext(request, responseRecorder)

		err = CompressorMiddleware(func(c echo.Context) error {
			return handlers.APIShortenHandler(c)
		})(c)
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
		c := e.NewContext(request, responseRecorder)

		err := CompressorMiddleware(func(c echo.Context) error {
			return handlers.APIShortenHandler(c)
		})(c)
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
