# ADR 0004 — Padrões de código Go

- **Status:** aceito
- **Data:** 2026-07-07
- **Task/PRD relacionados:** confronto de padrões de código (2026-07-07); `.golangci.yml` nasce no E0

## Contexto

O usuário propôs **Object Calisthenics** como disciplina de código. A análise dos agentes concluiu: as 9 regras nasceram para Java/OO — parte já É o idioma Go, parte precisa de adaptação e duas (na forma literal) produzem anti-Go (wrappers sobre código gerado do sqlc; brigar com injeção de dependências por construtor). O usuário aceitou no confronto o subconjunto adaptado com enforcement por lint. Este ADR também fixa os padrões transversais que faltavam na proposta (erros, transação, DTOs, naming, context, concorrência).

## Escopo

Cobre: estilo e disciplina de código Go, tratamento de erros, transação, validação, DTOs, nomenclatura/idioma, context, concorrência, tooling de enforcement. Não cobre: estrutura de camadas (ADR 0003), DDD e patterns (ADR 0005), testes (ADR 0006).

## Decisão

### Object Calisthenics adaptado (regra a regra)

| # | Regra original | Veredito | Forma no Morfeu | Enforcement |
|---|---|---|---|---|
| 1 | Um nível de indentação | Adaptada | máx. ~2 níveis; código feliz na margem esquerda; exceção: table-driven tests | `gocognit`, `nestif` |
| 2 | Não usar ELSE | Adotada | early return/guard clause — já é o idioma | `revive` (early-return, superfluous-else, indent-error-flow) |
| 3 | Encapsular primitivos | Adaptada | **só onde há invariante**: `Centavos`, `StatusPedido`, `StatusHold`, `AssentoCodigo` (ADR 0005); o resto usa tipos nus/gerados | convenção + `exhaustive` p/ enums |
| 4 | Coleções de 1ª classe | Adaptada | só onde há regra: `MapaDeAssentos` (valida contra layout, máx. 6); slice sem comportamento fica slice | convenção |
| 5 | Um ponto por linha | Diretriz | não atravessar structs (`a.B.C.D`); sem lint (falso-positivo) | revisão/auditoria |
| 6 | Não abreviar | Invertida parcialmente | convenção Go: nome proporcional ao escopo (`ctx`, `tx`, `q`, `i` ok em escopo curto); exportados e domínio nunca abreviados | revisão; sem `varnamelen` |
| 7 | Entidades pequenas | Adotada | funções ≤ ~80 linhas, complexidade cognitiva ≤ ~15 | `funlen`, `gocognit` |
| 8 | Máx. 2 campos por struct | **Rejeitada** | quebraria structs de config e injeção de 3–5 dependências; espírito mantido: struct 10+ campos de negócio = cheiro de revisão | revisão |
| 9 | Sem getters/setters | Não-questão | corolário adotado: nunca `GetX()` (acessor Go é `X()`); aggregates críticos com campos não exportados + `NewX` validando invariantes | convenção |

### Padrões transversais

