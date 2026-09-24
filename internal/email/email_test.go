package email_test

import (
	"testing"

	"github.com/Uvic-ECSS/Lockers/internal/email"
)

func TestFormValidate(t *testing.T) {
	validEmail := []string{
		"foobar@uvic.ca",
		"Goobarba123z@uvic.ca",
		"Goobarba_123z@uvic.ca",
		"jo.smith@uvic.ca",
		"a-b@uvic.ca",
	}

	for _, addr := range validEmail {
		if !email.ValidUVicEmail(addr) {
			t.Fatal(addr)
		}
	}

	invalidEmails := []string{
		"foobar@uvic.caa",
		"Goobarba123z@uvic.com",
		"Goobarba_123z@gmail.uk",
		`"<img src=x onerror=alert(1)>"@uvic.ca`,
		`"quoted"@uvic.ca`,
		"foo<b>@uvic.ca",
		"foo bar@uvic.ca",
		"foo@uvic.ca@uvic.ca",
		"@uvic.ca",
		"foo@uvic.ca\n",
	}

	for _, addr := range invalidEmails {
		if email.ValidUVicEmail(addr) {
			t.Fatal(addr)
		}
	}
}
