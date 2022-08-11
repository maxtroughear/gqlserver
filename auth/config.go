package auth

type AuthConfig struct {
	// Enable Firebase auth
	FirebaseEnabled bool `env:"AUTH_FIREBASE_ENABLED"`

	// Path to Firebase credentials JSON file
	FirebaseCredentialsFile string `env:"AUTH_FIREBASE_CREDENTIALS_FILE"`

	// Firebase credentials JSON
	FirebaseCredentialsJSON string `env:"AUTH_FIREBASE_CREDENTIALS_JSON"`
}
