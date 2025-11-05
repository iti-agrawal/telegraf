package gdchauth

import (
	"crypto"

	"github.com/golang-jwt/jwt/v5"
)

type ServiceAccountKey struct {
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	Project             string `json:"project"`
	ServiceIdentityName string `json:"name"`
	TokenURI            string `json:"token_uri"`

	parsedKey     crypto.Signer
	signingMethod jwt.SigningMethod
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
