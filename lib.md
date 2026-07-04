# lib.md — Registro de Dependências

Toda dependência do projeto é registrada aqui **antes** de entrar, com justificativa (regras em `roles.md` §6.9). Nenhuma dependência "de carona".

## Dependências de runtime

| Nome | Versão | Finalidade | Onde é usada | Justificativa | Alternativas consideradas | Riscos | CVEs conhecidas | Última validação | Docs |
|------|--------|-----------|--------------|---------------|---------------------------|--------|-----------------|------------------|------|
| — | | *nenhuma ainda — o scaffold não tem dependências* | | | | | | | |

## Ferramentas de desenvolvimento

| Nome | Versão | Finalidade | Onde é usada | Justificativa | Alternativas consideradas | Riscos | CVEs conhecidas | Última validação | Docs |
|------|--------|-----------|--------------|---------------|---------------------------|--------|-----------------|------------------|------|
| @upstash/context7-mcp | latest (via npx) | servidor MCP de documentação atualizada de bibliotecas | `.mcp.json` (sessões Claude Code) | reduz decisões baseadas em docs desatualizadas (roles.md §6.12) | consulta manual a docs (mais lenta, sujeita a versão errada) | execução via npx baixa pacote na 1ª execução (exige rede) | nenhuma conhecida | 2026-07-03 | https://github.com/upstash/context7 |

## Skills vendorizadas (não são dependências de código)

Todas em `.claude/skills/` — o repositório é autossuficiente ao clonar.

- **De [affaan-m/ECC](https://github.com/affaan-m/ECC) (MIT), 2026-07-03**: `tdd-workflow`, `e2e-testing`, `github-ops`, `git-workflow`, `architecture-decision-records`, `coding-standards`, `api-design`, `error-handling`, `deployment-patterns`, `security-scan`.
- **Migradas do nível de usuário em 2026-07-04** (frontmatter `origin: ECC`): `postgres-patterns`, `docker-patterns`, `database-migrations`, `benchmark`, `browser-qa`, `design-system`, `jpa-patterns`.
- **Migradas do nível de usuário em 2026-07-04** (sem marcador de origem no frontmatter; procedência provável ECC, não confirmada): `backend-patterns`, `frontend-patterns`, `security-review`.

Atualização: re-baixar do repositório de origem e reaplicar o rodapé de crédito. Plugin `ui-ux-pro-max` declarado em `.claude/settings.json` (marketplace GitHub `nextlevelbuilder/ui-ux-pro-max-skill`).
