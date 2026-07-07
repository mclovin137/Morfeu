# ADR 0003 — Fronteiras de módulos e camadas do monolito

- **Status:** aceito
- **Data:** 2026-07-07
- **Task/PRD relacionados:** descoberta (`doc.md` §4), confronto de padrões de código (2026-07-07)

## Contexto

O usuário propôs padronizar a estrutura em cinco camadas fixas — **Controller → Actor → Resolver → Service → DAO** — em todos os módulos. A análise dos agentes (arquiteto, backend-dev, QA) mostrou que a pilha fixa contradiz o desenho aprovado na descoberta (package-by-domain), custaria +4–7 semanas de plumbing (~2.500–4.000 linhas que só delegam), degradaria testes (mock camada a camada) e produziria um dialeto inexistente no ecossistema Go — o risco nº 1 do roadmap ("escrever Java em Go"). Porém **cada intenção por trás das cinco camadas é legítima e tem endereço no desenho**. Esta decisão fixa esse mapeamento. O usuário aceitou a recomendação unânime no confronto de 2026-07-07.

## Escopo

Cobre: estrutura interna dos módulos, camadas mínimas e opcionais, regra de imports entre módulos, composition root, layout de arquivos. Não cobre: estilo de código (ADR 0004), DDD tático e patterns (ADR 0005).

## Decisão

**Camadas mínimas obrigatórias `handler → service → sqlc` em todo módulo; camadas adicionais só com critério declarado. As cinco intenções do usuário mapeiam assim:**

| Conceito proposto | Onde vive no Morfeu | Obrigatório? |
|---|---|---|
| Controller | `handler.go`: bind, validação de forma, DTO↔domínio, erro→HTTP. Zero regra de negócio; `echo.Context` não passa do handler | Sim |
| Actor (orquestra caso de uso) | **Orquestrador** — só quando o caso de uso tem múltiplos passos com compensação: a saga em `pedido/checkout/saga.go` (único do MVP); primos assíncronos: sweeper de `reserva`, consumidor de `notificacao` | Opcional (critério: multi-passo com compensação) |
| Resolver (dependências) | **Composition root em `cmd/morfeu/main.go`**: wiring manual, escolha real×fake por config em tempo de boot | Sim (único no binário) |
| Resolver (caminhos de execução) | Decisão de domínio → máquina de estados (ADR 0005); decisão de infra → wiring | dissolvido |
| Service | `service.go`: regras de negócio + **único dono da transação** | Sim |
| DAO | Dados: pacote `db/` **gerado pelo sqlc**. Integrações: **ports & adapters** (ex.: `pagamento/gateway.go` + `stripe.go`/`fake.go`) | Dados: sim (gerado); repository manual: opcional (ADR 0005) |

**Regras de fronteira entre módulos** (materializam o ownership do `doc.md` §4):
1. Módulo não importa outro módulo — comunicação síncrona por interface pequena definida no consumidor; assíncrona por evento tipado via outbox→RabbitMQ.
2. Nenhum módulo lê tabela alheia (ownership estrito).
3. Interface só com segundo implementador real ou fronteira de módulo declarada.
4. Enforcement: `depguard` no golangci-lint (config validada no Context7 no E0) + item da auditoria.

**Layout padrão de módulo** (fixado no E0): `handler.go`, `service.go`, `errors.go`, `queries.sql`, `db/` (gerado), `*_test.go`. Módulo `pedido` adiciona `pedido.go` (aggregate), `checkout/saga.go`, `checkout/estados.go`, `eventos.go`, `pagamento/`. Sem sub-pacotes além disso até doer. Proibidos pacotes `utils`/`common`/`helpers`.

## Tecnologias ou padrões envolvidos

Package-by-domain sob `internal/`, composition root, ports & adapters, depguard.

## Benefícios

- Intenções do usuário 100% preservadas com nomes idiomáticos → aprende-se o Go que existe no ecossistema.
- ~40% menos arquivos por módulo CRUD; primeira regra de negócio no 2º arquivo aberto.
- Tasks cabem no limite de 30 arquivos (§6.3) com funcionalidade, não plumbing.
- Testes por caso de uso com fakes nas bordas (ADR 0006) em vez de mock camada a camada.

## Trade-offs

- Perde-se a uniformidade "todo módulo tem as mesmas 5 caixas" — módulos simples e críticos têm anatomia diferente.
- Critérios ("multi-passo com compensação", "segundo implementador real") exigem julgamento em refinamento, não aplicação mecânica.

## Riscos

- Camada opcional virar padrão por inércia (orquestrador em CRUD) → auditoria trata indireção sem segundo uso como achado de overengineering.
- Fronteira de import violada silenciosamente → depguard no CI.

## Estratégias para minimizar os trade-offs

- Anatomia por tipo de módulo documentada aqui e no `doc.md` §4 — o refinamento de cada task declara qual anatomia usa.
- Julgamento dos critérios → decidido no refinamento multiagente (§6.14), nunca ad hoc na implementação.

## Impacto esperado

Estrutura de diretórios do E0 nasce deste ADR; PRDs referenciam a anatomia; auditoria ganha dois itens verificáveis (imports entre módulos; indireção sem uso).

## Alternativas consideradas e descartadas

- **Cinco camadas fixas em todo módulo (proposta original)** — descartada: +4–7 semanas de plumbing, stack traces 4 saltos maiores, mock-do-mock nos testes, dialeto fora do ecossistema Go; contradiz o desenho aprovado na descoberta. Confrontada e recusada pelo usuário em 2026-07-07.
- **Cinco camadas só no módulo `pedido`** — descartada: o checkout já ganha o orquestrador e os ports pelo critério geral; nomear as camadas de Actor/Resolver só ali criaria vocabulário exclusivo de um módulo.
- **DAO manual sobre o sqlc por padrão** — descartada: reescreveria à mão, método a método, o contrato que o gerador já produz tipado (manutenção dobrada).
- **Container/framework de DI** — descartada: wiring manual no `main` é a mitigação registrada do risco Java-em-Go.

## ADRs relacionados

- ADR 0001 (stack) e ADR 0002 (dados) — base desta estrutura.
- ADR 0004 (padrões de código) — complementa com regras de estilo dentro das camadas.
- ADR 0005 (DDD tático) — define quando o repository manual e o aggregate entram.
- ADR 0006 (testes) — testa no nível do caso de uso definido aqui.
