# ADR 0005 — DDD tático e padrões de projeto nos módulos críticos

- **Status:** aceito
- **Data:** 2026-07-07
- **Task/PRD relacionados:** confronto de padrões de código (2026-07-07); materializa-se nos épicos E4 (reserva) e E6 (pedido)

## Contexto

O usuário propôs **DDD para as rotinas mais críticas** e os padrões **Strategy, Factory e State**. Os agentes convergiram: as propostas são corretas *onde há variação e invariante reais* — e viram cerimônia onde não há. Este ADR fixa onde cada um entra, em qual forma Go, e a regra de contenção anti-pattern-por-pattern. O usuário aceitou as recomendações no confronto.

## Escopo

Cobre: quais módulos usam DDD tático e com qual kit; forma idiomática de Strategy/State/Factory; regra de contenção. Não cobre: fronteiras de módulos (ADR 0003), detalhes da saga (candidata a ADR própria no refinamento do E6).

## Decisão

### DDD tático em exatamente 2 módulos: `pedido` e `reserva`

São os módulos com invariantes de negócio reais (máx. 6 assentos, total server-side §14.2, transições da saga; TTL 10min + 1 extensão, unicidade de assento). Kit mínimo:

- **Aggregate** = struct com campos não exportados + construtor `NewX` + métodos que protegem invariantes (`Pedido`, `MapaDeAssentos` como read model de validação contra o layout).
- **Value objects** com invariante: `Centavos` (int64), `StatusPedido`/`StatusHold`, `AssentoCodigo` — mapeados via `overrides` do sqlc com parcimônia.
- **Domain events tipados** (`PedidoConfirmado` como struct) → outbox na mesma TX (dá contrato ao consumidor de `notificacao`).
- **Repository manual por aggregate** (mapeia sqlc↔domínio, participa da TX) — os únicos dois repositórios manuais do projeto (ADR 0003).
- **Nota de defesa em profundidade** (coerência com `doc.md` §5): o aggregate é a **primeira** linha (em memória, testável em unidade); o partial unique index e o `UPDATE ... WHERE status = $2` com guarda são a **última** (sob concorrência real). Um não substitui o outro.

**Sem DDD (transaction script)**: `catalogo`, `sessao`, `identidade`, `notificacao`, `plataforma` — handler→service→queries direto. `sessao` tem regra forte (não-sobreposição), mas é verificação de intervalo no banco, não aggregate. `identidade` é crítico em segurança, não em modelagem: procedural + testes rigorosos.

**Não importar do DDD-Java**: `Repository<T,ID>` genérico (o sqlc já é o repositório tipado), Specification pattern (SQL explícito é decisão do ADR 0002), value object para tudo, separação application/domain service, command bus/MediatR, marcador `AggregateRoot`.

### Padrões de projeto — forma Go e lugares reais

- **Strategy = interface pequena definida no consumidor + injeção no construtor.** Dois lugares com segundo implementador real (condição do R$ 0): `pagamento.Gateway` (Stripe × fake do modo load-test) e `notificacao.EmailSender` (Resend/Brevo × fake). Escolha no composition root. Sem registries, sem `Provider`.
- **State = enum + tabela de transições + função `Transicionar` que rejeita transição ilegal**, com linter `exhaustive` cobrindo todo switch sobre status. O enforcement real da corrida webhook × expiração é o **compare-and-swap no SQL** (`UPDATE pedidos SET status='pago' WHERE id=$1 AND status='aguardando_pagamento'`, checando linhas afetadas); a máquina em memória documenta e é testável por tabela. GoF State completo (um tipo por estado) fica documentado como **gatilho de evolução**: só se o comportamento por estado crescer a ponto de encher a função de switches — não é o caso do MVP (o estado é persistido e retomado; reconstruir "qual tipo" da linha do banco dobraria o código).
- **Factory = construtores `NewX(deps)` + uma função de seleção por config no `main`** (`novoGateway(cfg)`). Nenhum tipo `Factory`; AbstractFactory/registry sem caso de uso. Se surgir criação polimórfica real (ex.: múltiplos tipos de layout de sala), o pattern entra pelo PRD correspondente.

### Regra de contenção anti-pattern-por-pattern

1. Nenhum nome de pattern em identificador (`PedidoFactoryImpl` proibido).
2. Pattern entra citado no PRD com a **variação real** que o justifica.
3. Interface só com segundo implementador real ou fronteira de módulo declarada (ADR 0003).
4. Na auditoria, indireção sem segundo uso é achado de **overengineering**.

## Tecnologias ou padrões envolvidos

DDD tático (aggregate, value object, domain event, repository), Strategy/State/Factory em forma idiomática Go, `exhaustive`, `overrides` do sqlc.

## Benefícios

- Invariantes de dinheiro e inventário protegidas por tipo e construtor — as classes de bug mais caras do domínio ficam inexpressáveis.
- A matriz da saga vira table-driven test puro (máquina de estados sem I/O) — o teste mais valioso do projeto custa milissegundos.
- Patterns entram onde já havia variação estrutural (fakes exigidos pelo modo load-test) — custo marginal ~zero.
- Módulos CRUD continuam honestos e rasos — esforço vai para onde há domínio.

## Trade-offs

- Dois "níveis de cerimônia" no codebase (aggregate vs transaction script) — anatomia não uniforme entre módulos.
- Repository manual em 2 módulos = mapeamento sqlc↔domínio escrito à mão.
- Enum + tabela é menos "OO de livro" que GoF State — perde-se o exemplo canônico de portfólio.

## Riscos

- DDD-creep (aggregate surgindo em módulo CRUD) → regra de contenção 2 + auditoria.
- Aggregate e constraint divergirem (invariante só em memória) → PRD dos módulos críticos exige a dupla verificação (teste de unidade no aggregate + teste de integração na constraint, ADR 0006).

## Estratégias para minimizar os trade-offs

- Anatomia dupla → declarada por módulo no ADR 0003; refinamento decide, nunca a implementação ad hoc.
- Mapeamento manual → restrito aos 2 aggregates; o resto usa sqlc direto.
- Evolução do State → gatilho documentado acima; supersede parcial deste ADR se acionado.

## Impacto esperado

E4 e E6 implementam o kit; PRDs desses épicos referenciam este ADR; o refinamento do E6 pode gerar o ADR próprio da saga (candidata nº 5 do `doc.md` §11) com os contratos finais.

## Alternativas consideradas e descartadas

- **DDD em todos os módulos** — descartada: `catalogo`/`sessao`/`identidade` não têm invariantes que paguem aggregates; viraria cerimônia (roles.md §1).
- **Nenhum DDD** — descartada: `pedido` e `reserva` concentram dinheiro + inventário; transaction script ali espalharia invariantes por services e queries.
- **GoF State com um tipo por estado** — descartada para o MVP (estado persistido/retomado; código dobrado); documentada como evolução com gatilho.
- **Strategy/Factory na forma Java** (registries, providers, tipos Factory) — descartada: em Go o pattern é interface + construtor + escolha no wiring.

## ADRs relacionados

- ADR 0002 (sqlc) — `overrides` para value objects; repository manual mapeia sobre ele.
- ADR 0003 (camadas) — define onde aggregate/orquestrador/ports vivem.
- ADR 0004 (código) — regras 3/4/9 do calisthenics adaptado se materializam aqui.
- Futuro: ADR da saga do checkout (refinamento do E6).
