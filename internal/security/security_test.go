package security_test

import (
	"testing"

	"github.com/silvernemesis/jwtproject/internal/security"
)

func TestUser(t *testing.T) {
	_, err := security.AddUser("admin", "mypass", "admin")

	if err != nil {
		t.Error("error creating user admin")
	}

	_, err = security.VerifyUser("admin", "mypass")

	if err != nil {
		t.Error("error verifying user admin")
	}

	_, err = security.VerifyUser("admin", "mypass1")

	if err.Error() != "security: password does not match for username admin" {
		t.Error("error verifying user admin", err)
	}

	_, err = security.VerifyUser("admin1", "mypass")

	if err.Error() != "security: username admin1 does not exist" {
		t.Error("error verifying user admin", err)
	}
}
