package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/raul/BookingSystem/internal/config"
	"github.com/raul/BookingSystem/internal/driver"
	"github.com/raul/BookingSystem/internal/forms"
	"github.com/raul/BookingSystem/internal/helpers"
	"github.com/raul/BookingSystem/internal/models"
	"github.com/raul/BookingSystem/internal/render"
	"github.com/raul/BookingSystem/internal/repository"
	"github.com/raul/BookingSystem/internal/repository/dbrepo"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home renders the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "home.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// About renders the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "about.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomById(res.RoomId)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't find room!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	startDate := res.StartDate.Format("2006-01-02")
	endDate := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	data := make(map[string]any)
	data["reservation"] = res

	err = render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		err := render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
			Form: form,
			Data: data,
		})
		if err != nil {
			helpers.ServerError(w, err)
		}
		return
	}

	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomId:        reservation.RoomId,
		ReservationId: newReservationId,
		RestrictionId: 2,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send notifications - first to guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s:, <br>
		This is to confirm your reservation from %s to %s.
`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "me@here.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	// send notifications - second to owner
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong><br>
		A reservation has been made from %s to %s.
`, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:       "me@here.com",
		From:     "me@here.com",
		Subject:  "Reservation Notification",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "generals.page.gohtml", &models.TemplateData{
		Form: forms.New(nil)})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "majors.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "search-availability.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// PostAvailability renders the search availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	startDateString := r.Form.Get("start")
	endDateString := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateString)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, endDateString)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]any)
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	err = render.Template(w, r, "choose-room.page.gohtml", &models.TemplateData{Data: data})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomId    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.Form.Get("start")
	endDateStr := r.Form.Get("end")

	fmt.Println("Here: ")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		fmt.Println("Here: ")
		helpers.ServerError(w, errors.New("error parsing start date"))
		return
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		fmt.Println("Here2: ")
		helpers.ServerError(w, errors.New("error parsing end date"))
		return
	}

	roomId, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		fmt.Println("Here3: ")

		helpers.ServerError(w, errors.New("error converting room id from string to int"))
		return
	}

	available, err := m.DB.SearchAvailabilityByDateByRoomId(startDate, endDate, roomId)
	if err != nil {
		fmt.Println("Here4: ")

		helpers.ServerError(w, errors.New("error searching for availability"))
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		RoomId:    strconv.Itoa(roomId),
		StartDate: startDateStr,
		EndDate:   endDateStr,
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	fmt.Println("JSON: ", out)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		fmt.Println("Here5: ")
		helpers.ServerError(w, err)
		return
	}

}

// Contact renders the search availability page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "contact.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// ReservationSummary renders the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get error from sessions")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]any)
	data["reservation"] = reservation

	startDateString := reservation.StartDate.Format("2006-01-02")
	endDateString := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = startDateString
	stringMap["end_date"] = endDateString

	err := render.Template(w, r, "reservation-summary.page.gohtml", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// ChooseRoom renders the choose-room page
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("reservation wasn't found in the session"))
		return
	}

	res.RoomId = roomId
	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom renders the book-room page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// id, s, e
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	startDateStr := r.URL.Query().Get("s")
	endDateStr := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomById(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var res models.Reservation
	res.RoomId = id
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// ShowLogin renders the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "login.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// PostShowLogin sends a POST request to the authentication
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		helpers.ClientError(w, http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	err := m.App.Session.Destroy(r.Context())
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashboard
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "admin-dashboard.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

}
