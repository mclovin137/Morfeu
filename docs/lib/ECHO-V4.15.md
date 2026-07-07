# Echo v4.15+

Context7:

- Resolvido como `/labstack/echo`
- Documentacao consultada em `/labstack/echo/v4.15.0`
- Topicos usados: grupos de rotas, middleware, `DefaultBinder`, `Validator`,
  `DefaultHTTPErrorHandler`, `StartConfig` e shutdown gracioso.

Registro local: [`lib.md`](../../lib.md) exige `labstack/echo/v4 >= 4.15.0`.

## Papel no Morfeu

Echo sera o framework HTTP da API. Ele deve ficar restrito a camada de entrega:
roteamento, middleware, parsing de request, serializacao de response e traducao
de erros. Regras de negocio ficam em services/use cases.

## Bootstrap recomendado

```go
func NewServer(deps Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = NewHTTPErrorHandler(deps.Logger)
	e.Validator = NewValidator()

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(requestLoggerConfig(deps.Logger)))
	e.Use(otelecho.Middleware("morfeu-api"))

	api := e.Group("/api")
	v1 := api.Group("/v1")

	RegisterHealthRoutes(v1, deps)
	RegisterAuthRoutes(v1, deps)
	RegisterCinemaRoutes(v1, deps)

	return e
}
```

Ordem de middleware importa. Recomendada:

1. request ID/correlacao;
2. recover;
3. logging;
4. tracing;
5. rate limit/CORS quando aplicavel;
6. autenticacao em grupos privados;
7. autorizacao por rota/caso de uso.

## Grupos de rota

Context7 confirmou que `Group(prefix, middleware...)` cria grupos hierarquicos
e subgrupos herdam middleware do grupo pai.

```go
api := e.Group("/api")
v1 := api.Group("/v1")

public := v1.Group("")
private := v1.Group("", echojwt.WithConfig(jwtConfig))

public.POST("/sessions", authHandler.CreateSession)
private.GET("/me", accountHandler.Me)
```

Evitar registrar handlers diretamente no `Echo` raiz exceto healthcheck ou
diagnosticos intencionais.

## Binding e validacao

O `DefaultBinder` do Echo aplica binding em ordem:

1. path params;
2. query params para `GET`, `DELETE` e `HEAD`;
3. body.

Para evitar sobrescrita acidental, DTOs devem separar entrada de path, query e
body quando houver risco de colisao.

```go
type CreateSessionRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
}

func (h *AuthHandler) CreateSession(c echo.Context) error {
	var req CreateSessionRequest
	if err := c.Bind(&req); err != nil {
		return NewBadRequest("invalid_json", err)
	}
	if err := c.Validate(req); err != nil {
		return NewValidationError(err)
	}

	out, err := h.auth.CreateSession(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}
```

Nao passar `echo.Context` para services, repositorios ou clients. Passar apenas
`c.Request().Context()` e dados ja validados.

## Erros HTTP

Context7 confirmou que Echo centraliza tratamento por `HTTPErrorHandler` e evita
responder se a response ja foi committed.

Contrato recomendado:

```json
{
  "error": {
    "code": "seat_unavailable",
    "message": "assento indisponivel",
    "request_id": "..."
  }
}
```

Nao expor `err.Error()` de falhas internas em producao. Mapear erros conhecidos:

- validacao: `400`;
- credencial invalida: `401`;
- sem permissao: `403`;
- inexistente: `404`;
- concorrencia/assento indisponivel: `409`;
- rate limit: `429`;
- inesperado: `500`.

## Shutdown

Context7 confirmou suporte a shutdown gracioso com timeout. Usar timeout curto e
explicito:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

go func() {
	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("http server failed", "error", err)
	}
}()

<-ctx.Done()

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := e.Shutdown(shutdownCtx); err != nil {
	logger.Error("http shutdown failed", "error", err)
}
```

Workers e conexoes externas devem ser fechados no mesmo fluxo de encerramento.

