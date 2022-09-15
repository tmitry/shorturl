package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type ContextKey string

const (
	jwtAlg = "HS256"
	jwtTyp = "jwt"
)

var ErrNoCorrectJWT = errors.New("jwt cookie not present or incorrect")

/*
JWTAuth middleware provide authentication using jwt (HS256).
It checks user and if user not exist or incorrect (including incorrect jwt signature) then creates a new.
The identifier is sent down the request context.
*/
func JWTAuth(signatureKey, jwtCookieName string, contextKeyUserID ContextKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		JWTAuthFunction := func(writer http.ResponseWriter, request *http.Request) {
			jwt, err := jwtRead(request, signatureKey, jwtCookieName)
			if err != nil && !errors.Is(err, ErrNoCorrectJWT) {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Println(err.Error())

				return
			}

			if err != nil {
				jwt, err = NewJWT(signatureKey)
				if err != nil {
					http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					log.Println(err.Error())

					return
				}

				err = jwtWrite(writer, jwt, jwtCookieName)
				if err != nil {
					http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					log.Println(err.Error())

					return
				}
			}

			request = request.WithContext(context.WithValue(request.Context(), contextKeyUserID, jwt.Payload.UserID))

			next.ServeHTTP(writer, request)
		}

		return http.HandlerFunc(JWTAuthFunction)
	}
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	UserID uuid.UUID `json:"user_id"`
}

type JWT struct {
	Header    jwtHeader
	Payload   jwtPayload
	Signature []byte
}

func NewJWT(signatureKey string) (*JWT, error) {
	userID := uuid.New()

	jwtHeader := jwtHeader{
		Alg: jwtAlg,
		Typ: jwtTyp,
	}

	jwtPayload := jwtPayload{
		UserID: userID,
	}

	jwt := &JWT{
		Header:  jwtHeader,
		Payload: jwtPayload,
	}

	var err error

	jwt.Signature, err = jwt.GenerateSignature(signatureKey)
	if err != nil {
		return nil, err
	}

	return jwt, nil
}

func (j JWT) GenerateSignature(signatureKey string) ([]byte, error) {
	jsonHeader, err := json.Marshal(j.Header)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal jwt header: %w", err)
	}

	base64Header := base64.StdEncoding.EncodeToString(jsonHeader)

	jsonPayload, err := json.Marshal(j.Payload)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal jwt payload: %w", err)
	}

	base64Payload := base64.StdEncoding.EncodeToString(jsonPayload)

	hash := hmac.New(sha256.New, []byte(signatureKey))
	hash.Write([]byte(fmt.Sprintf("%s.%s", base64Header, base64Payload)))

	return hash.Sum(nil), nil
}

func (j JWT) GetString() (string, error) {
	jsonHeader, err := json.Marshal(j.Header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal jwt header: %w", err)
	}

	header := base64.StdEncoding.EncodeToString(jsonHeader)

	jsonPayload, err := json.Marshal(j.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal jwt payload: %w", err)
	}

	payload := base64.StdEncoding.EncodeToString(jsonPayload)

	signature := base64.StdEncoding.EncodeToString(j.Signature)

	return fmt.Sprintf("%s.%s.%s", header, payload, signature), nil
}

func (j JWT) isValidHeader() bool {
	return j.Header.Typ == jwtTyp && j.Header.Alg == jwtAlg
}

func jwtParse(s string) (*JWT, error) {
	const jwtSize = 3

	jwtParts := strings.Split(s, ".")
	if len(jwtParts) != jwtSize {
		return nil, ErrNoCorrectJWT
	}

	base64Header := jwtParts[0]
	base64Payload := jwtParts[1]
	base64Signature := jwtParts[2]

	jsonHeader, err := base64.StdEncoding.DecodeString(base64Header)
	if err != nil {
		return nil, ErrNoCorrectJWT
	}

	jwtHeader := jwtHeader{
		Alg: "",
		Typ: "",
	}

	err = json.Unmarshal(jsonHeader, &jwtHeader)
	if err != nil {
		return nil, ErrNoCorrectJWT
	}

	jsonPayload, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		return nil, ErrNoCorrectJWT
	}

	jwtPayload := jwtPayload{
		UserID: uuid.UUID{},
	}

	err = json.Unmarshal(jsonPayload, &jwtPayload)
	if err != nil {
		return nil, ErrNoCorrectJWT
	}

	signature, err := base64.StdEncoding.DecodeString(base64Signature)
	if err != nil {
		return nil, ErrNoCorrectJWT
	}

	return &JWT{
		Header:    jwtHeader,
		Payload:   jwtPayload,
		Signature: signature,
	}, nil
}

func jwtWrite(writer http.ResponseWriter, jwt *JWT, jwtCookieName string) error {
	val, err := jwt.GetString()
	if err != nil {
		return err
	}

	jwtCookie := &http.Cookie{
		Name:  jwtCookieName,
		Value: val,
	}

	http.SetCookie(writer, jwtCookie)

	return nil
}

func jwtRead(request *http.Request, signatureKey, jwtCookieName string) (*JWT, error) {
	jwtCookie, err := request.Cookie(jwtCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, ErrNoCorrectJWT
		}

		return nil, fmt.Errorf("failed to read cookie: %w", err)
	}

	jwt, err := jwtParse(jwtCookie.Value)
	if err != nil {
		return nil, err
	}

	if !jwt.isValidHeader() {
		return nil, ErrNoCorrectJWT
	}

	correctSignature, err := jwt.GenerateSignature(signatureKey)
	if err != nil {
		return nil, err
	}

	if !hmac.Equal(correctSignature, jwt.Signature) {
		return nil, ErrNoCorrectJWT
	}

	return jwt, nil
}
