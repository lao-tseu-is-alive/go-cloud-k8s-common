package gohttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/rs/xid"
	"net/http"
	"strings"
	"time"
)

type JwtChecker interface {
	ParseToken(jwtToken string) (*JwtCustomClaims, error)
	GetTokenFromUserInfo(userInfo *UserInfo) (*jwt.Token, error)
	JwtMiddleware(next http.Handler) http.Handler
	GetLogger() golog.MyLogger
	GetJwtDuration() int
	GetIssuerId() string
}

// Context key for storing the JWT token
type contextKey string

const jwtTokenKey = contextKey("jwtToken")

// UserInfo are custom claims extending default ones.
type UserInfo struct {
	UserId    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	UserLogin string `json:"user_login"`
	IsAdmin   bool   `json:"is_admin"`
}

// JwtCustomClaims are custom claims extending default ones.
type JwtCustomClaims struct {
	jwt.RegisteredClaims
	User *UserInfo
}

type JwtInfo struct {
	Secret   string `json:"secret"`
	Duration int    `json:"duration"`
	IssuerId string `json:"issuer_id"`
	Subject  string `json:"subject"`
	logger   golog.MyLogger
}

func (ji *JwtInfo) ParseToken(jwtToken string) (*JwtCustomClaims, error) {
	l := ji.logger
	// create a new verifier
	verifier, err := jwt.NewVerifierHS(jwt.HS512, []byte(ji.Secret))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in ParseToken creating verifier: %s", err))
	}
	// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
	token, err := jwt.Parse([]byte(jwtToken), verifier)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in ParseToken parsing token: %s", err))
	}
	// get REGISTERED claims
	var newClaims jwt.RegisteredClaims
	err = json.Unmarshal(token.Claims(), &newClaims)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in ParseToken Unmarshaling RegisteredClaims: %s", err))
	}

	l.Debug("JWT ParseToken, Algorithm %v", token.Header().Algorithm)
	l.Debug("JWT ParseToken, Type      %v", token.Header().Type)
	l.Debug("JWT ParseToken, Claims    %v", string(token.Claims()))
	l.Debug("JWT ParseToken, Payload   %v", string(token.PayloadPart()))
	l.Debug("JWT ParseToken, Token     %v", string(token.Bytes()))
	l.Debug("JWT ParseToken, ParseTokenFunc : Claims:    %+v", string(token.Claims()))
	if newClaims.IsValidAt(time.Now()) {
		claims := JwtCustomClaims{}
		err := token.DecodeClaims(&claims)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error in ParseToken unable to decode JwtCustomClaims: %s", err))
		}
		l.Debug("JWT ParseToken,  : claims.ID", claims.ID)
		l.Debug("JWT ParseToken,  : claims.UserId", claims.User.UserId)
		l.Debug("JWT ParseToken,  : claims.UserLogin", claims.User.UserLogin)
		l.Debug("JWT ParseToken,  : claims.UserName", claims.User.UserName)
		l.Debug("JWT ParseToken,  : claims.UserEmail", claims.User.UserEmail)
		// maybe find a way to evaluate if User is de-activated ( like in a User microservice )
		//currentUserId := claims.UserId
		//if store.IsUserActive(currentUserId) {
		//	return token, nil // ALL IS GOOD HERE
		//} else {
		// status RETURN 401 Unauthorized
		// return nil, errors.New("token invalid because User account has been deactivated")
		//}
		return &claims, nil // ALL IS GOOD HERE
	} else {
		l.Error("JWT ParseTokenFunc,  : IsValidAt(%+v)", time.Now())
		return nil, errors.New("token has expired")
	}

}

func (ji *JwtInfo) JwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			TraceRequest("JwtMiddleware-AuthorizationHeaderMissing", r, ji.logger)
			// ok so classic Auth header is missing, let's check if sec-websocket-protocol is present in case of websocket
			authHeader = r.Header.Get("Sec-Websocket-Protocol")
			if authHeader == "" {
				ji.logger.Error("JwtMiddleware : Authorization header missing")
				http.Error(w, "Authorization header missing", http.StatusBadRequest)
				return
			}
		}
		// get the token from the request
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		tokenString = strings.Replace(authHeader, "Authorization, ", "", 1)
		// check if the token is valid
		jwtClaims, err := ji.ParseToken(tokenString)
		if err != nil {
			TraceRequest("JwtMiddleware-InvalidJwtToken", r, ji.logger)
			ji.logger.Error("Invalid token err: %s\ntoken: %s", err, tokenString)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		// Token is valid, proceed to the next handler
		// Store the valid JWT token in the request context
		ji.logger.Debug(fmt.Sprintf("JwtMiddleware : user: %s got valid Token %s", jwtClaims.User.UserLogin, jwtClaims.ID))
		ctx := context.WithValue(r.Context(), jwtTokenKey, jwtClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func (ji *JwtInfo) GetLogger() golog.MyLogger {
	return ji.logger
}

func (ji *JwtInfo) GetJwtDuration() int {
	return ji.Duration
}

func (ji *JwtInfo) GetIssuerId() string {
	return ji.IssuerId
}

func (ji *JwtInfo) GetTokenFromUserInfo(userInfo *UserInfo) (*jwt.Token, error) {
	guid := xid.New()
	claims := &JwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        guid.String(), // this is the JWT TOKEN ID
			Audience:  nil,
			Issuer:    ji.GetIssuerId(), // this is the JWT TOKEN ISSUER
			Subject:   "",
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Minute * time.Duration(ji.GetJwtDuration()))},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			NotBefore: &jwt.NumericDate{Time: time.Now()},
		},
		User: userInfo,
	}
	// Create token with claims
	signer, _ := jwt.NewSignerHS(jwt.HS512, []byte(ji.Secret))
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claims)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in GetTokenFromUserInfo: %s", err))
	}
	return token, nil
}

// GetJwtCustomClaimsFromContext returns the JWT Custom claims from the received  request with context
func GetJwtCustomClaimsFromContext(r *http.Request) *JwtCustomClaims {
	// Retrieve the JWT Claims from the request context
	jwtClaims := r.Context().Value(jwtTokenKey).(*JwtCustomClaims)
	//claims := JwtCustomClaims{}
	//err := token.DecodeClaims(&claims)
	return jwtClaims
}

// NewJwtChecker creates a new JwtChecker
func NewJwtChecker(secret, issuer, subject string, duration int, l golog.MyLogger) JwtChecker {
	return &JwtInfo{
		Secret:   secret,
		Duration: duration,
		IssuerId: issuer,
		Subject:  subject,
		logger:   l,
	}
}
