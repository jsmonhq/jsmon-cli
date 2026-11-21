package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func respond(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func clientWith(rt roundTripFunc) *Client {
	c := NewClient("key", nil)
	c.HTTPClient = &http.Client{Transport: rt}
	return c
}

func TestGetWorkspacesUsesOverview(t *testing.T) {
	var paths []string
	c := clientWith(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		if r.Header.Get("X-Jsmon-Key") != "key" {
			t.Errorf("missing API key header")
		}
		return respond(200, `{"message":"ok","total":1,"workspaces":[{"wkspId":"w1","name":"A","isShared":true}]}`), nil
	})
	res, err := c.GetWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != "/api/v2/getWorkspacesOverview" {
		t.Fatalf("paths = %v, want only /api/v2/getWorkspacesOverview", paths)
	}
	if len(res.Workspaces) != 1 || res.Workspaces[0].WkspID != "w1" || !res.Workspaces[0].IsShared {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestGetWorkspacesReturnsAPIErrorOnAuthFailure(t *testing.T) {
	calls := 0
	c := clientWith(func(r *http.Request) (*http.Response, error) {
		calls++
		return respond(401, `{"message":"invalid key"}`), nil
	})
	_, err := c.GetWorkspaces()
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Status != 401 {
		t.Fatalf("err = %v, want APIError 401", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (a single request, no retry)", calls)
	}
}
