package main

import (
	"github.com/indramahaarta/bookings/internal/config"
	"github.com/indramahaarta/bookings/internal/handlers"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.PostMakeReservation)
	mux.Get("/search-availability", handlers.Repo.Availbility)
	mux.Post("/search-availability", handlers.Repo.PostAvailbility)
	mux.Post("/search-availability-json", handlers.Repo.AvailbilityJson)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/rooms/generals-quarters", handlers.Repo.Generals)
	mux.Get("/rooms/majors-suites", handlers.Repo.Majors)
	mux.Get("/book-room", handlers.Repo.BookRoom)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)
	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostLogin)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	
	return mux
}
