package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTExpiration(t *testing.T) {
	id := uuid.New()
	secret := "secret"

	jwt, err := MakeJWT(id, secret, time.Microsecond*1)
	if err != nil {
		t.Errorf("expected to make jwt")
	}

	time.Sleep(time.Millisecond * 1)

	_, err = ValidateJWT(jwt, secret)
	if err == nil {
		t.Errorf("expected validation to fail for expired token")
	}
}

func TestJWTSecret(t *testing.T) {
	id := uuid.New()
	secret := "secret"

	jwt, err := MakeJWT(id, secret, time.Microsecond*1)
	if err != nil {
		t.Errorf("expected to make jwt")
	}

	time.Sleep(time.Millisecond * 1)

	_, err = ValidateJWT(jwt, (secret + "not the same"))
	if err == nil {
		t.Errorf("expected validation to fail for mismatched secret")
	}
}

func TestJWTSuccess(t *testing.T) {
	id := uuid.New()
	secret := "secret"

	jwt, err := MakeJWT(id, secret, time.Minute*5)
	if err != nil {
		t.Errorf("expected to make jwt")
	}

	validatedId, err := ValidateJWT(jwt, secret)
	if err != nil {
		t.Errorf("expected jwt validation to succeed")
	}

	if validatedId != id {
		t.Errorf("expected id from jwt to match")
	}
}

func TestHashSuccess(t *testing.T) {
	pw := "some_password"

	hashedPw, err := HashPassword(pw)
	if err != nil {
		t.Errorf("expected password hashing to succeed")
	}

	err = CheckPasswordHash(pw, hashedPw)
	if err != nil {
		t.Errorf("expected hash and password to match")
	}
}

func TestHashFail(t *testing.T) {
	pw := "some_password"

	hashedPw, err := HashPassword(pw)
	if err != nil {
		t.Errorf("expected password hashing to succeed")
	}

	err = CheckPasswordHash((pw + "not the same"), hashedPw)
	if err == nil {
		t.Errorf("expected hash and password to not match")
	}
}

func TestGetBearerTokenSucceed(t *testing.T) {
	r, err := http.NewRequest("GET", "url", nil)
	if err != nil {
		t.Errorf("expected to get a request")
	}

	inputToken := "token"

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", inputToken))

	token, err := GetBearerToken(r.Header)
	if err != nil {
		t.Errorf("expected to get a token")
	}

	if token != inputToken {
		t.Errorf("expected tokens to match, instead %s != %s", token, inputToken)
	}
}

func TestGetBearerTokenFail(t *testing.T) {
	r, err := http.NewRequest("GET", "url", nil)
	if err != nil {
		t.Errorf("expected to get a request")
	}

	r.Header.Set("Authorization", "this is not a proper token")

	token, err := GetBearerToken(r.Header)
	if err == nil {
		t.Errorf("expected an error, instead got token '%s'", token)
	}
}
