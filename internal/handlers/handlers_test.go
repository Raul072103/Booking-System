package handlers

import (
	"context"
	"github.com/raul/BookingSystem/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{

	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"generals-quarters", "/generals-quarters", "GET", http.StatusOK},
	{"majors-suite", "/majors-suite", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"search-availability", "/search-availability", "GET", http.StatusOK},
	{"reservation-summary", "/reservation-summary", "GET", http.StatusOK},
	//{"search-availability", "/search-availability", "POST", []postData{
	//	{key: "start", value: "2020-01-01"},
	//	{key: "end", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"search-availability-json", "/search-availability-json", "POST", []postData{
	//	{key: "start", value: "2020-01-01"},
	//	{key: "end", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"make-reservation", "/make-reservation", "GET", []postData{
	//	{key: "fist_name", value: "John"},
	//	{key: "last_name", value: "Smith"},
	//	{key: "email", value: "m@gmail.com"},
	//	{key: "phone", value: "555-555-555"},
	//}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, e := range theTests {
		response, err := testServer.Client().Get(testServer.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if response.StatusCode != e.expectedStatusCode {
			t.Errorf("Expected %d, but got %d", e.expectedStatusCode, response.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			Id:       1,
			RoomName: "General's Quarters",
		},
	}

	req, err := http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	responseRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(responseRecorder, req)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returnet wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusOK)
	}

	// test case where reservation not in session
	req, err = http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	responseRecorder = httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, req)
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test case with none existing room
	req, err = http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	responseRecorder = httptest.NewRecorder()
	reservation.RoomId = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(responseRecorder, req)
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
		return nil
	}
	return ctx
}
