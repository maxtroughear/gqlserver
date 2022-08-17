package auth

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4/request"
	"github.com/maxtroughear/gqlserver/middleware"
	"google.golang.org/api/option"
)

type firebaseAuthTokenContextKey struct{}

type firebaseAuthContextKey struct{}

type FirebaseAuth struct {
	app          *firebase.App
	jwtExtractor request.Extractor
}

func NewFirebaseAuth(cfg AuthConfig) FirebaseAuth {
	opts := []option.ClientOption{}
	if cfg.FirebaseCredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.FirebaseCredentialsFile))
	} else if cfg.FirebaseCredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.FirebaseCredentialsJSON)))
	}

	app, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		log.Fatalf("error initialising firebase app: %v", err)
	}

	return FirebaseAuth{
		app: app,
		jwtExtractor: request.MultiExtractor{
			request.AuthorizationHeaderExtractor,
			cookieTokenExtractor{},
		},
	}
}

// FirebaseAuthMiddleware retrieves and verifies a Firebase auth token via
// a header or cookie and passes the token into the current context.
// It also adds the FirebaseAuth instance to the current context
func (a *FirebaseAuth) FirebaseAuthMiddleware() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ctx := ginContext.Request.Context()

		log := middleware.LogrusFromContext(ctx)

		{
			newCtx := context.WithValue(ctx, firebaseAuthContextKey{}, a)
			ginContext.Request = ginContext.Request.WithContext(newCtx)
		}

		client, err := a.app.Auth(ctx)
		if err != nil {
			log.Errorf("error getting firebase Auth client: %v", err)
			ginContext.Next()
			return
		}

		idToken, err := a.jwtExtractor.ExtractToken(ginContext.Request)
		if err != nil {
			log.Debugf("failed to retrieve firebase auth token from request: %v", err)
			// skip if no token extracted
			ginContext.Next()
			return
		}

		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			log.Warnf("failed to verify firebase auth token: %v", err)
		}

		log.WithField("firebase.uid", token.UID).
			Debugf("firebase auth token verified")

		{
			newCtx := context.WithValue(ctx, firebaseAuthTokenContextKey{}, token)
			ginContext.Request = ginContext.Request.WithContext(newCtx)
		}
		ginContext.Next()
	}
}

// FirebaseAuthSetUserClaims sets the user with the passed uid's token claims.
// This can then be verified either with the middleware associated with this struct
// or with any standard JWT verification process.
func (a *FirebaseAuth) FirebaseAuthSetUserClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	client, err := a.app.Auth(ctx)
	if err != nil {
		return err
	}

	return client.SetCustomUserClaims(ctx, uid, claims)
}

// FirebaseAuthTokenFromContext retrives the verified Firebase auth token
// from the current context
func FirebaseAuthTokenFromContext(ctx context.Context) *auth.Token {
	token, ok := ctx.Value(firebaseAuthTokenContextKey{}).(*auth.Token)
	if !ok {
		return nil
	}
	return token
}

// FirebaseAuthFromContext retrives FirebaseAuth from the current context
func FirebaseAuthFromContext(ctx context.Context) *FirebaseAuth {
	firebaseAuth, ok := ctx.Value(firebaseAuthContextKey{}).(*FirebaseAuth)
	if !ok {
		return nil
	}
	return firebaseAuth
}

type cookieTokenExtractor struct{}

func (c cookieTokenExtractor) ExtractToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", request.ErrNoTokenInRequest
	}
	return cookie.Value, nil
}
