package utils

import (
	"log"
	"net/http"
	"time"
)

func HTTPWithRetry(fn func(string) (*http.Response, error), url string) (*http.Response, error) {
	maxAttempts := 10
	var timeOut time.Duration = 5
	var res *http.Response
	var err error

	for i := 0; i < maxAttempts; i++ {
		res, err = fn(url)
		if err != nil {
			log.Printf("request to %s failed: %s; retrying in %d seconds\n", url, err, timeOut)
		} else {
			break
		}
	}
	return res, err
}
