# Boas práticas de SQL para migrations

Referência da skill `criar-migration` (regra em roles.md §6.10). Vale para toda migration e para as queries que ela habilita. Exemplos em PostgreSQL — para outro banco, valide o equivalente no **Context7** antes de aplicar (roles.md §6.12). Aprofundamento: skills `database-migrations` (zero-downtime) e `postgres-patterns` (se PostgreSQL).

## 1. Índices — quando e como criar

- **Toda FK nova recebe índice.** O banco não cria automaticamente (PostgreSQL); FK sem índice vira full scan no JOIN e lock amplificado no DELETE/UPDATE do pai. Exceção: tabela minúscula e estática — justificar no cabeçalho da migration.
- **Indexe os padrões de acesso do PRD**: colunas de `WHERE`, `JOIN`, `ORDER BY` e `GROUP BY` dos fluxos críticos (doc.md §6). Índice se justifica pelo padrão de acesso declarado, nunca "por via das dúvidas".
- **Índice composto**: colunas de igualdade primeiro, range/ordenação por último (`(tenant_id, status, created_at)` atende `WHERE tenant_id = ? AND status = ? ORDER BY created_at`). A ordem importa — o prefixo do índice precisa casar com a query.
- **Índice parcial** quando a query sempre filtra um subconjunto (`WHERE deleted_at IS NULL`, `WHERE status = 'pending'`): menor, mais rápido, menos custo de escrita.
- **Covering index** (`INCLUDE`) apenas em query quente comprovada — evita ida à tabela (index-only scan).
- **Tipo certo**: B-tree é o default correto; GIN para `jsonb`, arrays e full-text; GiST/BRIN para casos específicos (geo, séries temporais enormes). Não presumir: Context7.
- **Não indexar por reflexo**: cada índice custa escrita, espaço e vacuum. Migration que cria mais de ~3 índices numa tabela de escrita intensa precisa de justificativa. Remover índice órfão é migration válida (com evidência de não-uso, ex.: `pg_stat_user_indexes`).

## 2. Evitar full scan

- **Query sargável**: nunca aplicar função ou cast sobre a coluna indexada no predicado.
  - ❌ `WHERE date(created_at) = '2026-07-03'` → quebra o índice.
  - ✅ `WHERE created_at >= '2026-07-03' AND created_at < '2026-07-04'` — ou crie **índice de expressão** se a função for inevitável.
- **`LIKE '%termo%'`** (curinga à esquerda) não usa B-tree. Se busca textual for requisito: full-text search ou índice trigram (`pg_trgm`) — decisão registrada, não improviso.
- **Tipos compatíveis**: comparação com cast implícito (`varchar` × `int`, collation divergente) derruba o índice silenciosamente.
- **`SELECT *` proibido em caminho crítico** — impede index-only scan e infla payload.
- **Paginação por keyset** (`WHERE (created_at, id) < (?, ?) ORDER BY ... LIMIT ?`) em vez de `OFFSET` grande — OFFSET lê e descarta tudo que pula.
- **Prova, não intuição**: rode `EXPLAIN (ANALYZE, BUFFERS)` nas queries dos fluxos críticos sobre volume representativo da volumetria assumida (doc.md §8). `Seq Scan` em tabela grande no caminho crítico **reprova a migration** até ser corrigido ou justificado por escrito (ex.: tabela de 200 linhas — seq scan é o plano certo).

## 3. Locks e zero-downtime

Obrigatório quando existir ambiente com tráfego; recomendado desde o dia 1 para criar o hábito.

- **`CREATE INDEX CONCURRENTLY`** em tabela com tráfego — não bloqueia escrita, mas não roda dentro de transação (a maioria das ferramentas de migration tem flag para isso; Context7).
- **`ADD COLUMN`**: sem `DEFAULT` volátil que force rewrite da tabela (comportamento varia por versão — validar no Context7). `NULL` + backfill é o caminho seguro.
- **`NOT NULL` em tabela grande**: `ADD CONSTRAINT ... CHECK (col IS NOT NULL) NOT VALID` → backfill em lotes → `VALIDATE CONSTRAINT` (não bloqueia leitura/escrita) → aí sim `SET NOT NULL` se necessário.
- **FK em tabela grande**: mesma técnica — `NOT VALID` na criação, `VALIDATE` depois.
- **Backfill em lotes** (1k–10k linhas por iteração), idempotente, **fora da transação do DDL** — nunca DDL + UPDATE massivo na mesma transação (lock longo + WAL gigante).
- **Rename/drop: padrão expand-contract** — adicionar o novo, migrar leitura/escrita, remover o antigo em **task futura**. Nunca rename destrutivo em um passo só com código antigo no ar.
- Transações curtas sempre: migration que segura lock por minutos derruba o sistema tanto quanto um bug.

## 4. Checklist de performance (anexar mentalmente ao checklist da skill)

- [ ] Toda FK nova tem índice (ou justificativa no cabeçalho)
- [ ] Índices novos casam com os padrões de acesso do PRD (ordem do composto verificada)
- [ ] Nenhum índice "por via das dúvidas" — cada um justificado
- [ ] Queries dos fluxos críticos validadas com `EXPLAIN (ANALYZE, BUFFERS)` sobre volume representativo — sem Seq Scan não justificado em tabela grande
- [ ] Predicados sargáveis (sem função/cast sobre coluna indexada; sem LIKE de curinga à esquerda sem estratégia)
- [ ] Operações compatíveis com tráfego: `CONCURRENTLY`, `NOT VALID`/`VALIDATE`, backfill em lotes, expand-contract
- [ ] Comportamento da versão do banco confirmado no Context7 (nada presumido)
