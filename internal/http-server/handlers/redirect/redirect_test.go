package redirect_test

import (
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{{
		name:  "Success",
		alias: "test_alias",
		url:   "http://google.com/",
	},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			//check the final URL after redirection
			assert.Equal(t, tc.url, redirectToURL)
		})
	}
}
