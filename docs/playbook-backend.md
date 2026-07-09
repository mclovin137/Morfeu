# Playbook Backend — padrões de sistemas distribuídos no Morfeu

> **Como usar**: este arquivo é referência de consulta seletiva. NÃO ler inteiro — o índice abaixo
> mapeia gatilho → seção; leia apenas as seções que a task corrente toca. Cada seção responde:
> onde o problema aparece no Morfeu, formas de resolver e qual forma usar em cada cenário.
>
> Contexto fixo do projeto: Go + Echo (ADR 0001), PostgreSQL com sqlc/pgx (ADR 0002), Redis,
> RabbitMQ, Stripe sandbox, e-mail com QR. Domínio **CP** (assentos e dinheiro nunca sacrificam
> consistência — ver §5). Dev solo, free tier, um único Postgres — vários padrões daqui existem
> para serem **conhecidos e descartados com justificativa**, não aplicados por default
> (anti-overengineering, roles.md §6).

| # | Tópico | Consultar quando a task envolver… |
|---|--------|-----------------------------------|
| 1 | [Idempotência](#1-idempotência) | webhook, consumer RabbitMQ, endpoint de escrita, retry |
| 2 | [Transações distribuídas](#2-transações-distribuídas) | fluxo que toca Postgres + Stripe/e-mail/fila |
| 3 | [Consistência eventual](#3-consistência-eventual) | decidir o que pode atrasar vs. o que é imediato |
| 4 | [Read replicas](#4-read-replicas) | escala de leitura, lag de replicação |
| 5 | [Teorema de CAP](#5-teorema-de-cap) | decisão sobre comportamento sob falha/partição |
| 6 | [Exactly once](#6-exactly-once) | garantia de entrega em fila/webhook |
| 7 | [Back pressure](#7-back-pressure) | consumer, worker pool, pico de tráfego |
| 8 | [Thundering herd](#8-thundering-herd) | cache miss em massa, expiração simultânea |
| 9 | [Cache invalidation](#9-cache-invalidation) | qualquer escrita em dado cacheado |
| 10 | [Hot rows / celebrity](#10-celebrity-problem--hot-shards) | sessão de estreia, contenção em uma linha |
| 11 | [Circuit breaker](#11-circuit-breaker) | chamada síncrona a Stripe/SMTP/serviço externo |
| 12 | [Feature flags](#12-feature-flags) | ligar/desligar comportamento sem deploy |
| 13 | [Schema evolution](#13-schema-evolution) | migration em tabela existente |
| 14 | [Backfill](#14-backfill) | preencher coluna/tabela nova com dados antigos |
| 15 | [Dual writes](#15-dual-writes) | escrever em dois destinos (banco + fila/cache) |
| 16 | [Shadow tables](#16-shadow-tables) | mudança de schema arriscada, migração de dados |
| 17 | [Rate limiting](#17-rate-limiting) | endpoint público, login, proteção de abuso |
| 18 | [Regras de ouro de cache](#18-cache-invalidation--regras-de-ouro) | reforço transversal do item 9 |
| 19 | [Cold start](#19-cold-start) | free tier que hiberna, boot da aplicação |
| 20 | [Checklist de system design](#20-system-design--checklist) | todo fluxo novo, antes de implementar |

---

## 1. Idempotência

**Onde acontece no Morfeu**
- **Webhook do Stripe**: o Stripe reenvia eventos até receber 2xx — o mesmo `payment_intent.succeeded` pode chegar 2+ vezes. Sem idempotência: ingresso duplicado, e-mail duplicado.
- **Consumer RabbitMQ**: entrega é at-least-once; redelivery após crash do consumer reprocessa a mesma mensagem (e-mail do ingresso enviado duas vezes).
- **Checkout**: duplo clique ou retry do frontend cria dois pedidos para os mesmos assentos.

**Formas de resolver**
1. **Operação naturalmente idempotente** — desenhar como estado absoluto (`UPDATE pedido SET status='pago'`, `UPSERT`), não como incremento/append.
2. **Tabela de eventos processados** — `processed_events(event_id UNIQUE)`; inserir o id **na mesma transação** do efeito; violação de unique = já processado, ack e descarta.
3. **Chave de idempotência do cliente** — header `Idempotency-Key` gerado pelo frontend; servidor guarda chave → resposta e devolve a resposta original em repetição.

**Qual usar quando**
- Webhook Stripe e consumer RabbitMQ → **(2)**, sempre. É a defesa canônica contra at-least-once.
- Endpoint de checkout → **(3)** (o frontend gera a chave ao montar o carrinho) combinado com a trava de assento, que já impede o dano maior.
- Todo o resto → **(1)** por design; é grátis e elimina a classe de problema.

## 2. Transações distribuídas

**Onde acontece no Morfeu**: o fluxo de compra toca três sistemas que não compartilham transação — Postgres (pedido/assentos), Stripe (cobrança) e e-mail (ingresso). Não existe commit atômico entre eles.

**Formas de resolver**
1. **2PC/XA** — commit coordenado entre recursos. Descartar: Stripe e SMTP não participam, e 2PC é frágil mesmo onde existe.
2. **Saga** (sequência de transações locais + compensações) — orquestrada (um coordenador dirige) ou coreografada (cada passo reage a eventos).
3. **Outbox** (ver §15) — o efeito externo vira uma linha na mesma transação do dado, publicada depois.

**Qual usar quando**
- Fluxo de pagamento → **saga orquestrada simples via máquina de estados do pedido**: `pendente → aguardando_pagamento → pago → emitido` (compensação: `expirado`/`cancelado` libera assentos). O "orquestrador" é o próprio handler + webhook; não criar framework de saga.
- Efeitos secundários pós-commit (e-mail, QR) → **outbox** (task 0002 existe para isso).
- Coreografia por eventos → só faria sentido com vários serviços; Morfeu é um monólito modular, não usar.

## 3. Consistência eventual

**Onde acontece no Morfeu**
- **Forte (imediata)**: mapa de assentos durante checkout, status do pedido, cobrança. Fonte da verdade: Postgres, transacional.
- **Eventual (pode atrasar segundos/minutos)**: e-mail com o ingresso, contadores de ocupação no backoffice, cartaz cacheado no Redis.

**Formas de resolver / conviver**
1. Definir explicitamente, por dado, a janela de staleness aceitável (0 = forte).
2. UI comunicar o assíncrono ("ingresso enviado por e-mail em instantes") em vez de fingir sincronismo.
3. Read-your-writes onde o autor precisa ver a própria escrita (operador salva filme → cartaz dele reflete na hora: invalidar cache na escrita, ver §9).

**Qual usar quando**: dinheiro e assento → forte, sem exceção (domínio CP). Notificação, relatório, visualização de catálogo → eventual, e dizer isso no PRD para o QA testar a janela, não a instantaneidade.

## 4. Read replicas

**Onde acontece no Morfeu**: hoje, **não acontece** — um único Postgres free tier. Entra aqui para o padrão ser descartado com consciência e para o dia em que leitura de cartaz dominar.

**Formas de resolver (quando existir)**
1. Réplica assíncrona + roteamento leitura/escrita na camada de dados.
2. Mitigar o **replication lag**: read-your-writes roteando para o primário logo após a escrita do próprio usuário; ou sticky no primário por N segundos pós-escrita.

**Qual usar quando**
- Morfeu hoje → **não usar**; o passo anterior e suficiente é cache Redis do cartaz (§9). Réplica só quando o cache não bastar — e o cartaz de um cinema único nunca vai chegar lá.
- Se um dia existir: só leituras tolerantes a staleness vão à réplica (cartaz, relatórios); checkout e mapa de assentos **sempre** no primário.

## 5. Teorema de CAP

**Contexto no Morfeu**: já decidido (pergunta CAP da descoberta + ADR 0002): o domínio é **CP**. Sob qualquer falha/partição entre app e Postgres, **falhar a operação** em vez de arriscar vender o mesmo assento duas vezes. Erro claro > resposta possivelmente errada.

**Formas de aplicar**
1. Escritas críticas: sem fallback "otimista" quando o banco está inacessível — retornar 503 e deixar o cliente tentar de novo.
2. Leituras de catálogo: podem ser servidas do cache (comportamento AP aceitável) porque cartaz desatualizado por segundos não causa dano.

**Qual usar quando**: a pergunta prática por fluxo é "se eu não conseguir confirmar no Postgres, o que faço?" — assento/pagamento: erro; cartaz: serve o cache. Registrar desvios disso em ADR, nunca decidir inline.

## 6. Exactly once

**Onde acontece no Morfeu**: a tentação aparece na fila (e-mail do ingresso "exatamente uma vez") e no webhook Stripe.

**Formas de resolver**
1. Aceitar a realidade: **exactly-once fim-a-fim não existe** em sistema distribuído com falhas. O que existe é **at-least-once + processamento idempotente = effectively once**.
2. No consumer RabbitMQ: `ack` **somente após** o efeito persistido; dedup pela tabela de eventos processados (§1.2) na mesma transação.
3. At-most-once (ack antes de processar) — descartar: perder um ingresso pago é inaceitável.

**Qual usar quando**: sempre **(1)+(2)** neste projeto. Se um PRD pedir "garantir envio único de e-mail", traduzir para: envio at-least-once + `message_id` deduplicado + no pior caso o cliente recebe o mesmo ingresso duas vezes (inofensivo — o QR é o mesmo).

## 7. Back pressure

**Onde acontece no Morfeu**
- Consumer de e-mail: RabbitMQ pode entregar mais rápido do que o SMTP free tier aceita.
- API em pico de estreia: mais requests simultâneos do que o pool pgx tem conexões.

**Formas de resolver**
1. **Prefetch/QoS no RabbitMQ** — consumer só recebe N mensagens sem ack; a fila segura o resto (a fila É o buffer, e é onde o excesso deve ficar).
2. **Limites explícitos no app** — pool pgx com `MaxConns` dimensionado, worker pool/semáforo (`golang.org/x/sync/semaphore` ou canal com buffer) para trabalho concorrente, `context` com timeout em toda chamada externa.
3. **Rejeitar rápido** — quando o limite estoura, responder 503/`Retry-After` imediatamente em vez de enfileirar request sem limite (fila implícita infinita = latência infinita e OOM).

**Qual usar quando**: consumer → **(1)** com prefetch baixo (começar em 1–5; e-mail não tem pressa). API síncrona → **(2)** para dimensionar e **(3)** como válvula. Nunca "resolver" pico aumentando buffer sem limite — isso só move o colapso para mais tarde.

## 8. Thundering herd

**Onde acontece no Morfeu**: estreia anunciada — centenas abrem o cartaz/mapa de assentos no mesmo minuto. Se a chave de cache do cartaz expira nesse instante, todos os requests vão juntos ao Postgres recalcular a mesma coisa (cache stampede). Variante: app reinicia (deploy/cold start) com cache 100% frio.

**Formas de resolver**
1. **`singleflight`** (`golang.org/x/sync/singleflight`) — N requests concorrentes pela mesma chave viram 1 consulta ao banco; os demais esperam o resultado.
2. **TTL com jitter** — TTL base + aleatório (ex.: 60s ± 15s) para chaves não expirarem em coro.
3. **Refresh antecipado** (serve o valor velho e renova em background antes de expirar) — mais complexo, só para chave realmente quente.

**Qual usar quando**: **(1)** é a defesa default no caminho cache-miss→banco deste projeto (barata, idiomática em Go, skill `golang-concurrency` cobre). **(2)** sempre que houver várias chaves criadas juntas. **(3)** só se métricas mostrarem picos de latência na expiração da chave do cartaz — não implementar preventivamente.

## 9. Cache invalidation

**Onde acontece no Morfeu**: cartaz e sessões em Redis (leitura dominante, muda só quando o operador edita no backoffice). O mapa de assentos é o caso oposto: muda a cada reserva.

**Formas de resolver**
1. **TTL puro** — simples, tolera staleness até o TTL.
2. **Delete-on-write** — toda escrita no backoffice deleta as chaves afetadas; próxima leitura repovoa (cache-aside). Preferir **deletar** a atualizar o valor no cache (atualizar tem race de versão velha vencer a nova).
3. **Versionamento de chave** — `cartaz:v{N}`; escrita incrementa N. Invalidação em massa barata, mas exige guardar/ler a versão.

**Qual usar quando**
- Cartaz/sessões → **(1)+(2)** combinados: TTL de segurança (ex.: 5 min, com jitter §8) + delete nas escritas do backoffice. Cobre operador vendo a própria mudança (§3) e cache órfão.
- Mapa de assentos → **não cachear no Redis**. Fonte da verdade é o Postgres com travas; cache aqui cria exatamente a inconsistência que o domínio CP proíbe. Se a leitura do mapa pesar, otimizar a query/índice antes de cachear.
- **(3)** → só se um dia a invalidação precisar varrer famílias inteiras de chaves.

## 10. Celebrity problem / Hot shards

**Onde acontece no Morfeu**: não há sharding (um Postgres), mas a versão local do problema existe: **hot rows**. A sessão de estreia concentra todo o tráfego — todas as transações de reserva disputam as mesmas linhas de assento/sessão. Um lock grosso "por sessão" serializaria todos os compradores.

**Formas de resolver**
1. **Granularidade fina de trava** — travar **por assento** (`SELECT ... FOR UPDATE` nas linhas dos assentos escolhidos, em ordem determinística para evitar deadlock), nunca a sessão inteira.
2. **`FOR UPDATE NOWAIT` / `SKIP LOCKED`** — falhar/pular imediatamente se o assento já está travado, devolvendo "assento indisponível" na hora em vez de enfileirar espera de lock.
3. **Evitar contadores quentes** — não manter `assentos_livres` como coluna incrementada na linha da sessão (toda compra disputaria essa linha); derivar por `COUNT` ou manter eventual.
4. Sharding/particionamento real — descartar; escala de um cinema nunca justifica.

**Qual usar quando**: **(1)+(2)** são o desenho do checkout (NOWAIT para UX de "alguém acabou de pegar esse assento"). **(3)** é regra de modelagem desde a primeira migration. Skill `golang-database` cobre `SELECT FOR UPDATE` com pgx.

## 11. Circuit breaker

**Onde acontece no Morfeu**: Stripe fora do ar durante checkout; SMTP recusando conexões. Sem breaker, cada request espera o timeout inteiro, ocupando conexão e goroutine — falha externa vira lentidão interna.

**Formas de resolver**
1. **Breaker clássico** (closed → open após N falhas → half-open testa) — ex.: `sony/gobreaker` (registrar em `lib.md` antes, roles.md §6.9).
2. **Timeout + retry com backoff exponencial e jitter** — suficiente quando a chamada não está no caminho síncrono do usuário.
3. **Fallback** — o que responder com o breaker aberto.

**Qual usar quando**
- Stripe (síncrono, usuário esperando) → **(1)** + timeout curto via `context`; fallback = erro honesto "pagamento indisponível, seus assentos ficam reservados por X min" (a trava com TTL já dá essa folga).
- E-mail via consumer (assíncrono) → **(2)** basta: retry com backoff + requeue/DLQ; breaker é opcional porque ninguém está esperando na linha.
- Postgres → **nem um nem outro**: sem banco não há fallback útil (domínio CP, §5); timeout + erro 503.

## 12. Feature flags

**Onde acontece no Morfeu**: alternar Stripe sandbox↔real; ligar backoffice novo gradualmente; desligar envio de e-mail em dev.

**Formas de resolver**
1. **Config/env no boot** — flag lida do ambiente na inicialização; mudar exige restart.
2. **Tabela no banco** — mutável em runtime via backoffice.
3. **Serviço externo** (LaunchDarkly etc.) — targeting, rollout %.

**Qual usar quando**: dev solo + free tier → **(1)** como default (simples, auditável no compose/env, zero dependência). **(2)** apenas se surgir flag que o operador precise alternar sem deploy — hoje não existe. **(3)** descartar: custo e complexidade sem múltiplos usuários de flag. Anti-padrão a evitar: flag que vive para sempre — toda flag nasce com critério de remoção no PRD.

## 13. Schema evolution

**Onde acontece no Morfeu**: toda alteração em tabela existente com app no ar (migrations via `golang-migrate`, obrigatórias — roles.md §6.10, skill `criar-migration`).

**Formas de resolver**
1. **Expand → migrate → contract** (o padrão): adicionar o novo sem quebrar o velho (coluna nullable/`DEFAULT`, tabela nova) → código passa a escrever nos dois/ler do novo → backfill (§14) → apertar constraint (`NOT NULL`) → remover o velho em migration posterior.
2. Mudança destrutiva num passo só (rename/drop direto) — aceitável **somente** enquanto não há produção/dados reais.

**Qual usar quando**: pré-produção (fase atual) → **(2)** é pragmático, mas escrever o `down` sempre. A partir do primeiro deploy com dados reais → **(1)** vira lei; nunca `DROP`/`RENAME` de coluna em uso no mesmo release do código que para de usá-la. Atenção Postgres: `ADD COLUMN` com `DEFAULT` volátil ou `NOT NULL` sem default reescreve/trava a tabela — em tabela grande, separar em passos.

## 14. Backfill

**Onde acontece no Morfeu**: coluna nova precisa de valor para linhas históricas (ex.: adicionar `qr_code_id` a ingressos já emitidos); repovoar tabela derivada.

**Formas de resolver**
1. **No próprio `UPDATE` da migration** — simples, transacional, trava a tabela pela duração.
2. **Em lotes fora da migration** — script/comando que processa `WHERE nova_coluna IS NULL ORDER BY id LIMIT N` com pausa entre lotes; idempotente e retomável (re-rodar continua de onde parou).

**Qual usar quando**: tabelas do Morfeu são pequenas (um cinema) → **(1)** na migration é o default e não vale complexidade extra. **(2)** quando o backfill chamar serviço externo (gerar QR, por exemplo — nunca chamar API externa dentro de migration) ou se alguma tabela crescer a ponto de o lock doer. Regra em ambos: backfill **idempotente** (§1) — re-execução não pode corromper.

## 15. Dual writes

**Onde acontece no Morfeu**: a tentação canônica — no handler do checkout, gravar o pedido no Postgres **e** publicar no RabbitMQ, um após o outro. Se o processo cai entre os dois: pedido pago sem e-mail (ou evento publicado de pedido que sofreu rollback).

**Formas de resolver**
1. **Outbox pattern** — o evento é `INSERT` numa tabela `outbox` **na mesma transação** do dado; um publicador separado lê a outbox e publica no RabbitMQ, marcando como enviado. Falha em qualquer ponto = retry seguro (consumer deduplica, §6).
2. **CDC** (Debezium lendo o WAL) — outbox sem código de publicação.
3. Escrever nos dois e "torcer" — é o anti-padrão; qualquer conforto ("é rápido, quase nunca falha") é ilusão.

**Qual usar quando**: **(1)** é a decisão do projeto — a task 0002 (Outbox + consumer RabbitMQ) implementa exatamente isso; todo efeito externo disparado por escrita no banco passa pela outbox, sem exceção. **(2)** descartar: infra pesada demais para o porte. Detectar **(3)** em revisão: qualquer `publish`/`send` no mesmo fluxo de um `INSERT/UPDATE` fora da outbox é bug de design, não estilo.

## 16. Shadow tables

**Onde acontece no Morfeu**: raro — mudanças arriscadas de modelagem (ex.: reestruturar a tabela de assentos já com vendas reais) ou necessidade de trilha de auditoria de mudanças de pedido.

**Formas de resolver**
1. **Shadow table de migração** — cria a tabela nova, escreve nas duas (idealmente via trigger, não código da app), compara consistência por um período, corta a leitura para a nova, aposenta a velha. É o "expand/contract" (§13) para reestruturação de tabela inteira.
2. **Shadow/audit table de histórico** — trigger `AFTER UPDATE` copia versão anterior para `pedidos_historico`; útil para disputa de cobrança no backoffice.

**Qual usar quando**: **(1)** só com dados de produção valiosos em jogo — antes disso, migration direta (§13.2). **(2)** avaliar quando o backoffice precisar de "quem mudou o quê"; alternativa mais simples: coluna `status` + tabela de eventos do pedido (que o fluxo de saga §2 já pede). Não confundir com dual write de aplicação (§15): shadow via trigger fica dentro da transação — é seguro.

## 17. Rate limiting

**Onde acontece no Morfeu**: login do backoffice (brute force), checkout (abuso/bots segurando assentos), endpoints públicos de catálogo (scraping barato de aguentar, mas ainda assim).

**Formas de resolver**
1. **Token bucket** — permite burst controlado, suaviza média. `golang.org/x/time/rate` in-process, ou middleware do Echo (`middleware.RateLimiter`).
2. **Fixed/sliding window no Redis** — contador por chave (`INCR`+`EXPIRE` / sorted set); necessário quando houver mais de uma instância.
3. **Rate limit do provedor** — Stripe limita do lado deles: tratar 429 com backoff, nunca martelar.

**Qual usar quando**
- Instância única (hoje) → **(1)** via middleware Echo: por IP nos endpoints públicos; por conta+IP no login, com janela mais dura (ex.: 5 tentativas/min) — exigência de security.
- Se escalar horizontalmente → migrar contadores para **(2)** (limite in-process por instância deixa de valer).
- Resposta sempre `429` + `Retry-After`; logar excesso (observabilidade, roles.md §6.8) para distinguir ataque de estreia lotada.

## 18. Cache invalidation — regras de ouro

Reforço transversal do §9, na forma de regras de revisão:

1. **Cache é otimização, nunca fonte de verdade.** Qualquer dado que só exista no Redis é bug (exceção deliberada: contadores de rate limit §17, trava de assento **se** o ADR do checkout assim decidir).
2. **Todo caminho de escrita conhece suas chaves.** Ao revisar um `UPDATE`, a pergunta é "quais chaves de cache isso invalida?" — resposta "nenhuma" precisa ser demonstrável.
3. **Deletar > atualizar** o cache na escrita (elimina a race writer-lento-sobrescreve-valor-novo).
4. **Invalidação em massa lembra o herd** (§8): deletar mil chaves = mil misses simultâneos; jitter/singleflight lá.
5. **TTL sempre, mesmo com delete-on-write** — TTL é o backstop contra a chave que o item 2 esqueceu.

## 19. Cold start

**Onde acontece no Morfeu**: free tier (Fly/Render/etc.) hiberna a instância sem tráfego; a primeira request da noite paga boot do processo + pool pgx + TLS Stripe + cache Redis frio. Também em todo deploy.

**Formas de resolver**
1. **Aceitar e minimizar** — binário Go boota em ms (vantagem da stack, ADR 0001); manter init enxuto: nada de warm-up pesado bloqueando o `main`.
2. **Warming seletivo** — pré-carregar no boot apenas o essencial (chave de cache do cartaz), em goroutine, sem bloquear o listen.
3. **Ping externo** (cron/health checker a cada N min) para impedir hibernação — free tiers costumam proibir ou cobrar; verificar ToS.
4. **Pool lazy vs. eager** — abrir 1–2 conexões no boot (falha rápida se o banco caiu) e crescer sob demanda; `MinConns` baixo no pgx.

**Qual usar quando**: **(1)+(4)** sempre. **(2)** só se a primeira impressão do cartaz importar em métrica real. **(3)** decisão de custo/ToS do usuário, não do dev — escalar como pergunta, não implementar por conta.

## 20. System design — checklist

Antes de implementar qualquer fluxo novo, responder por escrito no PRD (ou apontar a seção daqui que responde):

1. **Fonte da verdade** — qual store manda neste dado? (default: Postgres; Redis nunca — §18.1)
2. **Consistência** — o que precisa ser imediato e o que pode ser eventual? (§3, §5)
3. **Falha externa** — cada chamada a Stripe/SMTP/Redis: o que acontece se falhar, demorar ou **repetir**? (§1, §11)
4. **Atomicidade** — o fluxo escreve em mais de um lugar? Se sim: outbox/saga, nunca dual write (§2, §15)
5. **Hot path** — onde a estreia lotada bate? Existe lock grosso ou contador quente escondido? (§8, §10)
6. **Limites** — o que impede este fluxo de consumir recursos sem teto? (§7, §17)
7. **Evolução** — a mudança de schema segue expand/contract? O backfill é idempotente? (§13, §14)
8. **Simplicidade** — qual desses padrões estou aplicando **sem** cenário que o justifique? Cortar (anti-overengineering, roles.md §6).
