package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/raul/BookingSystem/internal/config"
	"github.com/raul/BookingSystem/internal/driver"
	"github.com/raul/BookingSystem/internal/handlers"
	"github.com/raul/BookingSystem/internal/helpers"
	"github.com/raul/BookingSystem/internal/models"
	"github.com/raul/BookingSystem/internal/render"
	"log"
	"net/http"
	"os"
	"time"
)

const PORT_NUMBER = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main entry point in this server
func main() {
	db, err := connectToDB()
	if err != nil {
		log.Fatal("Failed connecting to the database!")
	}
	repo := handlers.NewRepo(&app, db)

	err = run(repo)
	if err != nil {
		log.Fatal("Failed running server!")
	}
	defer db.SQL.Close()

	fmt.Println(fmt.Sprintf("Starting application on port %s", PORT_NUMBER))

	srv := &http.Server{
		Addr:    PORT_NUMBER,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run(repo *handlers.Repository) error {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	// change this to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache("C:\\Users\\raula\\Desktop\\facultate\\anul 3 sem 1\\Software Engineering\\Golang\\Learning\\BookingSystem\\templates")
	if err != nil {
		log.Fatal("Cannot create template cache")
		return err
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction

	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return err
}

func connectToDB() (*driver.DB, error) {
	// connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=postgres user=postgres password=")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
		return nil, err
	}
	log.Println("Connected to database")

	return db, nil
}
