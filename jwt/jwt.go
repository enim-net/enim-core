package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RoleClaim struct {
	Roles []string `json:"role"`
}

type Claims struct {
	jwt.RegisteredClaims
	Type            string    `json:"typ,omitempty"`
	AuthorizedParty string    `json:"azp,omitempty"`
	SessionID       string    `json:"sid,omitempty"`
	Scope           string    `json:"scope,omitempty"`
	RealmAccess     RoleClaim `json:"realm_access,omitempty"`
	Email           string    `json:"email,omitempty"`
	EmailVerified   bool      `json:"email_verified,omitempty"`
	Name            string    `json:"name,omitempty"`
}
type Principal struct {
	UserID     string
	Username   string
	Email      string
	Name       string
	ClientID   string
	SessionID  string
	Scope      string
	RealmRoles []string
}
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	baseURL    string
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration, baseURL string) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		baseURL:    baseURL,
	}
}

func (m *Manager) GenerateAccess(realm string, p Principal) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.accessTTL)
	claims := Claims{
		RegisteredClaims: m.registered(realm, p, now, exp),
		Type:             "Bearer",
		AuthorizedParty:  p.ClientID,
		SessionID:        p.SessionID,
		Scope:            p.Scope,
		RealmAccess:      RoleClaim{Roles: p.RealmRoles},
		Email:            p.Email,
		Name:             p.Name,
	}
	signed, err := m.sign(claims)
	return signed, exp, err
}

func (m *Manager) GenerateRefresh(realm string, p Principal) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.refreshTTL)

	claims := Claims{
		RegisteredClaims: m.registered(realm, p, now, exp),
		Type:             "Refresh",
		AuthorizedParty:  p.ClientID,
		SessionID:        p.SessionID,
		Scope:            p.Scope,
	}
	signed, err := m.sign(claims)
	return signed, exp, err
}

func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (m *Manager) issuer(realm string) string {
	return m.baseURL + "/realms/" + realm
}

func (m *Manager) sign(claims Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}
func (m *Manager) registered(realm string, p Principal, now, exp time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    m.issuer(realm),
		Subject:   p.UserID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
}
