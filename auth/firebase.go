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
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
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

		tx := newrelic.FromContext(ctx)
		var segment *newrelic.Segment
		if tx != nil {
			segment = tx.StartSegment("Gin/Middleware/FirebaseAuth")
		}

		// always continue through middleware pipeline
		defer func(ginContext *gin.Context, segment *newrelic.Segment) {
			if segment != nil {
				segment.End()
			}

		}(ginContext, segment)

		ctx = context.WithValue(ctx, firebaseAuthContextKey{}, a)

		token, err := authenticateUser(ctx, ginContext, a)
		userAuthenticated := token != nil && err == nil
		if err != nil {
			log.WithError(err).Error("error while attempting to authenticate user. request continuing")
		} else if userAuthenticated {
			log.WithFields(logrus.Fields{
				"firebase.uid":      token.UID,
				"firebase.issueAt":  token.IssuedAt,
				"firebase.expires":  token.Expires,
				"firebase.authTime": token.AuthTime,
			}).Debugf("firebase auth token verified")

			if segment != nil {
				segment.AddAttribute("firebase.uid", token.UID)
				segment.AddAttribute("firebase.issueAt", token.IssuedAt)
				segment.AddAttribute("firebase.expires", token.Expires)
				segment.AddAttribute("firebase.authTime", token.AuthTime)
			}

			ctx = context.WithValue(ctx, firebaseAuthTokenContextKey{}, token)
		} else {
			log.Debugf("firebase auth token missing")
		}

		ginContext.Request = ginContext.Request.WithContext(ctx)

		if segment != nil {
			segment.End()
		}

		ginContext.Next()
	}
}

func authenticateUser(ctx context.Context, ginContext *gin.Context, a *FirebaseAuth) (*auth.Token, error) {
	client, err := a.app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	idToken, err := a.jwtExtractor.ExtractToken(ginContext.Request)
	if err == request.ErrNoTokenInRequest {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return token, err
	}

	return token, nil
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
