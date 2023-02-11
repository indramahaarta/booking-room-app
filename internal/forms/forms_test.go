package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/testing", nil)
	form := New(r.Form)

	isValid := form.Valid()

	if !isValid {
		t.Error("got invalid when it should be valid")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/anyone", nil)
	form := New(r.Form)

	data := []string{"a", "b", "c"}

	form.Required(data...)
	if form.Valid() {
		t.Error("There is no error, while it should be errors")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r.PostForm = postedData
	form = New(r.PostForm)

	form.Required(data...)
	if !form.Valid() {
		t.Error("There is errors, while it should be no errors")
	}

}

func TestFrom_IsEmail(t *testing.T) {
	r := httptest.NewRequest("POST", "/anyone", nil)
	postedData := url.Values{}
	postedData.Add("email", "indra@gmail.com")
	r.PostForm = postedData
	form := New(r.PostForm)

	form.IsEmail("email")
	if !form.Valid() {
		t.Error("There should be no error in field email")
	}

	postedData.Set("email", "indra@com")
	form = New(postedData)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("There should be an error in field email")
	}
}


func TestForm_MinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatsenfg", nil)
	postedData := url.Values{}
	postedData.Add("name", "indra")
	r.Form = postedData
	form := New(postedData)

	truth := form.MinLength("name", 3, r)
	if !truth {
		t.Error("There is an error, when it should be no error ")
	}


	postedDataFail := url.Values{}
	postedData.Add("name", "Joko")
	r.Form = postedDataFail
	form = New(postedDataFail)

	truth = form.MinLength("name", 4, r)
	if truth {
		t.Error("There is no error, when it should be an error")
	}
}

func TestFrom_Has(t *testing.T) {
	r := httptest.NewRequest("POST", "/anyonsks", nil)
	postedData := url.Values{}
	postedData.Add("a", "a")
	r.Form = postedData

	form := New(postedData)
	truth := form.Has("a", r)
	if !truth {
		t.Error("There is error, when it should be no error")
	}

	truth = form.Has("b", r)
	if !truth {
		err := form.Errors.Get("b")

		if err != "This field cannot be blank" {
			t.Error("Error message doesnt match")
		}
	} else {
		t.Error("There is no Error when it should be an error")
	}

	str := form.Errors.Get("a")
	if str != "" {
		t.Error("Error message should be empthy")
	}


}