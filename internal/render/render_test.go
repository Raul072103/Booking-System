package render

import (
	"github.com/raul/BookingSystem/internal/models"
	"net/http"
	"testing"
)

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates := "./../../templates"
	_, err := CreateTemplateCache(pathToTemplates)
	if err != nil {
		t.Error(err)
	}
}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates := "./../../templates"
	tc, err := CreateTemplateCache(pathToTemplates)
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	ww := myWriter{}

	err = Template(&ww, r, "home.page.gohtml", &models.TemplateData{})
	if err != nil {
		t.Error("rendered template that doesn't exist")
	}
}

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "flash_value")
	session.Put(r.Context(), "warning", "warning_value")
	session.Put(r.Context(), "error", "error_value")
	result := AddDefaultData(&td, r)

	if result.Flash != "flash_value" {
		t.Error("Flash value different!")
	}

	if result.Warning != "warning_value" {
		t.Error("Warning value different!")
	}

	if result.Error != "error_value" {
		t.Error("Error value different!")
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}
