package model

import "github.com/indramahaarta/bookings/internal/forms"

type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float64
	Data            map[string]interface{}
	CSRFToken       string
	Flash           string
	Error           string
	Warning         string
	Form            *forms.Form
	IsAuthenticated int
}
