package models

import "testing"

func TestUserRegisterValidate(t *testing.T) {
	ok := UserRegister{Email: "john.doe@example.com", Password: "secret"}
	if err := ok.Validate(); err != nil {
		t.Errorf("valid email reported as invalid: %v", err)
	}

	bad := UserRegister{Email: "not‑mail", Password: "secret"}
	if err := bad.Validate(); err == nil {
		t.Errorf("invalid email accepted")
	}
}