1. **Erros**: erros de domínio por módulo em `errors.go` (sentinelas/tipos com dados), cada um com **código estável** (base do i18n, `doc.md` §7); wrap com `%w` adicionando contexto em toda fronteira; comparação só com `errors.Is/As`; **um `HTTPErrorHandler` central** no Echo mapeia domínio→status+código JSON (o 409 da disputa nasce aqui); erro logado **uma vez, na borda** (proibido log-and-return); erro interno nunca vaza na resposta; panic nunca para fluxo (recover middleware).
2. **Transação**: o **service (ou orquestrador) é o único dono da TX** via `plataforma.WithTx(ctx, pool, fn)` + `queries.WithTx(tx)`; handler nunca vê `tx`. Regra de ouro: outbox e dedup de webhook **na mesma TX do efeito de negócio**.
3. **Validação em duas linhas**: forma/tipo no handler (bind + validação de DTO — lib decidida no refinamento do E0, registrada em `lib.md` antes de usar); invariantes de negócio no domínio (`NewPedido` valida máx. 6 mesmo se o handler falhar).
4. **DTOs**: structs de request/response com tags JSON vivem só no handler; struct de domínio e struct gerada pelo sqlc **nunca** serializam para HTTP; mapeamento à mão (5 linhas visíveis, sem automapper); nos módulos transaction-script, sqlc row → DTO direto é permitido.
5. **Naming/idioma** (decisão do confronto): **domínio em português sem acentos** (`Pedido`, `Ingresso`, `Sessao`, `HoldAssento` — coerente com `doc.md`, métricas e tabelas), **técnico em inglês** (`handler`, `service`, `ctx`, `NewX`, `WithTx`); tabelas/colunas em PT; glossário no README para leitores externos. A regra ruim é o misto sem critério — este É o critério.
6. **Context**: `ctx context.Context` primeiro parâmetro de toda função pública com I/O; timeout explícito em toda chamada externa; ctx carrega trace/deadline, nunca dado de negócio; contexto de trace propagado na mensagem RabbitMQ.
7. **Concorrência**: goroutines de longa duração só em `plataforma`/workers (sweeper, relay, consumidor), com dono e shutdown por cancelamento de ctx; **código de domínio é 100% síncrono**; `go test -race` como gate.
8. **Tempo injetado**: funções de domínio sensíveis a tempo (expiração de hold, janela de 2h) recebem `now time.Time`; interface `Clock` só se a necessidade crescer.
9. **Duas convenções anti-Java-em-Go**: "aceite interfaces, retorne structs"; graceful shutdown padronizado em `plataforma` (para de aceitar → drena → fecha).
10. **Tooling**: `gofumpt` + `goimports` obrigatórios; `.golangci.yml` **versionado** com: base (`errcheck`, `govet`, `staticcheck`, `revive`) + regras acima + `errorlint`, `exhaustive`, `gocritic`, `gosec`, `sqlclosecheck`, `rowserrcheck`, `misspell`, `depguard` (ADR 0003); `sqlc vet` e `govulncheck` no CI; Makefile com `make lint test gen`. Nomes/formato da config (v2) validados no Context7 na task do E0 — não presumidos daqui.

## Tecnologias ou padrões envolvidos

golangci-lint (config v2), gofumpt, goimports, govulncheck, convenções go.dev/Google Go Style Guide.

## Benefícios

- O que é automatizável roda em todo PR — padrão que não vira lint vira item de auditoria, nunca opinião solta em review.
- Cada regra tem justificativa própria e verificável (a auditoria §6.4 consegue cobrar).
- Códigos de erro estáveis pagam o scaffolding de i18n prometido na descoberta.
- Dono único de TX + outbox na mesma TX tornam a regra central da saga verificável em revisão.

## Trade-offs

- Perde-se o rótulo "seguimos as 9 regras de Object Calisthenics" — adota-se o espírito, não a letra.
- Lint estrito gera atrito nos primeiros PRs (rampa de aprendizado).
- PT no domínio: recrutador internacional lê o código com glossário — trade-off aceito no confronto (alternativa EN era defensável).

## Riscos

- Lint desabilitado pontualmente "para passar" → `//nolint` exige justificativa na linha e vira item de auditoria.
- Deriva do padrão com o tempo → o `.golangci.yml` é a fonte executável; mudanças nele exigem atualizar este ADR (supersede parcial).

## Estratégias para minimizar os trade-offs

- Atrito inicial → thresholds calibrados (~15/~80/~4) em vez de valores punitivos; ajustáveis por PR com justificativa.
- Rótulo calisthenics → a tabela deste ADR documenta a relação regra-a-regra para leitores que conheçam o original.

## Impacto esperado

`.golangci.yml`, Makefile e `HTTPErrorHandler` nascem no E0; todos os PRs passam pelo mesmo funil; auditoria ganha itens objetivos (nolint justificado, log único na borda, dono da TX).

## Alternativas consideradas e descartadas

- **Object Calisthenics integral (9 regras literais)** — descartada: regras 3 e 8 produzem wrappers sobre o sqlc e brigam com DI por construtor (+1–2 semanas de atrito); regras 5/6/8 sem critério verificável em Go viram opinião em review. Confrontada e recusada pelo usuário.
- **Sem disciplina formal (só gofmt)** — descartada: dev solo aprendendo a linguagem precisa de guarda-corpos executáveis; é o cenário exato onde o hábito JVM vaza.
- **EN integral no código** — descartada no confronto: criaria dicionário permanente código↔doc.md; PT no domínio + EN técnico venceu.

## ADRs relacionados

- ADR 0003 (camadas) — as regras daqui valem dentro daquela estrutura.
- ADR 0005 (DDD tático) — define os value objects citados na regra 3.
- ADR 0006 (testes) — `-race`, table-driven e tempo injetado se conectam às regras 7/8 daqui.
