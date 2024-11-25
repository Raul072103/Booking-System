package main

import (
	"encoding/gob"
	"flag"
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

var inProduction *bool
var useCache *bool
var dbHost *string
var dbName *string
var dbUser *string
var dbPass *string
var dbPort *string
var dbSSL *string

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main entry point in this server
func main() {
	parseCommandLineFlags()

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
	defer close(app.MailChan)

	listenForMail()

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

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// change this to true when in production

	app.InProduction = *inProduction
	app.UseCache = *useCache

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

	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return err
}

func connectToDB() (*driver.DB, error) {
	// connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
		return nil, err
	}
	log.Println("Connected to database")

	return db, nil
}

func parseCommandLineFlags() {
	inProduction = flag.Bool("production", true, "Application is in production")
	useCache = flag.Bool("cache", true, "Use template cache")
	dbHost = flag.String("dbhost", "localhost", "Database host")
	dbName = flag.String("dbname", "postgres", "Database name")
	dbUser = flag.String("dbuser", "postgres", "Database user")
	dbPass = flag.String("dbpass", "", "Database password")
	dbPort = flag.String("dbport", "5432", "Database port")
	dbSSL = flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, required)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}
}
