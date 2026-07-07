# Autenticacao, JWT e criptografia

Bibliotecas registradas:

- `golang-jwt/jwt/v5 >= 5.2.2`
- `labstack/echo-jwt`
- `golang.org/x/crypto >= 0.52.0`, com uso de Argon2id

## Papel no Morfeu

JWT sera usado para tokens de acesso. Senhas devem ser armazenadas somente como
hash Argon2id com parametros versionados. A API deve separar autenticacao
(`quem e o usuario`) de autorizacao (`o que ele pode fazer`).

## JWT

Claims minimos recomendados:

```json
{
  "sub": "user_id",
  "typ": "access",
  "iat": 1720000000,
  "exp": 1720000900,
  "roles": ["customer"]
}
```

Regras:

- usar expiracao curta em access token;
- validar algoritmo explicitamente;
- nao aceitar `alg=none`;
- separar segredo/chave por ambiente;
- nao guardar dados sensiveis no payload;
- rotacionar chaves com `kid` se o projeto evoluir para multiplas chaves.

Exemplo conceitual:

```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	"sub": userID.String(),
	"typ": "access",
	"iat": time.Now().Unix(),
	"exp": time.Now().Add(15 * time.Minute).Unix(),
})

signed, err := token.SignedString(secret)
```

## Echo JWT

Aplicar middleware em grupo privado, nao globalmente:

```go
private := v1.Group("")
private.Use(echojwt.WithConfig(echojwt.Config{
	SigningKey: []byte(cfg.JWTSecret),
	NewClaimsFunc: func(c echo.Context) jwt.Claims {
		return new(AppClaims)
	},
}))
```

Depois do middleware, extrair claims e converter para identidade interna. Nao
espalhar dependencia direta de Echo ou JWT pelos services.

## Argon2id

Usar Argon2id para senha com parametros configurados e persistidos junto do hash
em formato parseavel:

```text
$argon2id$v=19$m=65536,t=3,p=2$<salt_base64>$<hash_base64>
```

Regras:

- salt aleatorio por senha;
- comparar hash com tempo constante;
- versionar parametros para permitir rehash futuro;
- limitar tamanho maximo de senha recebida para evitar abuso de CPU/memoria;
- nao logar senha, hash ou salt.

Exemplo:

```go
hash := argon2.IDKey(password, salt, 3, 64*1024, 2, 32)
if subtle.ConstantTimeCompare(hash, expected) != 1 {
	return ErrInvalidCredentials
}
```

## Fluxo recomendado

1. Handler valida JSON e tamanho dos campos.
2. Service busca usuario por email.
3. Hash Argon2id e comparado em tempo constante.
4. Service emite access token.
5. Handler retorna token e dados publicos.

Para login, retornar erro generico para email inexistente ou senha invalida.

