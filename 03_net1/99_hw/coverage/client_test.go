package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestBadToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: "some_invalid_access_token",
		URL:         ts.URL,
	}
	req := SearchRequest{}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestInvalidUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         "some_server_url",
	}
	req := SearchRequest{}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestLimitAndOffsetOnLessThanZero(t *testing.T) {
	requests := []SearchRequest{
		{
			Limit: -1,
		}, {
			Offset: -5,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	for _, req := range requests {
		_, err := sc.FindUsers(req)
		if err == nil {
			t.Errorf("%s", err)
		}
	}
}

func TestMaxLimit(t *testing.T) {
	requests := []SearchRequest{
		{
			Limit: 100500,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	for _, req := range requests {
		_, err := sc.FindUsers(req)
		if err != nil {
			t.Errorf("%s", err)
		}
	}
}

func TestRequestsEqualLimit(t *testing.T) {
	u := Users{}
	u.parseUsersXML("dataset.xml") //nolint:errcheck
	users := make([]User, 0)
	for i := 0; i < len(u.List); i++ {
		user := User{
			ID:     u.List[i].ID,
			Name:   u.List[i].FirstName + " " + u.List[i].LastName,
			Age:    u.List[i].Age,
			About:  u.List[i].About,
			Gender: u.List[i].Gender,
		}
		users = append(users, user)
	}
	req := SearchRequest{
		Limit: 5,
	}
	expected := SearchResponse{
		Users:    users[0:req.Limit],
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	resp, err := sc.FindUsers(req)
	if err != nil {
		t.Errorf("%s", err)
	}
	if !reflect.DeepEqual(*resp, expected) {
		t.Errorf("invalid result!")
	}
}

func TestBadResponse(t *testing.T) {
	requests := []SearchRequest{
		{
			OrderField: "LALALALA",
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	for _, req := range requests {
		_, err := sc.FindUsers(req)
		if err == nil {
			t.Errorf("%s", err)
		}
	}
}

func TestOrderBy(t *testing.T) {
	requests := []SearchRequest{
		{
			OrderField: "Id",
			OrderBy:    1,
		}, {
			OrderField: "Id",
			OrderBy:    -1,
		}, {
			OrderField: "Age",
			OrderBy:    1,
		}, {
			OrderField: "Age",
			OrderBy:    -1,
		}, {
			OrderBy: 1,
		}, {
			OrderBy: -1,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	for _, req := range requests {
		_, err := sc.FindUsers(req)
		if err != nil {
			t.Errorf("%s", err)
		}
	}
}

func TestCrash(t *testing.T) {
	req := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestTimeout(t *testing.T) {
	req := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 3)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestBadRequestUnpackJSON(t *testing.T) {
	req := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestBadRequestUnknownError(t *testing.T) {
	req := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"msg": "lol"}`)) //nolint:errcheck
	}))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}

func TestOkUnpackJSON(t *testing.T) {
	req := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"msg": "lol"}`)) //nolint:errcheck
	}))
	defer ts.Close()
	sc := &SearchClient{
		AccessToken: CorrectAccessToken,
		URL:         ts.URL,
	}
	_, err := sc.FindUsers(req)
	if err == nil {
		t.Errorf("%s", err)
	}
}
