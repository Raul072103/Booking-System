package main

import (
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/raul/BookingSystem/pkg/config"
	"github.com/raul/BookingSystem/pkg/handlers"
	"github.com/raul/BookingSystem/pkg/render"
	"log"
	"net/http"
	"time"
)

const PORT_NUMBER = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the main entry point in this server
func main() {
	var app config.AppConfig

	// change this to true when in prouction
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)

	//http.HandleFunc("/", handlers.Repo.Home)
	//http.HandleFunc("/about", handlers.Repo.About)

	fmt.Println(fmt.Sprintf("Starting application on port %s", PORT_NUMBER))
	// _ = http.ListenAndServe(PORT_NUMBER, nil)

	srv := &http.Server{
		Addr:    PORT_NUMBER,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
