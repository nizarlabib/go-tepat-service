package spesial

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type payloadType int

const (
	PayloadTypeAccess payloadType = iota
	PayloadTypeRefresh
)

type Payload struct {
	UserID uint `json:"-"`

	JTI            string      `json:"jti"`
	ID             uint        `json:"id"`
	RoleID         uuid.UUID   `json:"role_id"`
	Role           string      `json:"role"`
	PermissionMask string      `json:"permission_mask"`
	SessionID      string      `json:"sid"`
	InstituteID    uuid.UUID   `json:"institute_id"`
	Type           payloadType `json:"type"`
	ImpersonatedBy uint        `json:"imp_by,omitempty"`
	ImpersonatedAt int64       `json:"imp_at,omitempty"`
	jwt.RegisteredClaims
}

func (p *Payload) IsImpersonated() bool {
	return p.ImpersonatedBy != 0
}

func (p *Payload) ReleaseImpersonation() {
	p.ImpersonatedBy = 0
	p.ImpersonatedAt = 0
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	Key   string `json:"key"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(p *Payload) (string, error) {
	expiredInMinutes := time.Duration(viper.GetInt("JWT_ACCESS_TOKEN_EXPIRY"))

	if p.JTI == "" {
		p.JTI = uuid.NewString()
	}
	p.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiredInMinutes * time.Minute))
	return createToken(p)
}

func GenerateRefreshToken(p *Payload) (string, error) {
	expiredInMinutes := time.Duration(viper.GetInt("JWT_REFRESH_TOKEN_EXPIRY"))

	if p.JTI == "" {
		p.JTI = uuid.NewString()
	}
	p.Type = PayloadTypeRefresh
	p.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiredInMinutes * time.Minute))
	return createRefreshToken(p)
}

func createRefreshToken(p *Payload) (string, error) {
	p.Type = PayloadTypeRefresh
	p.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
	jwtKey := viper.GetString("JWT_KEY")
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", handling.NewHttpError500(err)
	}
	return tokenString, nil
}

func createToken(p *Payload) (string, error) {
	issuer := viper.GetString("APP_ENV")
	issuer += ":" + viper.GetString("APP_SERVICE_NAME")
	issuer += "@" + viper.GetString("APP_VERSION")

	p.Issuer = issuer
	p.Type = PayloadTypeAccess
	p.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
	jwtKey := viper.GetString("JWT_KEY")
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", handling.NewHttpError500(err)
	}
	return tokenString, nil
}

func GenerateResetPasswordToken(email string, key string) (string, error) {
	duration := viper.GetInt("OTP_DURATION")
	if duration <= 0 {
		duration = 1800
	}
	expTime := time.Now().Add(time.Duration(duration) * time.Second)
	return createResetPasswordToken(email, expTime, key)
}

func ParseResetPasswordToken(tokenString string) (*ResetPasswordPayload, error) {
	claims := &ResetPasswordPayload{}
	jwtKey := []byte(viper.GetString("JWT_KEY"))
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.GetSigningMethod("HS256").Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, handling.NewHttpError(err, http.StatusBadRequest, "Token reset kata sandi sudah kedaluwarsa.", constant.ERR_TOKEN_EXPIRED)
	}
	return claims, nil
}

func createResetPasswordToken(email string, expTime time.Time, key string) (string, error) {
	iat := time.Now()
	payload := &ResetPasswordPayload{
		Email: email,
		Key:   key,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	jwtKey := viper.GetString("JWT_KEY")
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", handling.NewHttpError500(err)
	}
	return tokenString, nil
}

func GetJWTKeyFunc() jwt.Keyfunc {
	jwtKey := viper.GetString("JWT_KEY")
	return func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.GetSigningMethod("HS256").Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtKey), nil
	}
}

func ParseBearerToken(r *http.Request) (string, error) {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return "", handling.NewHttpError401(nil)
	}

	var tokenString string
	parts := strings.Split(bearerToken, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenString = parts[1]
	} else {
		return "", handling.NewHttpError401(errors.New("malformed authorization header"))
	}
	return tokenString, nil
}

func ValidateToken(tokenStr string) (*Payload, error) {
	claims := &Payload{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, GetJWTKeyFunc())

	if err != nil || !token.Valid {
		return nil, handling.NewHttpError401(err)
	}

	if claims.Role != "integration" && claims.SessionID == "" {
		return nil, handling.NewHttpError401(err)
	}

	if claims.Type != PayloadTypeAccess {
		return nil, handling.NewHttpError401(err)
	}

	return claims, nil
}

func GenerateIntegrationApiKeyToken(p *Payload) (string, error) {
	oneDay := time.Hour * 24
	p.ExpiresAt = jwt.NewNumericDate(time.Now().Add(oneDay))
	return createToken(p)
}
