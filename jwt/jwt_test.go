package jwt

import (
	"strings"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

var secret = strings.Repeat("k", 32)

func TestAccessRoundTrip(t *testing.T) {
	m := NewManager(secret, time.Minute, time.Hour, "https://id.example")
	p := Principal{UserID: "u1", Email: "a@b.co", Name: "A", ClientID: "web", SessionID: "s1", Scope: "openid", RealmRoles: []string{"admin"}}
	tok, exp, err := m.GenerateAccess("isp", p)
	if err != nil || time.Until(exp) > time.Minute || time.Until(exp) <= 0 {
		t.Fatalf("generate: %v %v", exp, err)
	}
	c, err := m.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject != "u1" || c.Type != "Bearer" || c.Issuer != "https://id.example/realms/isp" || c.RealmAccess.Roles[0] != "admin" || c.AuthorizedParty != "web" {
		t.Fatalf("claims %+v", c)
	}
	r, _, _ := m.GenerateRefresh("isp", p)
	if c, err := m.Verify(r); err != nil || c.Type != "Refresh" {
		t.Fatalf("refresh: %+v %v", c, err)
	}
}

func TestVerifyRejects(t *testing.T) {
	m := NewManager(secret, time.Minute, time.Hour, "x")
	other := NewManager(strings.Repeat("o", 32), time.Minute, time.Hour, "x")
	tok, _, _ := other.GenerateAccess("r", Principal{UserID: "u"})
	if _, err := m.Verify(tok); err == nil {
		t.Error("wrong secret accepted")
	}
	expired := NewManager(secret, -time.Minute, time.Hour, "x")
	tok, _, _ = expired.GenerateAccess("r", Principal{UserID: "u"})
	if _, err := m.Verify(tok); err == nil {
		t.Error("expired token accepted")
	}
	none, _ := gojwt.NewWithClaims(gojwt.SigningMethodNone, gojwt.MapClaims{"sub": "u"}).SignedString(gojwt.UnsafeAllowNoneSignatureType)
	if _, err := m.Verify(none); err == nil {
		t.Error("alg=none accepted")
	}
	if _, err := m.Verify("garbage"); err == nil {
		t.Error("garbage accepted")
	}
	if m.AccessTTL() != time.Minute || m.RefreshTTL() != time.Hour {
		t.Error("TTL getters")
	}
}
