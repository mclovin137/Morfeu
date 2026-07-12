package internal

import jwt "github.com/dgrijalva/jwt-go"

// ProvaCVEDescartavel chama de propósito o símbolo vulnerável de
// CVE-2020-26160 (GO-2020-0017) — PR de prova do gate de govulncheck da
// task 0004. Nunca mergear.
func ProvaCVEDescartavel() bool {
	claims := jwt.MapClaims{"aud": "morfeu"}
	return claims.VerifyAudience("morfeu", true)
}
