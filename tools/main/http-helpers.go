package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func requestToProperties(r *http.Request) (map[string]string, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, false
	}

	bodyS := string(body[:])

	// Not great, '&' included in body will be unescaped here producing weird results
	bodyS, err = url.QueryUnescape(bodyS)
	if err != nil {
		return nil, false
	}
	//fmt.Println(bodyS)

	return bodyStringToProperties(bodyS), true
}

func bodyStringToProperties(body string) map[string]string {
	propMap := make(map[string]string)
	props := strings.Split(body, "&")
	for _, prop := range props {
		keyValue := strings.Split(prop, "=")
		if len(keyValue) > 1 {
			propMap[keyValue[0]] = keyValue[1] // 1: ?
		}
	}
	return propMap
}
