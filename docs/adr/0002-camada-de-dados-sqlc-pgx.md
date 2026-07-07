# ADR 0002 — Camada de dados: sqlc + pgx/v5

- **Status:** aceito
- **Data:** 2026-07-07
- **Task/PRD relacionados:** descoberta (`doc.md` §4/§9), roadmap E0

## Contexto

O domínio exige recursos avançados de PostgreSQL que são o coração dos fluxos críticos: partial unique index para a trava de assento, `INSERT ... ON CONFLICT DO NOTHING` atômico, `UPDATE ... WHERE status = $2` com guarda (compare-and-swap da saga), transações multi-tabela (hold→vendido + ingresso + outbox). O projeto segue migration-first (roles.md §6.10, schema versionado em SQL via golang-migrate). Era preciso decidir como o Go conversa com o banco antes do primeiro módulo.

## Escopo

Cobre: driver, geração de código de acesso, gestão de pool e transação. Não cobre: modelagem de schema (migrations, §6.10), fronteiras de quem chama o quê (ADR 0003).

## Decisão

**sqlc gerando código type-safe sobre pgx/v5, com pool via `pgxpool`.** SQL explícito em arquivos `queries.sql` por módulo; o código gerado (`db/`) não é editado à mão. pgx puro é permitido para SQL dinâmico (filtros do backoffice). Transações via helper único `plataforma.WithTx` (pgx `Begin`/`defer Rollback`/`Commit`, padrão validado no Context7) + `Queries.WithTx(tx)` do sqlc. Instrumentação OTel via `otelpgx`. `overrides` do sqlc usados com parcimônia para mapear colunas a value objects com invariante (ADR 0005).

## Tecnologias ou padrões envolvidos

sqlc (codegen, dev tool), pgx/v5 (**≥ 5.9.2** — CVE-2026-33815/33816 corrigidas em 5.9.0, CVE-2026-41889 em 5.9.2), pgxpool, otelpgx, `sqlc vet` no CI.

## Benefícios

- SQL explícito → o dev aprende PostgreSQL de verdade (objetivo do projeto); as queries críticas (ON CONFLICT, guardas) ficam visíveis e revisáveis no PR.
- Type-safety em compile time sem custo de runtime (sem reflection).
- `WithTx` encaixa naturalmente na saga: o service compõe queries de módulos na mesma transação.
- O SQL é a lógica: testes de integração contra PG real (testcontainers) validam o que roda em produção — repositório fake perde utilidade (ADR 0006).
- Coexistência com pgx puro dá saída para SQL dinâmico sem quebrar o padrão.

## Trade-offs

- Passo de codegen no build (`make gen`) e código gerado versionado.
- Rigidez para queries dinâmicas — mitigada pela válvula pgx puro.
- Menos "mágica" que um ORM: paginação, mapeamentos e joins são escritos à mão.

## Riscos

- Query quebrada contra schema novo passa despercebida → mitigado por `sqlc vet` + job de migrations no CI (ADR 0006).
- Divergência entre structs geradas e domínio → mitigada pela regra de DTOs (ADR 0004): struct do sqlc nunca serializa para HTTP.

## Estratégias para minimizar os trade-offs

- Codegen → alvo `make gen` + verificação no CI de que o gerado está em dia (diff vazio).
- Boilerplate de mapeamento → permitido mapear sqlc row → DTO direto nos módulos transaction-script (ADR 0003); mapeamento triplo cerimonial proibido onde não há domínio no meio.

## Impacto esperado

Cada módulo carrega `queries.sql` + `db/` gerado; repositórios manuais só onde o ADR 0005 os justifica (aggregates `pedido`/`reserva`). O fluxo crítico da disputa de assento depende diretamente dos recursos PG expostos por esta escolha.

## Alternativas consideradas e descartadas

- **GORM** — descartada: reflection em runtime, SQL implícito, atrito com partial unique index/ON CONFLICT/guardas — esconde exatamente o que o projeto quer aprender e o que os fluxos críticos exigem.
- **ent** — descartada: schema-as-code conflita com o fluxo migration-first (§6.10); abstração pesada para ~10 tabelas.
- **pgx puro em tudo** — descartada como padrão: boilerplate de scan manual em todo módulo; permanece como válvula para SQL dinâmico.
- **database/sql + lib/pq** — descartada: lib/pq em manutenção mínima; pgx é o driver recomendado e mais performático.

## ADRs relacionados

- ADR 0001 (stack) — este depende dele.
- ADR 0003 (camadas) — define quem chama as queries (service) e onde vivem.
- ADR 0005 (DDD tático) — define os value objects mapeados via `overrides` e os repositórios manuais dos aggregates.
- ADR 0006 (testes) — testcontainers valida o SQL real; `sqlc vet` no CI.
