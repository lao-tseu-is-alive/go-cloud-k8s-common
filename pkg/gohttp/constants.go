package gohttp

import "time"

const (
	defaultProtocol            = "http"
	defaultSecondsToSleep      = 3
	secondsShutDownTimeout     = 5 * time.Second  // maximum number of second to wait before closing server
	defaultReadTimeout         = 10 * time.Second // max time to read request from the client
	defaultWriteTimeout        = 10 * time.Second // max time to write response to the client
	defaultIdleTimeout         = 2 * time.Minute  // max time for connections using TCP Keep-Alive
	htmlHeaderStart            = `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css"/>`
	charsetUTF8                = "charset=UTF-8"
	MIMEAppJSON                = "application/json"
	MIMEHtml                   = "text/html"
	MIMEAppJSONCharsetUTF8     = MIMEAppJSON + "; " + charsetUTF8
	HeaderContentType          = "Content-Type"
	httpErrMethodNotAllow      = "ERROR: Http method not allowed"
	initCallMsg                = "INITIAL CALL TO %s()\n"
	defaultNotFound            = "404 page not found"
	defaultNotFoundDescription = "🤔 ℍ𝕞𝕞... 𝕤𝕠𝕣𝕣𝕪 :【𝟜𝟘𝟜 : ℙ𝕒𝕘𝕖 ℕ𝕠𝕥 𝔽𝕠𝕦𝕟𝕕】🕳️ 🔥"
	formatErrRequest           = "ERROR: Http method not allowed [%s] %s  path:'%s', RemoteAddrIP: [%s]\n"
	fmtErrNewRequest           = "### ERROR http.NewRequest %s on [%s] error is :%v\n"
)
