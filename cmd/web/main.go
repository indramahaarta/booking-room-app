package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/indramahaarta/bookings/internal/config"
	"github.com/indramahaarta/bookings/internal/driver"
	"github.com/indramahaarta/bookings/internal/handlers"
	"github.com/indramahaarta/bookings/internal/helpers"
	"github.com/indramahaarta/bookings/internal/model"
	"github.com/indramahaarta/bookings/internal/render"

	"github.com/alexedwards/scs/v2"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

const portNumber = ":8080"

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	log.Println("Starting mail listener...")
	listenForMail()

	// routing
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	log.Println("Starting Application on port ", portNumber)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}

func run() (*driver.DB, error) {
	// What i am going to store in session
	gob.Register(model.Reservation{})
	gob.Register(model.User{})
	gob.Register(model.Room{})
	gob.Register(model.Restriction{})

	// Handling Mail Channel
	mailChan := make(chan model.MailData)
	app.MailChan = mailChan

	// change this if you're in production
	app.InProduction = false

	// infolog & error log
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// session setting
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// Connect to database
	log.Println("Connecting to database ...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=user")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
		return nil, err
	}
	log.Println("Connect to database")

	// create template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		return db, err
	}
	app.TemplateCache = tc
	app.UseCache = false

	// send template into render and handlers
	render.NewRenderer(&app)
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandler(repo)
	helpers.NewHelpers(&app)

	return db, nil
}
