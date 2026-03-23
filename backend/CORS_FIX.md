# CORS Fix Documentation

## Problem
Frontend requests from `http://localhost:5173` (React/Vite) to backend `http://localhost:8080` were being blocked by browser CORS policy.

## Root Cause
The CORS middleware was initially created but the container needed a full restart (`podman compose down && podman compose up`) to properly apply the changes. Container restart (`podman compose restart`) was not sufficient due to port binding conflicts.

## Solution Implemented

### 1. CORS Middleware (`internal/middleware/cors.go`)
Created a CORS middleware that:
- Validates origin against an allowlist
- Sets appropriate CORS headers (Origin, Methods, Headers, Credentials, Max-Age)
- Handles OPTIONS preflight requests with 204 No Content
- Passes other requests to the next handler

```go
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            // Check if origin is allowed
            allowed := false
            if origin != "" {
                for _, allowedOrigin := range allowedOrigins {
                    if origin == allowedOrigin || allowedOrigin == "*" {
                        allowed = true
                        break
                    }
                }
            }
            
            // Set CORS headers if origin is allowed
            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
                w.Header().Set("Access-Control-Allow-Credentials", "true")
                w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
            }
            
            // Handle preflight requests
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 2. Router Changes (`internal/router/router.go`)
Modified the Router struct to cache the CORS-wrapped handler:

```go
type Router struct {
    handler        http.Handler // CORS-wrapped mux
    mux            *http.ServeMux
    authHandler    *handler.AuthHandler
    jwtService     *auth.JWTService
    allowedOrigins []string
}

func New(db *database.DB, jwtService *auth.JWTService, authService *auth.Service) *Router {
    mux := http.NewServeMux()
    
    r := &Router{
        mux:         mux,
        authHandler: handler.NewAuthHandler(authService),
        jwtService:  jwtService,
        allowedOrigins: []string{
            "http://localhost:5173",
            "http://localhost:5174",
            "http://localhost:5175",
            "http://localhost:5176",
            "http://localhost:3000",
        },
    }

    r.setupRoutes(db)
    
    // Wrap mux with CORS middleware once
    r.handler = middleware.CORSMiddleware(r.allowedOrigins)(r.mux)
    
    return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.handler.ServeHTTP(w, req)
}
```

## Key Design Decisions

1. **Middleware wrapping at router initialization**: The CORS middleware wraps the entire mux once during router creation, avoiding repeated wrapper creation per request.

2. **Origin allowlist**: Explicitly lists allowed origins rather than using `*` to maintain security while allowing credentials.

3. **Preflight handling**: OPTIONS requests are handled by the middleware and return immediately with 204, preventing them from reaching route handlers that might reject OPTIONS.

4. **Header configuration**:
   - `Access-Control-Allow-Origin`: Echoes the request origin if allowed
   - `Access-Control-Allow-Methods`: All standard HTTP methods
   - `Access-Control-Allow-Headers`: Includes Authorization for JWT tokens
   - `Access-Control-Allow-Credentials`: true (required for cookies/auth)
   - `Access-Control-Max-Age`: 24 hours to reduce preflight requests

## Testing

### Test OPTIONS preflight:
```bash
curl -i -X OPTIONS http://localhost:8080/auth/login \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST"
```

Expected response:
```
HTTP/1.1 204 No Content
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: Accept, Authorization, Content-Type, X-CSRF-Token
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Max-Age: 86400
```

### Test actual POST request:
```bash
curl -i -X POST http://localhost:8080/auth/register \
  -H "Origin: http://localhost:5173" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"Test1234","role":"nurse"}'
```

Expected response includes CORS headers plus the 201 response body.

## Deployment Notes

When updating CORS middleware:
1. Make code changes
2. Run `podman compose build backend`
3. Run `podman compose down` followed by `podman compose up -d` (full restart required)
4. Verify with curl tests

**Note**: `podman compose restart backend` may not work due to port binding issues. Always use down/up for middleware changes.

## Future Considerations

- Add environment variable for allowed origins (currently hardcoded)
- Consider using a router library (chi, gorilla/mux) for more sophisticated middleware chaining
- Add CORS configuration to environment/config file for different environments (dev, staging, prod)
