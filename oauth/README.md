# OAuth Go Service Middleware
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/oauth)

OAuth 2.0 middleware that supports HS256 (HMAC) token validation

1.  [Prerequisites](#prerequisites)
2.  [Info](#info)
3.  [GRPC Interceptors](#grpc-interceptors)
4.  [HTTP Middleware](#http-middleware)

## Prerequisites
1.  [Go 1.9](https://golang.org/dl/)

## Info

DIAGRAM GOES HERE

```
1. Service receives token
2. Service locally validates token
3. Service queries OAuth server for token validation
4. Service receives user viewing permissions
3. Service sends back data based on user viewing permissions
```

## Usage

1.  Go tooling:
```
go get -u github.com/deciphernow/gm-fabric-go/oauth
```

2.  Using [dep](https://github.com/golang/dep)
```
dep add github.com/deciphernow/gm-fabric-go/oauth
```

## With Dex

### GRPC Interceptors

1.  If you wish to use interceptors, follow this approach. We rely on [go-grpc-middleware grpc_auth](https://github.com/grpc-ecosystem/go-grpc-middleware)
```go
// Inject token validation as a middleware interceptor
// This use case has less flexibility. We recommend the other way of validating
ctx := oauth.ContextWithOptions(context.Background(), oauth.WithProvider("http://127.0.0.1:5556/dex"), oauth.WithClientID("example-app"))

server := grpc.NewServer(
    grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(oauth.ValidateToken(ctx))),
    grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(oauth.ValidateToken(ctx))),
)

return server
```

2.  Alternatively (recommended), you can validate on server methods individually. We recommend this route as you have more control over what data should be blocked.
```go
func (s *Server) GetItems(ctx context.Context, in *store.Item) (*store.Items, error) {
    // Auth check
    _, err := oauth.ValidateToken(
        oauth.ContextWithOptions(
            ctx,
            oauth.WithProvider("http://127.0.0.1:5556/dex"),
            oauth.WithClientID("example-app"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // If token validation was successful, get user permissions
    permissions := oauth.RetrievePermissionsFromContext(ctx)
    if permissions == nil {
        return nil, errors.New("user permissions can not be nil")
    }

    // Do logic
    // Only return items a user is allowed to see

    return items, nil
}
```

### HTTP Middleware

```go
// NOTE: If the token is base64 encoded, decoded it before passing into the middleware

// Inject the JWT middleware
stack := middleware.Chain(
    // Adding this to the stack will require all API queries to provide a token in the auth http header
    // Always pass the signing algorithm to expect and any necessary supporting data such as a signing secret or key path
    oauth.HTTPAuthenticate(oauth.WithProvider("http://127.0.0.1:5556/dex"), oauth.WithClientID("example-app")),
)

// Create basic HTTP server
s := http.Server{
    Addr: "0.0.0.0:8080",
    // Wrap our mux router in the middleware stack
    Handler: stack.Wrap(router),
}

// Start the server
s.ListenAndServe()
```


## Without Dex

### GRPC Interceptors
1.  If you wish to use interceptors, follow this approach. We rely on [go-grpc-middleware grpc_auth](https://github.com/grpc-ecosystem/go-grpc-middleware)
```go
// Inject token validation as a middleware interceptor
// This use case has less flexibility. We recommend the other way of validating
ctx := oauth.ContextWithOptions(oauth.WithSigningAlg(HS256), oauth.WithTokenSecret("KbtfnXOHH3ezniXIsHYSd4WhZPBXH1vB"))

server := grpc.NewServer(
    grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(oauth.ValidateToken(ctx))),
    grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(oauth.ValidateToken(ctx))),
)

return server
```

2.  Alternatively (recommended), you can validate on server methods individually. We recommend this route as you have more control over what data should be blocked.
```go
func (s *Server) GetItems(ctx context.Context, in *store.Item) (*store.Items, error) {
    // Auth check
    _, err := oauth.ValidateToken(
        oauth.ContextWithOptions(
            ctx,
            oauth.WithProvider("http://127.0.0.1:5556/dex"),
            oauth.WithClientID("example-app"),
        ),
    )

    // If token validation was successful, get user permissions
    permissions := oauth.RetrievePermissionsFromContext(ctx)
    if permissions == nil {
        return nil, errors.New("user permissions can not be nil")
    }

    // Do logic
    // Only return items a user is allowed to see

    return items, nil
}
```

### HTTP Middleware

```go
// NOTE: If the token is base64 encoded, decoded it before passing into the middleware

// Inject the JWT middleware
stack := middleware.Chain(
    // Adding this to the stack will require all API queries to provide a token in the auth http header
    // Always pass the signing algorithm to expect and any necessary supporting data such as a signing secret or key path
    oauth.HTTPAuthenticate(oauth.WithSigningAlg(HS256), oauth.WithHMACSecret("KbtfnXOHH3ezniXIsHYSd4WhZPBXH1vB")),
)

// Create basic HTTP server
s := http.Server{
    Addr: "0.0.0.0:8080",
    // Wrap our mux router in the middleware stack
    Handler: stack.Wrap(router),
}

// Start the server
s.ListenAndServe()
```
