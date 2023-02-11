package render

import (
	"net/http"
	"testing"

	"github.com/indramahaarta/bookings/internal/model"
)

func TestAddDefaultData(t *testing.T) {
	var td model.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)
	if result.Flash != "123" {
		t.Error("Flash value of 123 not found in session")
	}
}

func TestRenderTemplates(t *testing.T) {
	pathToTemplate = "./../../templates"

	tc, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	var ww myWriter

	err = Template(&ww, r, "home.page.html", &model.TemplateData{})

	if err != nil {
		t.Error(err)
	}

	err = Template(&ww, r, "homedoestnotexist.page.html", &model.TemplateData{})

	if err == nil {
		t.Error("Errors shoudn't be exist!")
	}

}

func TestNewTemplate(t *testing.T) {
	NewRenderer(&testApp)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplate = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
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
