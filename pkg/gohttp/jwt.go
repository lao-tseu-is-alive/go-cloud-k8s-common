package gohttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"net/http"
	"strings"
	"time"
)

// Context key for storing the JWT token
type contextKey string

const jwtTokenKey = contextKey("jwtToken")

type JwtInfo struct {
	Secret   string `json:"secret"`
	Duration int    `json:"duration"`
}

// JwtCustomClaims are custom claims extending default ones.
type JwtCustomClaims struct {
	jwt.RegisteredClaims
	Id       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

func ParseJwtTokenFunc(JwtSecret, jwtToken string, l golog.MyLogger) (*jwt.Token, error) {

	verifier, err := jwt.NewVerifierHS(jwt.HS512, []byte(JwtSecret))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in ParseJwtTokenFunc creating verifier: %s", err))
	}
	// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
	token, err := jwt.Parse([]byte(jwtToken), verifier)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in ParseJwtTokenFunc parsing token: %s", err))
	}
	// get REGISTERED claims
	var newClaims jwt.RegisteredClaims
	err = json.Unmarshal(token.Claims(), &newClaims)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in ParseJwtTokenFunc Unmarshaling RegisteredClaims: %s", err))
	}

	l.Debug("JWT ParseTokenFunc, Algorithm %v", token.Header().Algorithm)
	l.Debug("JWT ParseTokenFunc, Type      %v", token.Header().Type)
	l.Debug("JWT ParseTokenFunc, Claims    %v", string(token.Claims()))
	l.Debug("JWT ParseTokenFunc, Payload   %v", string(token.PayloadPart()))
	l.Debug("JWT ParseTokenFunc, Token     %v", string(token.Bytes()))
	l.Debug("JWT ParseTokenFunc, ParseTokenFunc : Claims:    %+v", string(token.Claims()))
	if newClaims.IsValidAt(time.Now()) {
		claims := JwtCustomClaims{}
		err := token.DecodeClaims(&claims)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error in ParseJwtTokenFunc unable to decode JwtCustomClaims: %s", err))
		}
		// maybe find a way to evaluate if user is de-activated ( like in a User microservice )
		//currentUserId := claims.Id
		//if store.IsUserActive(currentUserId) {
		//	return token, nil // ALL IS GOOD HERE
		//} else {
		// status RETURN 401 Unauthorized
		// return nil, errors.New("token invalid because user account has been deactivated")
		//}
		return token, nil // ALL IS GOOD HERE
	} else {
		l.Error("JWT ParseTokenFunc,  : IsValidAt(%+v)", time.Now())
		return nil, errors.New("token has expired")
	}

}

func JwtMiddleware(next http.Handler, JwtSecret string, l golog.MyLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return

		}
		// get the token from the request
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		// check if the token is valid
		token, err := ParseJwtTokenFunc(JwtSecret, tokenString, l)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		// Token is valid, proceed to the next handler
		// Store the valid JWT token in the request context
		ctx := context.WithValue(r.Context(), jwtTokenKey, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

// GetJwtCustomClaims returns the JWT Custom claims from the received context jwtdata
func GetJwtCustomClaims(r *http.Request) (JwtCustomClaims, error) {
	// Retrieve the JWT token from the request context
	token := r.Context().Value(jwtTokenKey).(*jwt.Token)
	claims := JwtCustomClaims{}
	err := token.DecodeClaims(&claims)
	return claims, err
}
