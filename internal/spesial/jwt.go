package spesial

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

type payloadType int

const (
	PayloadTypeAccess payloadType = iota
	PayloadTypeRefresh
)

type Payload struct {
	UserID uint `json:"-"`

	ID             uint        `json:"id"`
	RoleID         uuid.UUID   `json:"role_id"`
	Role           string      `json:"role"`
	PermissionMask string      `json:"permission_mask"`
	SessionID      string      `json:"sid"`
	InstituteID    uuid.UUID   `json:"institute_id"`
	Type           payloadType `json:"type"`
	ImpersonatedBy uint        `json:"imp_by,omitempty"`
	ImpersonatedAt int64       `json:"imp_at,omitempty"`
	jwt.StandardClaims
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
	jwt.StandardClaims
}

func GenerateAccessToken(p *Payload) (string, error) {
	expiredInMinutes := time.Duration(viper.GetInt("JWT_ACCESS_TOKEN_EXPIRY"))

	p.ExpiresAt = time.Now().Add(expiredInMinutes * time.Minute).Unix()
	return createToken(p)
}

func GenerateRefreshToken(p *Payload) (string, error) {
	expiredInMinutes := time.Duration(viper.GetInt("JWT_REFRESH_TOKEN_EXPIRY"))

	p.ExpiresAt = time.Now().Add(expiredInMinutes * time.Minute).Unix()
	return createRefreshToken(p)
}

func createRefreshToken(p *Payload) (string, error) {
	p.Type = PayloadTypeRefresh
	p.IssuedAt = time.Now().Unix()
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
	p.IssuedAt = time.Now().Unix()

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
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  iat.Unix(),
			ExpiresAt: expTime.Unix(),
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
		return []byte(jwtKey), nil
	}
}

func ParseBearerToken(r *http.Request) (string, error) {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return "", handling.NewHttpError401(nil)
	}

	var tokenString string
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	}
	return tokenString, nil
}

func ValidateToken(tokenStr string) (*Payload, error) {
	claims := &Payload{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, GetJWTKeyFunc())

	if err != nil || !token.Valid {
		log.Printf("JWT parse error: %T: %v", err, err)
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
	p.ExpiresAt = time.Now().Add(oneDay).Unix()
	return createToken(p)
}
