package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestToPropertiesDecodesFormValues(t *testing.T) {
	body := "stagename=team-blue%3A0-0&team=sky-blue&note=a%2Bb+space&duplicate=first&duplicate=last"
	request := httptest.NewRequest("POST", "/insert", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	properties, ok := requestToProperties(request)
	if !ok {
		t.Fatal("requestToProperties() rejected a valid form body")
	}

	want := map[string]string{
		"stagename": "team-blue:0-0",
		"team":      "sky-blue",
		"note":      "a+b space",
		"duplicate": "last",
	}
	for key, wantValue := range want {
		if properties[key] != wantValue {
			t.Errorf("requestToProperties()[%q] = %q, want %q", key, properties[key], wantValue)
		}
	}
}
