package minion

import (
	"net/url"
	"testing"
)

func TestURL(t *testing.T) {
	uri := "ip/url/app?cache"

	v, err := url.Parse(uri)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	print(v.Path)
	print(v.Query().Has("cache"))

}
