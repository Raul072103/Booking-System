package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)
	form.Required("field1", "field2", "field3")

	isValid := form.Valid()
	if isValid {
		t.Error("expected invalid: got valid")
	}

	postedData := url.Values{}
	postedData.Add("field1", "x")
	postedData.Add("field2", "x")
	postedData.Add("field3", "x")

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("field1", "field2", "field3")
	if !form.Valid() {
		t.Error("expected valid: got invalid")
	}
}

func TestForm_Has(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	hasField1 := form.Has("field1")
	if hasField1 {
		t.Error("expected false: got true")
	}

	postedData := url.Values{}
	postedData.Add("field1", "x")

	r.PostForm = postedData
	form = New(r.PostForm)
	hasField1 = form.Has("field1")
	if !hasField1 {
		t.Error("expected true: got false")
	}
}

type emailExpectedResult struct {
	email          string
	expectedResult bool
}

var emailTest = map[string]emailExpectedResult{
	"email1": {email: "wrong.email", expectedResult: false},
	"email2": {email: "wrong.email242.com", expectedResult: false},
	"email3": {email: "wrong.", expectedResult: false},
	"email4": {email: "good.email@yahoo.com", expectedResult: true},
	"email5": {email: "raul@gmail.com", expectedResult: true},
	"email6": {email: "student@student.utcluj.ro", expectedResult: true},
}

func TestForm_IsEmail(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	postedData := url.Values{}
	for emailName, email := range emailTest {
		postedData.Add(emailName, email.email)
	}

	r.PostForm = postedData
	form := New(r.PostForm)

	for emailName, email := range emailTest {
		form.IsEmail(emailName)
		if email.expectedResult == false {
			if form.Errors.Get(emailName) == "" {
				t.Errorf("expected false for %s: but got true", emailName)
			}
		} else {
			if form.Errors.Get(emailName) != "" {
				t.Errorf("expected truje for %s: but got false", emailName)
			}
		}
	}

}

func TestForm_MinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	postedData := url.Values{}
	postedData.Add("field1", "abcdef")
	postedData.Add("field2", "abcdefghifsdfsdfsdj")
	postedData.Add("field3", "abcdefghijklmnop")

	r.PostForm = postedData
	form = New(r.PostForm)
	form.MinLength("field1", 10)
	form.MinLength("field2", 9)
	form.MinLength("field3", 100)

	if form.Errors.Get("field1") == "" {
		t.Errorf("expected min length = false for %s: but got true", "field1")
	}

	if form.Errors.Get("field2") != "" {
		t.Errorf("expected min length = true for %s: but got false", "field2")
	}

	if form.Errors.Get("field3") == "" {
		t.Errorf("expected min length = false for %s: but got true", "field3")
	}

}
