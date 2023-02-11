package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/indramahaarta/bookings/internal/config"
	"github.com/indramahaarta/bookings/internal/driver"
	"github.com/indramahaarta/bookings/internal/forms"
	"github.com/indramahaarta/bookings/internal/helpers"
	"github.com/indramahaarta/bookings/internal/model"
	"github.com/indramahaarta/bookings/internal/render"
	"github.com/indramahaarta/bookings/internal/repository"
	"github.com/indramahaarta/bookings/internal/repository/dbrepo"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewTestingRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

func NewHandler(r *Repository) {
	Repo = r
}

// function that rendered template
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.html", &model.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.html", &model.TemplateData{})
}

func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(model.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New(strings.ToUpper("gaada cok di session")))
	}
	room, err := m.DB.GetRoomByID(res.RoomId)
	if err != nil {
		helpers.ServerError(w, err)
	}
	res.Room = room
	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	sm := make(map[string]string)
	sm["start_date"] = sd
	sm["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	form := forms.New(nil)
	render.Template(w, r, "make-reservation.page.html", &model.TemplateData{Form: form, Data: data, StringMap: sm})
}

func (m *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(model.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("CANNOT GET FROM SESSION"))
	}
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first-name")
	reservation.LastName = r.Form.Get("last-name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)
	form.Required("first-name", "last-name", "email", "phone")
	form.MinLength("first-name", 3, r)
	form.MinLength("phone", 10, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.html", &model.TemplateData{Form: form, Data: data})
		return
	}

	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	restriction := model.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomId:        reservation.RoomId,
		ReservationId: newReservationId,
		RestrictionId: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notification to user
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s, <br>
		This is confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := model.MailData{
		To:       reservation.Email,
		From:     "indramhrt@gmail.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	// send notification to user
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Nontification</strong><br>
		A reservation has been made from %s to %s.
	`, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = model.MailData{
		To:       "indramhrt@gmail.com",
		From:     "indramhrt@gmail.com",
		Subject:  "Reservation Nontification",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	m.App.Session.Put(r.Context(), "flash", "reservation has been made!")
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) Availbility(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "book.page.html", &model.TemplateData{})
}

func (m *Repository) PostAvailbility(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start-date")
	end := r.Form.Get("end-date")

	start = strings.ReplaceAll(start, "/", "-")
	end = strings.ReplaceAll(end, "/", "-")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchRoomAvailbilityByDate(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No Availbility")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	res := model.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms
	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "chooseroom.page.html", &model.TemplateData{Data: data})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(model.Reservation)
	if !ok {
		helpers.ServerError(w, err)
	}

	res.RoomId = roomId
	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusTemporaryRedirect)
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (m *Repository) AvailbilityJson(w http.ResponseWriter, r *http.Request) {

	layout := "2006-01-02"
	st := r.Form.Get("start-date")
	startDate, _ := time.Parse(layout, st)
	ed := r.Form.Get("end-date")
	endDate, _ := time.Parse(layout, ed)
	roomId, _ := strconv.Atoi(r.Form.Get("room-id"))

	avail, err := m.DB.SearchAvailbilityByDatesByRoomId(roomId, startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
	}

	resp := jsonResponse{}

	if avail {
		resp.OK = true
		resp.Message = "Room is Available"
	} else {
		resp.OK = false
		resp.Message = "Room is not Available"
	}

	out, err := json.MarshalIndent(resp, "", "    ")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.html", &model.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.html", &model.TemplateData{})
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomId, _ := strconv.Atoi(r.URL.Query().Get("id"))
	log.Println(roomId)
	sd := r.URL.Query().Get("sd")
	ed := r.URL.Query().Get("ed")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := model.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomId:    roomId,
	}
	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/make-reservation", http.StatusTemporaryRedirect)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &model.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(model.Reservation)

	if !ok {
		m.App.ErrorLog.Println("Can't get error From Session")
		m.App.Session.Put(r.Context(), "error", "Can't get error From Session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")
	m.App.Session.Put(r.Context(), "success", "Your Reservation has been recorded!")

	data := make(map[string]interface{})
	data["reservation"] = reservation
	render.Template(w, r, "reservation-summary.page.html", &model.TemplateData{Data: data, Flash: "Your Reservation has been recorded!"})
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &model.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	if !form.Valid() {
		// TODO - Take user back to login page
		stringMap := make(map[string]string)
		stringMap["email"] = email
		stringMap["password"] = password
		render.Template(w, r, "login.page.html", &model.TemplateData{Form: form, StringMap: stringMap})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid error credential")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "succesfully login")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
