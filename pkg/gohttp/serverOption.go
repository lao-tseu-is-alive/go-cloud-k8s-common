package gohttp

import "net/http"

type ServerOption func(*Server)

func WithAuthentication(auth Authentication) ServerOption {
	return func(s *Server) {
		s.Authenticator = auth
	}
}

// WithJwtChecker is an option to set the JwtChecker.
func WithJwtChecker(jwtCheck JwtChecker) ServerOption {
	return func(s *Server) {
		s.JwtCheck = jwtCheck
	}
}

// WithProtectedRoutes is an option to add JWT-protected routes.
func WithProtectedRoutes(jwtCheck JwtChecker, protectedRoutes map[string]http.HandlerFunc) ServerOption {
	return func(s *Server) {
		if s.JwtCheck == nil {
			s.JwtCheck = jwtCheck
		}

		for path, handler := range protectedRoutes {
			// The middleware is now applied here, specifically for these routes
			s.router.Handle(path, s.JwtCheck.JwtMiddleware(handler))
		}
	}
}

func WithVersionReader(ver VersionReader) ServerOption {
	return func(s *Server) {
		s.VersionReader = ver
	}
}
