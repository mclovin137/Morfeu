# Playbook Database — problemas comuns de PostgreSQL e SQL no Morfeu

> **Como usar**: referência de consulta seletiva do `backend-dev` — NÃO ler inteiro; o índice
> mapeia gatilho → seção. Cada seção responde: onde o problema aparece no Morfeu, **como
> identificar** (sintoma + ferramenta) e **como resolver** (com "qual opção quando").
>
> Contexto fixo: PostgreSQL único (free tier), acesso via **sqlc + pgx v5.9** (ADR 0002),
> migrations via `golang-migrate` (roles.md §6.10). Domínio: cartaz/sessões (leitura pesada),
> mapa de assentos com trava (contenção), pedidos/pagamentos (integridade acima de tudo).
> Skills de apoio relacionadas: `golang-database`, `postgres-patterns`.

| # | Problema | Consultar quando… |
|---|----------|-------------------|
| 1 | [N+1 queries](#1-n1-queries) | loop chamando query; listagem com relacionamento |
| 2 | [Falta de índice](#2-falta-de-índice--seq-scan) | query lenta, filtro/join/order sem índice |
| 3 | [Índice inútil ou em excesso](#3-índice-inútil-ou-em-excesso) | criar índice "por via das dúvidas" |
| 4 | [Deadlock](#4-deadlock) | erro 40P01; transações travando entre si |
| 5 | [Contenção de lock / transação longa](#5-contenção-de-lock--transação-longa) | latência sob concorrência; lock esperando lock |
| 6 | [Pool esgotado / conexão vazada](#6-pool-esgotado--conexão-vazada) | timeout ao pegar conexão; conexões `idle in transaction` |
| 7 | [Race check-then-act](#7-race-check-then-act) | "verifica se existe, depois insere/atualiza" |
| 8 | [Isolamento e anomalias](#8-isolamento-e-anomalias) | lost update, leitura inconsistente entre duas queries |
| 9 | [Paginação com OFFSET](#9-paginação-com-offset) | listagens paginadas (pedidos, sessões) |
| 10 | [Armadilhas de NULL](#10-armadilhas-de-null) | NOT IN, UNIQUE, agregados, comparações com NULL |
| 11 | [Tipos errados](#11-tipos-errados) | dinheiro, datas/horas, IDs, texto livre |
| 12 | [Validação só na aplicação](#12-validação-só-na-aplicação) | regra de integridade sem constraint no banco |
| 13 | [Bloat e autovacuum](#13-bloat-e-autovacuum) | tabela com muito UPDATE/DELETE crescendo/lenta |
| 14 | [Diagnóstico de query lenta](#14-diagnóstico-de-query-lenta) | qualquer investigação de performance |
| 15 | [Timeouts](#15-timeouts) | query presa segurando recursos; lock infinito |
| 16 | [Busca de texto](#16-busca-de-texto) | busca de filme/título no cartaz |
| 17 | [Pegadinhas sqlc/pgx](#17-pegadinhas-sqlcpgx) | scan de NULL, batch, tipos gerados, transação com pgx |
| 18 | [Checklist de revisão de query](#18-checklist-de-revisão-de-query) | toda query nova em PR |

---

## 1. N+1 queries

**Onde aparece no Morfeu**: listar sessões do dia e, para cada uma, buscar o filme (1 query para N sessões + N queries de filme). Backoffice listando pedidos + itens.

**Como identificar**
- Código: query dentro de `for` sobre resultado de outra query — revisão pega isso de graça.
- Runtime: logs de SQL mostrando a mesma query repetida com parâmetros diferentes em rajada; latência da rota cresce linear com o tamanho da lista.

**Como resolver**
1. **JOIN** trazendo tudo numa query (sqlc gera o struct composto) — default para leitura 1:1 ou 1:poucos.
2. **Duas queries + junção em memória** — `WHERE filme_id = ANY($1)` com os IDs coletados; melhor que JOIN quando a relação multiplica linhas (1:N grande) ou os dados do lado N são pesados.
3. Nunca "resolver" com cache: N+1 cacheado continua N+1 no miss.

## 2. Falta de índice / seq scan

**Onde aparece no Morfeu**: `sessoes WHERE data = $1`, `assentos WHERE sessao_id = $1`, `pedidos WHERE usuario_id = $1`, FKs em geral — Postgres **não** indexa FK automaticamente.

**Como identificar**
- `EXPLAIN (ANALYZE, BUFFERS)` mostrando `Seq Scan` em tabela que cresce, ou `Rows Removed by Filter` alto.
- Query lenta só quando a tabela enche (dev com 10 linhas nunca revela).

**Como resolver**
1. Índice b-tree na(s) coluna(s) do filtro/join — ordem importa em índice composto: coluna de igualdade primeiro, range depois (`(sessao_id, status)` e não o inverso se filtra `sessao_id =` e `status =` varia menos).
2. Índice parcial quando só um subconjunto é consultado: `CREATE INDEX ... ON pedidos (usuario_id) WHERE status != 'expirado'`.
3. Covering index (`INCLUDE`) só se o `EXPLAIN` mostrar Heap Fetches doendo — não antecipar.
4. Toda FK criada em migration ganha índice na mesma migration, salvo justificativa.

## 3. Índice inútil ou em excesso

**Onde aparece no Morfeu**: tentação de indexar tudo "por segurança". Cada índice custa escrita (INSERT/UPDATE nas tabelas quentes de assento/pedido pagam por todos os índices) e espaço.

**Como identificar**
- `pg_stat_user_indexes.idx_scan = 0` depois de tempo razoável de uso real.
- Índice em coluna de baixíssima cardinalidade sozinho (ex.: `status` com 4 valores) — o planner raramente usa.
- Índices redundantes: `(a)` e `(a, b)` — o composto cobre o simples.

**Como resolver**: remover os não usados (migration de `DROP INDEX`); preferir 1 índice composto certo a 2 simples; índice em coluna de baixa cardinalidade só como parcial (§2.2) ou parte de composto.

## 4. Deadlock

**Onde aparece no Morfeu**: dois checkouts travando os mesmos assentos em ordens diferentes (A trava assento 5 e quer o 7; B travou o 7 e quer o 5). Erro `40P01 deadlock detected` — o Postgres mata uma das transações.

**Como identificar**
- Erro 40P01 nos logs (o log do Postgres mostra as duas queries envolvidas — ler, ele entrega o par exato).
- Testes de concorrência (QA deve ter um cenário de checkout simultâneo dos mesmos assentos).

**Como resolver**
1. **Ordem determinística de travamento** — sempre travar assentos em ordem crescente de ID (`ORDER BY id FOR UPDATE`). Elimina a classe do problema; é a regra do checkout.
2. Transações curtas (§5) — menos tempo segurando lock, menos janela de ciclo.
3. Retry da transação inteira no erro 40P01 (é seguro: uma delas foi revertida por completo) — camada de repositório, com limite de tentativas.
4. `NOWAIT`/`SKIP LOCKED` onde esperar não faz sentido (assento já travado = "indisponível" na hora — ver playbook-backend §10).

## 5. Contenção de lock / transação longa

**Onde aparece no Morfeu**: transação do checkout que abre, trava assentos e então chama o Stripe **dentro da transação** — os assentos ficam travados pela latência de rede de terceiro. Migration com `ALTER TABLE` esperando lock atrás de uma transação longa, enfileirando todo mundo.

**Como identificar**
- `pg_stat_activity` com `state = 'idle in transaction'` ou `wait_event_type = 'Lock'`; `pg_locks` com `granted = false`.
- Latência p99 alta só sob concorrência, com CPU do banco baixa (todo mundo esperando, ninguém trabalhando).

**Como resolver**
1. **Regra de ouro: nenhuma chamada externa (Stripe, SMTP, Redis, HTTP) dentro de transação aberta.** Padrão do checkout: transação 1 trava e marca assentos `reservado` com TTL; fora de transação chama Stripe; transação 2 confirma. A "trava" que atravessa a chamada externa é o **estado + expiração**, não o lock do banco.
2. Transação = menor unidade de escrita consistente; ler antes, escrever rápido, commitar.
3. `lock_timeout` curto em migrations (§15) para não enfileirar produção atrás de um `ALTER`.

## 6. Pool esgotado / conexão vazada

**Onde aparece no Morfeu**: free tier limita conexões (~20-25); pgx `MaxConns` além disso gera erro no servidor, e vazamento (Rows não fechado, transação sem commit/rollback) esgota o pool e a API inteira para.

**Como identificar**
- Erro `too many clients already` (servidor) ou timeout no `Acquire` do pool (app).
- `pg_stat_activity`: conexões `idle in transaction` acumulando = transação vazada; contagem colada no `MaxConns` constante = pool subdimensionado ou vazamento.
- Métrica do pgxpool (`AcquireDuration`, `TotalConns`) no OTel — exigência de observabilidade (roles.md §6.8).

**Como resolver**
1. Vazamento: `defer rows.Close()` sempre; transação sempre com `defer tx.Rollback(ctx)` imediatamente após `Begin` (rollback pós-commit é no-op seguro). sqlc gerado já fecha certo — o risco é código manual com pgx.
2. Dimensionar: `MaxConns` = limite do servidor − margem (superuser/migrations); pool é a **válvula de back pressure** (playbook-backend §7) — esgotou, requests esperam com timeout de contexto e falham rápido, não derrubam o banco.
3. `MinConns` baixo (1–2) no free tier (cold start, playbook-backend §19).

## 7. Race check-then-act

**Onde aparece no Morfeu**: "SELECT para ver se o assento está livre → INSERT da reserva" — dois requests passam no SELECT juntos e ambos inserem. Idem para "e-mail já cadastrado?", "evento já processado?" (outbox/webhook).

**Como identificar**: qualquer par SELECT-depois-escreve no código onde a decisão depende do SELECT. Em revisão, perguntar: "o que acontece se dois requests executarem isso na mesma milissegundo?" Teste de concorrência do QA confirma.

**Como resolver**
1. **Constraint UNIQUE + `ON CONFLICT`** — o banco é o árbitro: `INSERT ... ON CONFLICT DO NOTHING` e verificar linhas afetadas (0 = perdeu a corrida). Default para dedup (eventos processados, e-mail único).
2. **`SELECT ... FOR UPDATE`** antes da decisão — serializa os concorrentes na linha; para o fluxo de assentos (que precisa ler estado + decidir + escrever).
3. **UPDATE condicional atômico** — `UPDATE assentos SET status='reservado' WHERE id=$1 AND status='livre'`, checar `RowsAffected` — a alternativa sem lock explícito; ótima quando a transição de estado é simples.
4. Nunca resolver com mutex em Go: não sobrevive a 2 instâncias nem a restart.

## 8. Isolamento e anomalias

**Onde aparece no Morfeu**: `READ COMMITTED` (default do Postgres) permite **lost update**: duas transações leem o pedido, ambas calculam e escrevem — a última sobrescreve a primeira. Relatório do backoffice somando valores enquanto pedidos mudam entre queries (read skew).

**Como identificar**: revisão de todo fluxo read-modify-write sem `FOR UPDATE`; totais de relatório que "não batem" por pouco; bugs irreproduzíveis que só acontecem sob carga.

**Como resolver**
1. Manter `READ COMMITTED` + **locking explícito** (`FOR UPDATE`, §7.2) ou update atômico (§7.3) nos fluxos de escrita crítica — é a estratégia do projeto: barata, previsível.
2. **Lock otimista** (coluna `version`, `UPDATE ... WHERE version = $n`) — quando o conflito é raro e segurar lock atrapalha; candidato para edições do backoffice.
3. `REPEATABLE READ` para relatórios multi-query que precisam de snapshot consistente.
4. `SERIALIZABLE` — resolve tudo ao custo de retries obrigatórios (erro 40001); só se um fluxo provar que precisa; não é o default do projeto.

## 9. Paginação com OFFSET

**Onde aparece no Morfeu**: histórico de pedidos no backoffice, listagens administrativas.

**Como identificar**: `LIMIT $1 OFFSET $2` em query de listagem. Sintomas em produção: páginas altas cada vez mais lentas (OFFSET lê e descarta tudo antes); itens pulados/duplicados quando inserem linhas entre uma página e outra.

**Como resolver**
1. **Keyset/cursor**: `WHERE (criado_em, id) < ($cursor_ts, $cursor_id) ORDER BY criado_em DESC, id DESC LIMIT $n` — estável e O(página). Default para listagem que cresce sem limite (pedidos). Exige índice na chave do cursor e desempate por coluna única (`id`).
2. OFFSET é aceitável quando: tabela pequena e limitada (cartaz, salas), UI precisa de "pular para página 7", ou uso interno raro. Não reescrever esses.

## 10. Armadilhas de NULL

**Onde aparece no Morfeu**: colunas opcionais (ex.: `pedidos.pago_em`, `sessoes.cancelada_em`).

**Como identificar / casos clássicos**
- `WHERE coluna != 'x'` **exclui** as linhas NULL silenciosamente (NULL não é `!= 'x'`).
- `NOT IN (subquery)` retorna **vazio** se a subquery contiver um NULL — bug clássico e silencioso.
- `UNIQUE` permite múltiplos NULLs (dois registros "sem CPF" passam num UNIQUE de CPF).
- `COUNT(coluna)` ignora NULL vs `COUNT(*)` que conta tudo — relatórios divergem.

**Como resolver**
1. Preferir `NOT NULL` + valor semântico/default sempre que existir um ("não pago" pode ser `pago_em IS NULL` legítimo — mas então **todas** as queries tratam o NULL conscientemente).
2. `NOT EXISTS` no lugar de `NOT IN` com subquery — imune ao NULL e geralmente plano melhor.
3. `IS DISTINCT FROM` quando NULL deve comparar como valor.
4. Postgres 15+: `UNIQUE NULLS NOT DISTINCT` se NULL duplicado for indesejado.
5. No Go: coluna nullable → tipo nullable no sqlc (§17); nunca mapear NULL para zero-value mudo.

## 11. Tipos errados

**Onde aparece no Morfeu**: preço do ingresso, horários de sessão (timezone!), IDs públicos de pedido/ingresso.

**Como identificar / como resolver**
| Dado | Errado | Certo no Morfeu |
|------|--------|-----------------|
| Dinheiro | `float`/`real` (erro de arredondamento acumula), `money` (tipo problemático) | **centavos em `BIGINT`** (ou `NUMERIC(12,2)` se precisar de fração); moeda implícita BRL, explicitar se um dia variar |
| Data/hora | `timestamp` (sem tz — hora da sessão ambígua no horário de verão/UTC) | **`timestamptz`** sempre; app trabalha em UTC, converte na borda |
| ID público | `serial`/sequencial exposto na URL (enumerável — alguém itera `/pedidos/1,2,3`) | **UUID** (`gen_random_uuid()`) para tudo que aparece em URL/QR; sequencial pode existir como PK interna se justificar |
| Texto | `varchar(255)` cargo-cult | `text` + `CHECK (char_length(...) <= n)` quando o limite é regra de negócio |
| Enum de status | string livre | `CHECK (status IN (...))` ou tipo ENUM — CHECK é mais fácil de evoluir em migration |

Errar tipo é caro de consertar depois (migration + backfill + código) — acertar na primeira migration.

## 12. Validação só na aplicação

**Onde aparece no Morfeu**: "o código já garante que não vende assento duplicado" — até o bug, o retry, o script manual ou a segunda instância.

**Como identificar**: em revisão de migration, perguntar por cada invariante do domínio: onde o **banco** garante isso? Invariantes do Morfeu que exigem constraint: um ingresso por assento por sessão (`UNIQUE (sessao_id, assento_id)` nos ingressos ativos — parcial se cancelado libera), valor não negativo (`CHECK (valor_centavos >= 0)`), FKs com `ON DELETE` **explícito e pensado** (`RESTRICT` por default — nunca cascade acidental apagando pedidos).

**Como resolver**: constraint no banco é a última linha de defesa e a única à prova de concorrência (§7); validação na app é UX (erro bonito antes), não integridade. As duas coexistem — mas só a do banco é obrigatória.

## 13. Bloat e autovacuum

**Onde aparece no Morfeu**: tabelas de alta rotatividade — travas/reservas de assento expirando (UPDATE/DELETE constantes), outbox (INSERT + marca enviado + limpeza). Dead tuples acumulam, tabela e índices incham, scans ficam lentos.

**Como identificar**: `pg_stat_user_tables` (`n_dead_tup` alto vs `n_live_tup`, `last_autovacuum` antigo); tabela cujo tamanho físico só cresce com contagem estável.

**Como resolver**
1. Confiar no autovacuum e **não desligar**; em tabela quente, afinar por tabela (`autovacuum_vacuum_scale_factor` menor) — só com evidência do §14.
2. Outbox: apagar enviados em lotes pequenos e frequentes (não `DELETE` gigante mensal); considerar reter pouco (dias, não meses).
3. Design que evita bloat: reserva de assento como **linha com estado + expiração consultável** (`WHERE expira_em > now()`) atualizada in-place é ok em volume de um cinema; não criar/deletar linhas por tentativa de checkout.
4. `VACUUM FULL` trava a tabela — último recurso, janela de manutenção.

## 14. Diagnóstico de query lenta

**Ferramentas, na ordem**
1. **`EXPLAIN (ANALYZE, BUFFERS)`** na query suspeita com parâmetros reais — ler de dentro para fora; procurar: `Seq Scan` em tabela grande (§2), estimativa vs real discrepante (`rows=1` estimado, 50k real → estatísticas velhas, rodar `ANALYZE`), `Sort`/`Hash` derramando para disco, loops de Nested Loop com inner caro.
2. **`pg_stat_statements`** — habilitar desde já no compose local; responde "quais queries mais custam no agregado" (a lenta ocasional importa menos que a média×frequência).
3. **`log_min_duration_statement`** (ex.: 200ms em dev) — pega as lentas em uso real sem instrumentar nada.
4. Métricas OTel da app (duração por query name do sqlc) cruzadas com traces (roles.md §6.8).

**Regra**: nenhuma otimização (índice novo, reescrita, cache) sem `EXPLAIN ANALYZE` antes/depois anexado ao PR — "achei que ficaria mais rápido" não é evidência.

## 15. Timeouts

**Onde aparece no Morfeu**: query presa (lock, plano ruim) segurando conexão do pool minguado do free tier; migration esperando lock atrás de transação longa.

**Como resolver (camadas, todas)**
1. **`context.WithTimeout` em toda query** via pgx — cancela do lado do cliente **e** o pgx cancela a query no servidor. Default do projeto: timeout curto em rota síncrona (ex.: 5s), maior em job/consumer.
2. **`statement_timeout`** no nível da conexão/role como backstop do servidor (pega o que escapar do contexto).
3. **`lock_timeout` curto em migrations** (ex.: `5s`) + retry: o `ALTER` falha rápido em vez de enfileirar a produção inteira atrás dele.
4. `idle_in_transaction_session_timeout` — mata transação vazada (§6) antes de virar incidente.

## 16. Busca de texto

**Onde aparece no Morfeu**: buscar filme por título no cartaz.

**Como identificar o problema**: `WHERE titulo ILIKE '%bat%'` — wildcard à esquerda **não usa índice b-tree**; seq scan em toda busca.

**Como resolver**
1. Cartaz de um cinema tem dezenas de filmes: **seq scan está ok** — não otimizar; registrar a decisão e seguir (anti-overengineering).
2. Se crescer (histórico grande): extensão `pg_trgm` + índice GIN (`gin_trgm_ops`) — atende `ILIKE '%x%'` e busca com typo (`similarity`).
3. Full-text search (`tsvector`/`websearch_to_tsquery`) — para busca linguística real (múltiplas palavras, relevância); mais peças móveis, só com requisito concreto.
4. Elasticsearch/typesense — descartar neste porte.

## 17. Pegadinhas sqlc/pgx

**Onde aparece no Morfeu**: toda a camada de dados (ADR 0002). Erros recorrentes da dupla:

1. **Coluna nullable → tipo gerado nullable** (`pgtype.Text`, `*string` conforme config do sqlc). Scan de NULL em tipo não-nullable = erro em runtime, não em compile — conferir no schema, não descobrir em produção.
2. **`Begin`/`Commit` manual**: padrão obrigatório `tx, err := pool.Begin(ctx)` → `defer tx.Rollback(ctx)` **na linha seguinte** → trabalho → `tx.Commit(ctx)`. Rollback após commit é no-op; esquecê-lo vaza transação (§6). Repositórios recebem a interface de querier do sqlc (`WithTx`) para participar da transação do caso de uso — transação pertence ao caso de uso, não ao repositório (fronteiras, ADR 0003).
3. **Erro de "no rows"**: pgx retorna `pgx.ErrNoRows` — traduzir para erro de domínio (`ErrPedidoNaoEncontrado`) na camada de dados; `errors.Is`, nunca comparação de string.
4. **Violações de constraint**: capturar `*pgconn.PgError` e mapear `Code` (23505 unique, 23503 FK, 40P01 deadlock §4, 40001 serialização §8) para erros de domínio — é assim que o §7.1 (ON CONFLICT/unique como árbitro) chega limpo ao handler.
5. **Query dinâmica**: sqlc é estático por design. Filtros opcionais → padrão `WHERE ($1::text IS NULL OR coluna = $1)`, ou queries separadas. **Nunca** concatenar SQL por fora (injection — playbook-security).
6. **`CollectRows`/generics do pgx v5** para código manual; não reimplementar scan na mão.
7. **Batch (`pgx.Batch`)** para muitos INSERTs pequenos (ex.: publicar N linhas na outbox) — 1 round-trip; dentro de transação normal.

## 18. Checklist de revisão de query

Para toda query/migration nova no PR:

1. Tem índice para o filtro/join/order? (`EXPLAIN` se houver dúvida — §2, §14)
2. Roda em loop? (N+1 — §1)
3. Par check-then-act sem constraint/lock? (§7)
4. Transação: curta, sem chamada externa dentro, `defer Rollback` presente? (§5, §17.2)
5. Locks em ordem determinística? `NOWAIT`/`SKIP LOCKED` onde esperar não vale? (§4)
6. NULL: alguma comparação/`NOT IN`/UNIQUE afetada? (§10)
7. Tipos: dinheiro em centavos, `timestamptz`, ID público não-enumerável? (§11)
8. Invariante de negócio tem constraint correspondente no banco? (§12)
9. Paginação: keyset se a tabela cresce sem limite? (§9)
10. Timeout de contexto na chamada; `lock_timeout` se for migration com `ALTER`? (§15)
