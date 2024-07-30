package gohttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"time"
)

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
