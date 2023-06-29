package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Config struct {
	Secret  string `toml:"secret"`
	ExpTime int    `toml:"exp_time"`
}

type JWT interface {
	CreateAccessToken(claim JWTClaim) (*JWTToken, error)
	CreateRefreshToken(claim JWTClaim) (*JWTToken, error)
	GetClaims(token string) (*JWTClaim, error)
}

type JWTClaim struct {
	jwt.StandardClaims
	InvitationID string `json:"invitation_id,omitempty"`
	Type         string `json:"type,omitempty"`
	GroupName    string `json:"group_name,omitempty"`
}

type JWTToken struct {
	ID        string
	Token     string
	Claim     JWTClaim
	ExpiresAt time.Time
	Scheme    string
}

type jwtImpl struct {
	cfg Config
}

func NewJWT(cfg Config) JWT {
	j := &jwtImpl{
		cfg: cfg,
	}

	return j
}

func (j *jwtImpl) signToken(claim *JWTClaim) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, err := token.SignedString([]byte(j.cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *jwtImpl) CreateAccessToken(claim JWTClaim) (*JWTToken, error) {
	now := time.Now()
	expAt := now.Add(time.Duration(j.cfg.ExpTime) * time.Minute)
	exp := expAt.Unix()
	iat := now.Unix()

	stdClaim := jwt.StandardClaims{
		ExpiresAt: exp,
		IssuedAt:  iat,
	}
	claim.StandardClaims = stdClaim

	signedToken, err := j.signToken(&claim)
	if err != nil {
		return nil, fmt.Errorf("Error creating access token: %w", err)
	}

	jwtToken := &JWTToken{
		Token:     signedToken,
		Claim:     claim,
		ExpiresAt: expAt,
		Scheme:    "Bearer",
	}

	return jwtToken, nil
}

func (j *jwtImpl) CreateRefreshToken(claim JWTClaim) (*JWTToken, error) {
	now := time.Now()
	expAt := now.Add(72 * time.Hour)
	exp := expAt.Unix()
	iat := now.Unix()
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	stdClaim := jwt.StandardClaims{
		Id:        uuid.String(),
		ExpiresAt: exp,
		IssuedAt:  iat,
	}
	claim.StandardClaims = stdClaim

	signedToken, err := j.signToken(&claim)
	if err != nil {
		return nil, fmt.Errorf("Error creating refresh token: %w", err)
	}

	jwtToken := &JWTToken{
		ID:        uuid.String(),
		Token:     signedToken,
		Claim:     claim,
		ExpiresAt: expAt,
		Scheme:    "Bearer",
	}

	return jwtToken, nil
}

func (j *jwtImpl) GetClaims(tokenString string) (*JWTClaim, error) {
	claim := &JWTClaim{}
	_, err := jwt.ParseWithClaims(tokenString, claim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	return claim, nil
}
