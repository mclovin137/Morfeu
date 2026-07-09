# Playbook Security — riscos, identificação e correção no Morfeu

> **Como usar**: referência de consulta seletiva do agente `security` (e do `backend-dev` ao
> implementar área sensível). NÃO ler inteiro — o índice mapeia superfície → seção. Cada seção
> traz: onde o risco existe no Morfeu, **como identificar** (sinais em revisão, padrões de grep),
> **como resolver** e **severidade típica** (calibrar com contexto — roles.md §6.6; risco teórico
> sem vetor real é informativo, não bloqueante).
>
> Contexto fixo: Go + Echo, Postgres via sqlc/pgx (SQL estático), Redis, RabbitMQ, Stripe
> **sandbox**, e-mail com QR, backoffice de operador único, dados de clientes (nome, e-mail —
> **LGPD se aplica**). Repo público com secret scanning + push protection.
> Skills relacionadas: `security-review` (checklist geral), `security-scan` (config `.claude/`).

| # | Superfície | Consultar quando… |
|---|-----------|-------------------|
| 1 | [SQL injection](#1-sql-injection) | qualquer SQL fora do sqlc; query dinâmica |
| 2 | [XSS](#2-xss) | HTML/template renderizando dado do usuário |
| 3 | [CSRF](#3-csrf) | rota de escrita autenticada por cookie |
| 4 | [Armazenamento de senha](#4-armazenamento-de-senha) | login/cadastro do backoffice |
| 5 | [Brute force e enumeração](#5-brute-force-e-enumeração) | login, reset de senha, mensagens de erro de auth |
| 6 | [Sessão e cookies](#6-sessão-e-cookies) | emissão/validação de sessão, atributos de cookie |
| 7 | [Autorização e IDOR](#7-autorização-e-idor) | rota com ID de recurso na URL; rotas do backoffice |
| 8 | [Secrets](#8-secrets) | config, env, docker-compose, CI, logs |
| 9 | [Webhook Stripe](#9-webhook-stripe) | endpoint de webhook, verificação de assinatura |
| 10 | [Integridade de pagamento](#10-integridade-de-pagamento) | valor, moeda, status do pedido |
| 11 | [Ingresso e QR](#11-ingresso-e-qr) | geração/validação do QR, check-in |
| 12 | [Validação de entrada / mass assignment](#12-validação-de-entrada--mass-assignment) | bind de request em struct |
| 13 | [Exposição de dados e erros](#13-exposição-de-dados-e-erros) | structs de resposta, handler de erro |
| 14 | [Headers e transporte](#14-headers-e-transporte) | middleware HTTP, TLS, CORS |
| 15 | [SSRF](#15-ssrf) | servidor buscando URL externa (pôster, integração) |
| 16 | [Upload de arquivos](#16-upload-de-arquivos) | upload de pôster no backoffice |
| 17 | [DoS e exaustão de recursos](#17-dos-e-exaustão-de-recursos) | limites de body, timeouts, rate limit |
| 18 | [Dependências e supply chain](#18-dependências-e-supply-chain) | dependência nova/atualizada, CI |
| 19 | [Logs, PII e LGPD](#19-logs-pii-e-lgpd) | qualquer log/trace com dado de cliente |
| 20 | [Timing e comparação de segredos](#20-timing-e-comparação-de-segredos) | comparação de token/assinatura/hash |
| 21 | [Checklist de auditoria](#21-checklist-de-auditoria) | passe de julgamento pré-push (roles.md §6.4) |

---

## 1. SQL injection

**Onde no Morfeu**: a stack praticamente elimina por construção — sqlc gera SQL estático parametrizado. O risco mora nas exceções: query montada na mão com pgx (relatório dinâmico do backoffice, `ORDER BY` vindo do request), `migrate` com input externo.

**Como identificar**
- Grep: `fmt.Sprintf` ou concatenação (`+`) perto de `SELECT|INSERT|UPDATE|DELETE|ORDER BY|WHERE`; chamadas `pool.Query(ctx, variável)` onde a query não é constante.
- Revisão: qualquer identificador (nome de coluna/direção de sort) vindo do request.

**Como resolver**: valores → sempre placeholder (`$1`); identificadores dinâmicos → allowlist fechada no código (`map[string]string{"data": "criado_em", ...}`), nunca o valor do request na string. Filtros opcionais → padrão sqlc `($1 IS NULL OR col = $1)` (playbook-database §17.5).

**Severidade**: crítica se explorável; a existência de SQL concatenado com input é no mínimo alta mesmo sem PoC.

## 2. XSS

**Onde no Morfeu**: depende da entrega do frontend. Templates server-side (`html/template`) escapam por default — o risco é `template.HTML`/`template.JS` (bypass explícito) sobre dado do usuário. Se SPA: o equivalente é `dangerouslySetInnerHTML`/`v-html`. Dados perigosos: título/descrição de filme cadastrado no backoffice (XSS armazenado que dispara no cliente comprador), nome do cliente refletido em página/e-mail.

**Como identificar**
- Grep: `template.HTML(`, `template.JS(`, `template.URL(`, `dangerouslySetInnerHTML`, `innerHTML`.
- E-mails HTML montados por concatenação com nome do cliente.
- Endpoint devolvendo `Content-Type: text/html` com dado não escapado.

**Como resolver**: nunca desativar escaping sobre dado externo; JSON API com `Content-Type: application/json` correto (Echo faz certo por default); CSP (§14) como segunda camada; sanitização de HTML rico só se um dia existir campo rico (bluemonday) — evitar ter o campo.

**Severidade**: alta (armazenado, atinge compradores) / média (refletido).

## 3. CSRF

**Onde no Morfeu**: backoffice autenticado por **cookie de sessão** — rota de escrita (cancelar sessão de cinema, mudar preço) pode ser disparada por página maliciosa no browser do operador logado. Se a autenticação for por header (`Authorization: Bearer`), CSRF não se aplica — o browser não anexa headers automaticamente.

**Como identificar**: rotas mutantes (POST/PUT/DELETE) autenticadas via cookie sem token CSRF nem `SameSite` restritivo; `GET` que muda estado (duplamente errado).

**Como resolver**
1. `SameSite=Lax` (ou `Strict`) no cookie de sessão — elimina o vetor clássico em browsers modernos; **base obrigatória** (§6).
2. Middleware CSRF do Echo (`middleware.CSRF`) com token por sessão — defesa explícita para o backoffice; adotar se houver qualquer dúvida sobre o SameSite cobrir (subdomínios, iframes).
3. Nenhuma mutação via GET, nunca.

**Severidade**: alta no backoffice (ações administrativas), com SameSite ausente.

## 4. Armazenamento de senha

**Onde no Morfeu**: conta do operador do backoffice (e contas de cliente, se houver login de comprador).

**Como identificar**
- Grep: `md5|sha1|sha256` perto de `password|senha`; qualquer senha em SELECT com comparação direta; coluna `senha` sem sufixo `_hash`.
- Senha logada em qualquer nível (§19); senha em resposta de API.

**Como resolver**: **argon2id** (`golang.org/x/crypto/argon2`, parâmetros OWASP: 19 MiB, t=2, p=1 como piso) ou bcrypt (custo ≥ 12) — registrar a escolha em `lib.md`; salt único por senha (ambos embutem); hash com formato versionado (`$argon2id$...`) para rehash futuro; validação com comparação da própria lib (constant-time embutido — §20). Política: comprimento mínimo 8+, sem regras de composição bizantinas, checar contra senhas vazadas se barato.

**Severidade**: crítica (hash fraco/ausente = comprometimento total na primeira exfiltração).

## 5. Brute force e enumeração

**Onde no Morfeu**: login do backoffice é alvo direto (painel administrativo exposto na internet); reset de senha; e — específico do domínio — **enumeração de pedidos/ingressos** por ID (coberto em §7/§11).

**Como identificar**
- Login sem rate limit por conta+IP (grep pela rota de login vs middleware de rate limit).
- Mensagens distintas: "usuário não existe" vs "senha incorreta" — enumeração de contas. Idem em reset ("e-mail enviado" vs "e-mail não cadastrado") e no tempo de resposta (hash só roda se o usuário existe → timing revela — §20).

**Como resolver**: rate limit duro no login (ex.: 5/min por conta+IP, resposta 429 — playbook-backend §17); mensagem única "credenciais inválidas"; no reset, sempre "se existir, enviamos"; rodar o hash mesmo para usuário inexistente (hash dummy) para igualar o tempo; lockout progressivo/backoff em vez de bloqueio permanente (que vira DoS contra o operador); log + métrica de tentativas falhas (roles.md §6.8) para detectar campanha.

**Severidade**: alta (login backoffice sem rate limit).

## 6. Sessão e cookies

**Onde no Morfeu**: sessão do operador do backoffice; sessão/carrinho do comprador se autenticado.

**Como identificar**: grep `SetCookie|http.Cookie` e conferir atributos; token de sessão previsível ou curto; sessão sem expiração; ID de sessão que não muda após login.

**Como resolver**
1. Cookie de sessão: `HttpOnly` (JS não lê — mitiga roubo via XSS), `Secure` (só HTTPS), `SameSite=Lax` mínimo (§3), `Path` restrito, sem `Domain` amplo.
2. Token de sessão: 128+ bits de `crypto/rand` (nunca `math/rand` — grep por `math/rand` em código de auth é achado imediato), armazenado server-side (Redis/Postgres) com expiração absoluta + deslizante.
3. **Rotacionar o ID de sessão no login** (session fixation) e invalidar server-side no logout.
4. JWT: só se surgir necessidade real (não há múltiplos serviços); se usar — assinatura verificada com algoritmo fixo (nunca aceitar `alg` do token, rejeitar `none`), expiração curta, e aceitar que revogação exige lista server-side (o que devolve ao ponto 2 — por isso sessão server-side é o default do projeto).

**Severidade**: alta (cookie sem HttpOnly/Secure em produção; token previsível = crítica).

## 7. Autorização e IDOR

**Onde no Morfeu**: o clássico do domínio — `GET /pedidos/{id}` ou `/ingressos/{id}` devolvendo pedido **de outro cliente** porque só validou autenticação, não posse. E a divisão de privilégio: rotas de backoffice acessíveis sem papel de operador.

**Como identificar**
- Toda rota com ID de recurso: a query filtra por dono (`WHERE id = $1 AND usuario_id = $2`) ou só por ID? Grep nas queries sqlc por `WHERE id =` sem cláusula de posse em recursos possuídos.
- Rotas de backoffice: middleware de papel presente no **grupo** de rotas (não handler a handler, onde um esquecimento passa)?
- IDs sequenciais expostos amplificam (enumeráveis — playbook-database §11).

**Como resolver**
1. **Posse na query, não no código**: buscar-por-id-e-dono numa query só; "não achou" = 404 (não 403, que confirma existência).
2. Autorização por **grupo de rotas** no Echo (middleware no `e.Group("/admin")`), deny-by-default; rota nova nasce protegida sem lembrar de nada.
3. UUIDs em URLs públicas (defesa em profundidade, não substituto do check de posse).
4. Teste do QA: usuário A tenta recurso de B → 404; anônimo tenta backoffice → 401/302. Exigir esse teste em todo PRD com recurso possuído.

**Severidade**: crítica (IDOR em pedido/ingresso = dados pessoais + ingresso de terceiro; backoffice aberto = comprometimento total).

## 8. Secrets

**Onde no Morfeu**: chaves Stripe (sandbox hoje, real um dia), credencial SMTP, DSN do Postgres, senha do Redis, assinatura de sessão/QR. Repo é **público** — um vazamento é imediato e permanente (histórico).

**Como identificar**
- Grep no diff e na árvore: `sk_live|sk_test|whsec_|password=|passwd|secret|api_key|BEGIN.*PRIVATE KEY|AKIA`; strings de alta entropia em código/config commitado.
- `docker-compose.yml` com senha real em vez de referência a env; `.env` fora do `.gitignore`; secret em log (§19), em URL (vai para log de acesso), em mensagem de erro.
- Histórico também conta: secret removido em commit posterior **continua exposto** — exige rotação, não só remoção.

**Como resolver**: env vars carregadas na borda (config struct no boot, falha rápida se ausente); `.env` local ignorado + `.env.example` sem valores; secrets de CI no cofre do GitHub Actions; **vazou → rotacionar imediatamente** (a chave, não o commit); secret scanning + push protection já ativos no repo — a auditoria pré-push é a camada humana disso (roles.md §6.4: secret exposto = REPROVADO sem exceção).

**Severidade**: crítica, sempre. Sandbox vaza também: normaliza o hábito e o mesmo caminho carrega a chave real depois.

## 9. Webhook Stripe

**Onde no Morfeu**: o endpoint de webhook é quem **confirma pagamento** — se aceitar evento forjado, um atacante "paga" pedidos sem pagar. É a rota mais sensível do sistema junto com o login.

**Como identificar**
- Handler do webhook sem `webhook.ConstructEvent` (stripe-go) / verificação de `Stripe-Signature`.
- Segredo do webhook (`whsec_`) hardcoded (§8).
- Corpo lido **depois** de passar por middleware que o consome/transforma (a verificação exige o corpo bruto, byte a byte).
- Lógica confiando em campos do evento sem revalidar contra o pedido local (§10).

**Como resolver**
1. `webhook.ConstructEvent(payloadBruto, header, whsec)` da lib oficial — verifica HMAC em tempo constante e a janela de tolerância de timestamp (anti-replay). Nunca implementar a verificação na mão.
2. Rota fora de qualquer middleware de parse de body; ler `io.ReadAll` com limite (§17) antes de tudo.
3. Idempotência por `event.ID` (playbook-backend §1) — replay legítimo do Stripe é esperado.
4. Processar só os tipos de evento esperados; ignorar (200) os demais sem efeito.
5. Responder 2xx rápido e processar efeitos pesados via outbox — timeout do Stripe gera retry e duplicação de trabalho.

**Severidade**: crítica (assinatura não verificada = ingressos grátis).

## 10. Integridade de pagamento

**Onde no Morfeu**: preço do ingresso, total do pedido, status. Regra: **o cliente escolhe assentos; o servidor decide preço**.

**Como identificar**
- Qualquer campo `valor|preco|total|amount` no payload de request de checkout que o servidor **usa** em vez de recalcular.
- Criação do PaymentIntent com amount vindo do request.
- Webhook marcando pedido como pago sem conferir `amount_received`/`currency` contra o pedido local.
- Transição de status sem máquina de estados (pedido `expirado` que vira `pago` por webhook atrasado — race legítima que precisa de resposta definida).

**Como resolver**: preço sempre do banco no momento do checkout, congelado no pedido (snapshot em `pedido_itens.valor_centavos` — preço pode mudar no backoffice depois); PaymentIntent criado server-side com o valor do pedido e `pedido_id` nos metadata; webhook valida evento↔pedido (id, valor, moeda) antes de transicionar; transições válidas explícitas (playbook-backend §2) — update condicional `WHERE status = 'aguardando_pagamento'` (playbook-database §7.3); divergência → não confirmar, logar como incidente, alertar.

**Severidade**: crítica.

## 11. Ingresso e QR

**Onde no Morfeu**: o QR **é** o ingresso — quem apresenta um QR válido entra. Ameaças: forjar, adivinhar, reusar (dupla entrada com screenshot compartilhado).

**Como identificar**: conteúdo do QR é um ID sequencial ou dado adivinhável? Validação no check-in consulta o banco ou "confia" no conteúdo? Existe marcação de uso?

**Como resolver**
1. Conteúdo do QR: **referência opaca** — UUID aleatório do ingresso ou token de 128+ bits `crypto/rand`. Alternativa: token assinado (HMAC) validável offline — só se o check-in precisar funcionar sem rede; caso contrário a referência opaca + consulta é mais simples e revogável.
2. Check-in valida no servidor: ingresso existe, sessão correta, **ainda não usado** — e marca como usado **atomicamente** (`UPDATE ... SET usado_em = now() WHERE id = $1 AND usado_em IS NULL`, RowsAffected = 1; dois porteiros escaneando juntos = um entra, playbook-database §7.3).
3. Endpoint de validação autenticado (é rota do operador) e com rate limit — é um oráculo de adivinhação se aberto.
4. Reuso detectado → mensagem clara ao operador ("já utilizado às 21h03"), não erro genérico.

**Severidade**: alta (QR adivinhável/reusável = entrada grátis; com IDOR §7 vira crítica).

## 12. Validação de entrada / mass assignment

**Onde no Morfeu**: handlers Echo fazendo `c.Bind(&struct)` — o bind preenche **qualquer** campo do struct presente no JSON. Se o struct de request for a entidade do domínio, o cliente escreve `status`, `valor_centavos`, `usado_em`.

**Como identificar**
- Grep: `c.Bind(` recebendo struct que também é modelo de banco/domínio (cruzar com os tipos do sqlc/domínio).
- Campos sem validação de faixa/formato após o bind (quantidade de assentos negativa, e-mail malformado, string de 10 MB num campo de nome).

**Como resolver**: **structs de request dedicados por endpoint** (DTO) contendo só os campos que o cliente pode enviar — mass assignment morre por construção (alinhado às fronteiras do ADR 0003); validação declarativa (`validate` tags + validator registrado no Echo) ou explícita logo após o bind: presença, faixa, comprimento, formato; rejeitar cedo com 400 e mensagem por campo; limites de tamanho de body antes do parse (§17).

**Severidade**: alta quando o bind alcança campo sensível; média como higiene geral.

## 13. Exposição de dados e erros

**Onde no Morfeu**: struct do domínio serializado direto na resposta (vaza `senha_hash` do operador, e-mail de outro cliente, campos internos); erro de banco cru no JSON (`pq: duplicate key ... usuarios_email_key` — confirma e-mail cadastrado + revela schema); stack trace em produção.

**Como identificar**
- Grep: `c.JSON(` retornando tipo do domínio/sqlc em vez de struct de resposta; `err.Error()` dentro de resposta HTTP.
- Modo debug do Echo ligado fora de dev; `pprof`/`expvar` exposto sem auth.

**Como resolver**: DTOs de resposta explícitos (espelho do §12 — nada sai sem struct de saída dedicado); handler de erro central do Echo: erro de domínio → status + mensagem controlada; erro inesperado → 500 genérico + log completo server-side com correlation id (o id vai na resposta para suporte, o detalhe não); endpoints de diagnóstico (`/metrics`, `pprof`) em porta/grupo interno.

**Severidade**: média a alta (depende do que vaza; hash de senha em resposta = crítica).

## 14. Headers e transporte

**Onde no Morfeu**: app atrás do TLS da plataforma (free tier termina TLS); backoffice e checkout são páginas sensíveis.

**Como identificar**: ausência de middleware de secure headers; CORS `AllowOrigins: ["*"]` com credenciais; cookie sem `Secure`; app aceitando HTTP puro em produção.

**Como resolver**
1. `middleware.Secure()` do Echo como base + ajustes: `Strict-Transport-Security` (com a plataforma servindo HTTPS), `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` (ou CSP `frame-ancestors`), `Referrer-Policy: strict-origin-when-cross-origin`.
2. CSP para as páginas do backoffice/checkout: `default-src 'self'` + exceções mínimas do Stripe (js.stripe.com etc. conforme docs oficiais — Context7). Camada 2 contra XSS (§2).
3. CORS: origem exata do frontend, nunca `*` com `AllowCredentials`.
4. Confiar no `X-Forwarded-For` **só** do proxy da plataforma (config de IP extractor do Echo) — senão o rate limit por IP (§5) é contornável por header forjado.

**Severidade**: média (defesa em profundidade); CORS aberto com credenciais = alta.

## 15. SSRF

**Onde no Morfeu**: só existe se o servidor buscar URLs — ex.: backoffice informa URL de pôster e o servidor baixa. O atacante (operador comprometido ou validação fraca) aponta para `http://169.254.169.254/` (metadata da cloud) ou serviços internos (`redis:6379`).

**Como identificar**: grep `http.Get(|http.NewRequest(` com URL derivada de input; qualquer feature "buscar de URL".

**Como resolver**: preferir **upload** (§16) a fetch-por-URL — elimina a classe; se fetch for inevitável: allowlist de esquema (https) e host, resolver DNS e bloquear IPs privados/link-local **na conexão** (custom `DialContext` — protege contra DNS rebinding), sem seguir redirects cegamente, timeout curto e limite de tamanho.

**Severidade**: alta se existir o vetor; hoje é informativo (não há fetch de URL).

## 16. Upload de arquivos

**Onde no Morfeu**: pôster do filme no backoffice (se o PRD optar por upload).

**Como identificar**: handler de upload confiando em `Content-Type` do request ou na extensão; arquivo salvo com nome vindo do cliente (path traversal `../../`); sem limite de tamanho; servido do mesmo host com Content-Type permissivo (SVG com script = XSS §2).

**Como resolver**: validar **magic bytes** (`http.DetectContentType` nos primeiros 512 bytes) contra allowlist curta (jpeg, png, webp — **não** SVG); nome novo gerado (UUID + extensão derivada do tipo real), nunca o original; `MaxBytesReader`/limite no Echo antes de ler; armazenar fora da árvore servida ou com `Content-Type` explícito + `nosniff`; dimensões máximas se houver processamento de imagem (decompression bomb).

**Severidade**: alta (upload sem validação em área autenticada; path traversal = crítica).

## 17. DoS e exaustão de recursos

**Onde no Morfeu**: free tier tem pouquíssima folga — não é preciso botnet, um script mal-educado derruba. Vetores: body gigante, conexões lentas (slowloris), rota cara martelada (busca, mapa de assentos), checkout de bot segurando travas de assento.

**Como identificar**: servidor HTTP sem `ReadTimeout`/`WriteTimeout`/`IdleTimeout`; ausência de `BodyLimit` no Echo; rota cara sem rate limit; trava de assento sem TTL (bot reserva o cinema inteiro para sempre).

**Como resolver**: `middleware.BodyLimit` global (ex.: 1 MB) com exceção dimensionada só no upload; timeouts do `http.Server` configurados (Go não tem default!); rate limiting nos públicos (playbook-backend §17); **TTL curto na trava de assento** (ex.: 10 min) com liberação automática — a defesa de domínio mais importante contra bot de reserva; back pressure interno (playbook-backend §7). Não prometer proteção contra DDoS volumétrico — isso é camada de plataforma/CDN, registrar como limitação aceita.

**Severidade**: média (hardening); trava sem TTL = alta (nega o negócio inteiro).

## 18. Dependências e supply chain

**Onde no Morfeu**: cada `go get` novo (roles.md §6.9: justificar + CVE + `lib.md` **antes** de usar).

**Como identificar**
- **`govulncheck ./...`** — específico de Go, acusa só vulnerabilidade em **função alcançável** (menos falso positivo que scanners de manifesto); rodar na revisão de dependência e no CI (E0c).
- Checagens de sanidade da lib: manutenção ativa, popularidade, autor identificável, typosquatting no path do módulo (`github.com/sirupsen/logrus` vs imitações).
- `go.sum` sempre commitado (integridade criptográfica); diff de `go.mod` em PR sem entrada correspondente no `lib.md` = achado imediato.

**Como resolver**: vulnerabilidade alcançável → atualizar para versão corrigida; sem correção → avaliar substituição ou mitigação documentada com prazo; manter dependências no menor conjunto (stdlib primeiro — cultura do ADR 0004); versões atualizadas em cadência (não deixar apodrecer 2 anos e migrar no desespero); CVEs sempre por fontes atuais (Context7/advisories — regra dura do agente, nunca memória de treinamento).

**Severidade**: conforme o CVE e alcançabilidade; processo violado (dep sem registro/CVE check) = REPROVADO por regra, independente de severidade técnica.

## 19. Logs, PII e LGPD

**Onde no Morfeu**: clientes fornecem nome e e-mail (PII sob LGPD). Logs estruturados (slog) + traces OTel podem vazar isso para plataformas de log de terceiros sem contrato/controle.

**Como identificar**
- Grep: `slog.` / atributos de log com `email|nome|senha|token|cookie|authorization|card`.
- Log de request body inteiro (middleware de dump); headers logados sem redação; DSN com senha em log de boot; token em query string (vai para access log de qualquer proxy).
- Dados de cartão: **não podem existir** em log nem banco — com Stripe Elements/Checkout o número nunca toca o servidor; qualquer campo de cartão em request próprio é achado crítico + problema de escopo PCI.

**Como resolver**: allowlist mental de log — IDs opacos sim, conteúdo não (`pedido_id`, `usuario_id` em vez de e-mail); redação central (implementar `slog.LogValuer` nos tipos com PII — o tipo se auto-redige em **qualquer** log, ninguém precisa lembrar); tokens/segredos jamais em URL; retenção curta e documentada; e-mail do cliente é dado de negócio (banco, protegido por §7), não dado de log.

**Severidade**: alta (PII em log de terceiro = incidente LGPD); credencial/token em log = crítica (regra dura: REPROVADO).

## 20. Timing e comparação de segredos

**Onde no Morfeu**: comparação de token de sessão/QR/API key com `==` ou `bytes.Equal` vaza informação por tempo de resposta (byte a byte). Hash de login que só roda para usuário existente (§5).

**Como identificar**: grep `== ` / `bytes.Equal|strings.EqualFold` em código que compara token, assinatura, hash, chave; branch de login que retorna antes de custo constante.

**Como resolver**: `crypto/subtle.ConstantTimeCompare` para comparação direta de segredos; melhor ainda — comparar **hashes** dos tokens (armazenar SHA-256 do token de sessão no banco: lookup por hash é seguro e o banco nunca guarda o token vivo); bibliotecas oficiais já fazem certo (stripe-go §9, argon2/bcrypt §4) — não reimplementar; para o timing de login, hash dummy no caminho "usuário não existe".

**Severidade**: média isolada (explorar timing remoto é difícil), mas é higiene barata; token vivo armazenado sem hash sobe para alta.

## 21. Checklist de auditoria

Passe de julgamento de segurança (roles.md §6.4) — percorrer o que o diff toca:

1. **Secrets**: diff e histórico limpos? Env/config sem valor real? (§8 — REPROVADO se violado)
2. **Auth**: rota nova exige autenticação correta? Grupo protegido? (§6, §7)
3. **Posse**: recurso possuído filtrado por dono na query? (§7)
4. **Entrada**: DTO dedicado + validação + BodyLimit? (§12, §17)
5. **Saída**: DTO de resposta, erro genérico, nada interno vazando? (§13)
6. **SQL**: algum SQL fora do sqlc? Concatenação? (§1)
7. **Dinheiro**: valor sempre server-side? Webhook verificado? (§9, §10)
8. **Logs**: PII/token em algum log novo? (§19 — REPROVADO se credencial)
9. **Dependência nova**: `lib.md` + CVE (govulncheck)? (§18 — REPROVADO se ausente)
10. **Segredos comparados**: constant-time / hash armazenado? (§20)

Classificar cada achado (crítica/alta/média/baixa) com vetor concreto; teórico sem vetor = informativo (regra do agente: não inflar achados).
