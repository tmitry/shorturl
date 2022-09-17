package middlewares_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app/middlewares"
)

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	type args struct {
		signatureKey     string
		jwtCookieName    string
		contextKeyUserID middlewares.ContextKey
		jwt              *middlewares.JWT
		isCorrect        bool
	}

	jwtIncorrectSignature, err := middlewares.NewJWT("incorrect_signature_key")
	require.NoError(t, err)

	jwt, err := middlewares.NewJWT("signature_key")
	require.NoError(t, err)

	tests := []struct {
		name string
		args args
		want func(next http.Handler) http.Handler
	}{
		{
			name: "incorrect empty jwt - generates new user uuid",
			args: args{
				signatureKey:     "signature_key",
				jwtCookieName:    "jwt_cookie",
				contextKeyUserID: "UserID",
				jwt:              nil,
				isCorrect:        false,
			},
		},
		{
			name: "incorrect jwt signature - generates new user uuid",
			args: args{
				signatureKey:     "signature_key",
				jwtCookieName:    "jwt",
				contextKeyUserID: "ID",
				jwt:              jwtIncorrectSignature,
				isCorrect:        false,
			},
		},
		{
			name: "correct jwt - uses uuid from request",
			args: args{
				signatureKey:     "signature_key",
				jwtCookieName:    "JWT",
				contextKeyUserID: "userID",
				jwt:              jwt,
				isCorrect:        true,
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()

			router := chi.NewRouter()

			router.Use(middlewares.JWTAuth(
				testCase.args.signatureKey,
				testCase.args.jwtCookieName,
				testCase.args.contextKeyUserID,
			))
			router.Post("/", func(writer http.ResponseWriter, r *http.Request) {
				userID, ok := r.Context().Value(testCase.args.contextKeyUserID).(uuid.UUID)
				if !ok {
					return
				}

				response, _ := userID.MarshalBinary()

				_, err := writer.Write(response)
				if err != nil {
					return
				}
			})

			body := []byte("The best content in the world. Than you for your attention.")
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))

			if testCase.args.jwt != nil {
				val, err := testCase.args.jwt.GetString()
				assert.NoError(t, err)

				cookie := &http.Cookie{
					Name:  testCase.args.jwtCookieName,
					Value: val,
				}
				req.AddCookie(cookie)
			}

			router.ServeHTTP(recorder, req)

			_, err := uuid.FromBytes(recorder.Body.Bytes())
			assert.NoError(t, err)

			if testCase.args.jwt != nil {
				uuidBinary, err := testCase.args.jwt.Payload.UserID.MarshalBinary()
				assert.NoError(t, err)

				if testCase.args.isCorrect {
					assert.Equal(t, recorder.Body.Bytes(), uuidBinary)
				} else {
					assert.NotEqual(t, recorder.Body.Bytes(), uuidBinary)
				}
			}

			res := recorder.Result()
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, res.StatusCode, http.StatusOK)
		})
	}
}
