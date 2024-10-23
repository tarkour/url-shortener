package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
)

// GetRedirect returns the final URL after redirection
func GetRedirect(url string) (string, error) {
	const op = "api.GetRedirect"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse //stop after 1st redirect
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", nil
	}

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("#{op}: #{ErrInvalidStatusCode}: #{resp.StatusCode}")
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.Header.Get("Location"), nil
}
