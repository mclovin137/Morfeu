# Banco de perguntas da descoberta

Roteiro da entrevista da skill `iniciar-projeto`. Conduzir na ordem, em blocos de até 4 perguntas por chamada de `AskUserQuestion`. As perguntas são o **mínimo** — aprofunde sempre que uma resposta abrir porta (regra: máximo de perguntas possíveis). Adapte: pule apenas o que o usuário confirmar como irrelevante.

## Bloco A — Identidade e visão

1. Qual o **nome do projeto**? (se não houver, propor 2–3 opções a partir do problema)
2. Que **problema** o projeto resolve? Qual a dor de quem sofre com ele hoje?
3. Qual o **objetivo de negócio** mensurável? (ex.: reduzir X, faturar Y, atender Z usuários)
4. Quem é o **público-alvo**? (perfil, tamanho, técnico ou leigo, B2B/B2C/interno)
5. Quais são os **atores** do sistema? (papéis humanos + sistemas externos que interagem)
6. Existe **concorrente ou referência**? O que o projeto faz de diferente?
7. Como o sucesso será medido nos primeiros 6 meses?
8. Há **prazo ou evento** que ancora o lançamento?

## Bloco B — Escopo funcional

1. Liste as **funcionalidades essenciais do MVP** (o que sem o quê o produto não existe).
2. Para cada funcionalidade: prioridade **MoSCoW** (must/should/could/won't).
3. O que está **explicitamente fora de escopo** nesta fase? (registrar — evita scope creep)
4. Quais são os **fluxos críticos** (os 2–3 caminhos que precisam funcionar perfeitamente)?
5. Existem **regras de negócio complexas** (cálculos, prazos, estados, permissões condicionais)?
6. Precisa de **área administrativa/backoffice**? Quem opera?
7. Há **relatórios, exportações ou dashboards** previstos?
8. O sistema tem **frontend próprio** (web/mobile), é só API, ou ambos?

## Bloco C — Requisitos não funcionais

1. **Latência** aceitável nas operações principais? (< 200 ms · < 1 s · segundos tudo bem · batch)
2. **Disponibilidade** alvo? (best effort · horário comercial · 99,9% · missão crítica)
3. O sistema trata **dados pessoais**? Quais? (nome/e-mail · documentos · saúde/financeiro · menores)
4. Há **compliance** aplicável? (LGPD, PCI-DSS, HIPAA, regulação setorial, nenhum)
5. Requisitos de **auditabilidade** — quem fez o quê, quando? Retenção por quanto tempo?
6. **Acessibilidade** (WCAG) e **internacionalização** (idiomas, moedas, fusos) são requisitos?
7. Tolerância a **perda de dados**? (nenhuma · minutos · horas — define RPO/backup)
8. Expectativa de **tempo de recuperação** em desastre? (minutos · horas · dias — define RTO)

## Bloco D — Volumetria

> Resposta vaga não encerra o assunto: ofereça faixas e registre a estimativa assumida.

1. Quantos **usuários totais** e quantos **ativos simultâneos** no lançamento?
2. **Requisições por segundo** esperadas (média e pico)? (faixas: < 10 · 10–100 · 100–1k · > 1k)
3. Qual o **crescimento projetado** em 12 meses? (estável · 2–5× · 10×+ · imprevisível)
4. **Volume de dados**: quantos registros/dia nas entidades principais? Tamanho médio?
5. Proporção **leitura × escrita**? (leitura pesada · equilibrado · escrita pesada)
6. Há **picos sazonais ou eventos** (campanhas, fechamento de mês, horários)?
7. Há **processamento pesado** (relatórios grandes, mídia, ML, jobs em lote)?
8. **Retenção**: dados crescem para sempre ou expiram/arquivam?

## Bloco E — Orçamento e restrições

1. **Budget mensal de infraestrutura**? (≈ R$ 0/free tier · < R$ 250 · R$ 250–1.500 · > R$ 1.500)
2. Há budget para **serviços pagos** (e-mail, SMS, mapas, IA, monitoramento)?
3. **Prazo**: quando o MVP precisa estar no ar? Há marcos intermediários?
4. **Equipe**: quem desenvolve, quem opera, quem decide produto? Nível de experiência na stack?
5. Restrições de **licença** (só open source? copyleft ok?) ou de **contrato** (cloud já contratada)?
6. Existe **sistema legado** a integrar ou substituir? Há migração de dados?
7. Quem paga e como o custo cresce se o projeto der certo? (limite de escala aceitável)

## Bloco F — Stack

> Toda preferência declarada aqui será validada no Context7 e confrontada pelos agentes na Fase 2–3.

1. Há **linguagem/plataforma preferida ou obrigatória**? Por quê? (competência do time conta)
2. Preferência de **framework**? Aberto a alternativas se o trade-off for melhor?
3. **Banco de dados**: relacional, documento, chave-valor? Já existe preferência (ex.: PostgreSQL)?
4. Precisa de **cache** (Redis etc.) desde o dia 1 ou é otimização futura?
5. Há necessidade de **processamento assíncrono/filas** (e-mails, jobs, integrações lentas)?
6. Stack de **frontend** (se houver): framework, SSR/SPA, design system?
7. **Mobile**: nativo, cross-platform, PWA ou fora de escopo?
8. Alguma **tecnologia proibida ou indesejada** (trauma, política interna, custo)?

## Bloco G — Arquitetura e infra

1. Estilo arquitetural: **monolito modular** (default anti-overengineering) ou há justificativa real para distribuir?
2. Onde roda: **cloud** (qual?), **on-premise**, **híbrido**? Já existe conta/contrato?
3. **Containers**: Docker Compose local (padrão do template) é suficiente? Kubernetes se justifica pela volumetria?
4. Quantos **ambientes** (dev · staging · prod)? O que difere entre eles?
5. **CI/CD**: GitHub Actions (padrão do template) atende? Deploy manual ou automático?
6. **Observabilidade**: logs estruturados + métricas + traces desde o dia 1 (roles.md §6.8) — alguma ferramenta já definida?
7. Estratégia de **backup e recuperação** compatível com o RPO/RTO do bloco C?
8. Há requisito de **multi-tenancy** (vários clientes isolados na mesma instância)?

## Bloco H — Dependências e integrações

1. Quais **APIs/sistemas de terceiros** o projeto consome? (docs disponíveis? SLA? custo?)
2. **Autenticação**: própria, IdP (Keycloak/Auth0/Cognito), login social, SSO corporativo?
3. **Autorização**: papéis simples (RBAC) ou permissões finas por recurso?
4. Há **pagamentos**? (gateway, PIX, assinatura recorrente, split)
5. **Notificações**: e-mail, SMS, push, WhatsApp? Volume esperado?
6. **Arquivos/mídia**: upload, armazenamento (S3-like), processamento de imagem/vídeo?
7. O projeto **expõe API para terceiros**? (versionamento, rate limit, documentação — skill `api-design`)
8. Alguma integração é **crítica para o MVP** (bloqueia o lançamento se atrasar)?

## Bloco I — Observabilidade e resiliência (etapa SRE)

> Etapa conduzida com o agente `sre-devops` — o sistema nasce observável (roles.md §6.8). As respostas dos blocos C (RNFs), D (volumetria) e E (orçamento) alimentam este bloco.

1. Quais **ferramentas de observabilidade**? Apresente os cenários com custo × esforço:
   - self-hosted: Prometheus + Grafana (métricas), Loki ou ELK (logs), Tempo/Jaeger (traces), OpenTelemetry como padrão de instrumentação
   - SaaS: Datadog, New Relic, Grafana Cloud (validar custo contra o budget do bloco E)
   - mínimo viável: logs estruturados + métricas do runtime + healthchecks (para MVPs de baixa volumetria)
2. Quais **métricas importam**?
   - técnicas: golden signals (latência, tráfego, erros, saturação) e RED por endpoint (rate, errors, duration); USE para recursos (CPU, memória, disco, pool de conexões)
   - de **negócio**: quais eventos medem sucesso (ex.: pedidos/min, cadastros, conversão)?
3. **SLOs**: quais objetivos mensuráveis derivam dos RNFs do bloco C (ex.: p95 < 300 ms, taxa de erro < 1%)? Haverá error budget?
4. **Logs — como serão monitorados**: formato estruturado (JSON) com correlation/trace ID em toda requisição? Níveis por ambiente? O que NUNCA logar (PII, tokens — roles.md §6.6)? Onde agregar e por quanto tempo reter?
5. **Alertas e dashboards**: o que acorda alguém de madrugada × o que vira ticket? Canal de alerta? Quais dashboards essenciais no dia 1?
6. **Tracing distribuído** é necessário desde o início (múltiplos serviços/filas conversando) ou correlation ID no monolito basta?
7. Quais **padrões de resiliência e consistência** o domínio exige?
   - transações que cruzam serviços/integrações → **saga** (coreografada × orquestrada) com **compensação/rollback definido por passo** — sem compensação escrita, a saga não existe
   - chamadas externas → **timeout + retry com backoff + circuit breaker** (skill `error-handling`)
   - mensageria/async → **outbox pattern**, **DLQ** e **idempotência de consumidores**
   - rollback de deploy → estratégia (blue/green, canary, feature flag) alinhada a migrations expand-contract (roles.md §6.10.4)
8. **Healthchecks**: liveness × readiness — o que cada um valida (banco? fila? dependência externa)?
