package middleware

// Auth and role checks currently live in internal/handler because they need
// template flashes, session cookies and user loading. This package is reserved
// for reusable net/http middleware that does not depend on handler internals.
