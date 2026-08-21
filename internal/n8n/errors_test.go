package n8n

import (
	"errors"
	"net/http"
	"testing"
)

func TestIsProjectRoleAdminUnlicensed(t *testing.T) {
	err := &APIError{
		StatusCode: http.StatusForbidden,
		Body:       `{"message":"Your license does not allow for feat:projectRole:admin. To enable feat:projectRole:admin, please upgrade to a license that supports this feature."}`,
	}
	if !IsProjectRoleAdminUnlicensed(err) {
		t.Fatal("expected licensed-feature 403")
	}
	if !IsProjectRoleAdminUnlicensed(errors.Join(errors.New("wrap"), err)) {
		t.Fatal("expected wrap to match")
	}
	if IsProjectRoleAdminUnlicensed(&APIError{StatusCode: http.StatusForbidden, Body: `{"message":"Forbidden"}`}) {
		t.Fatal("missing-scope 403 must not be treated as unlicensed")
	}
	if IsProjectRoleAdminUnlicensed(&APIError{StatusCode: http.StatusUnauthorized, Body: "feat:projectRole:admin"}) {
		t.Fatal("401 must not match")
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(&NotFoundError{Resource: "project", ID: "x"}) {
		t.Fatal("expected not found")
	}
	if IsNotFound(&APIError{StatusCode: 404}) {
		t.Fatal("APIError 404 is not NotFoundError")
	}
}
