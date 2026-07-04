#!/usr/bin/env bash
# obsidian-session-end.sh — SessionEnd hook: resumo da sessão → vault Obsidian
#
# Ao fechar o chat, extrai um digest do transcript e dispara um agente Claude
# headless (detached) que grava o resumo no vault seguindo as regras do
# _CLAUDE.md (Dev Logs + nota do projeto + daily note + operations log).
#
# Gates (o hook é inerte sem TODOS eles):
#   - OBSIDIAN_VAULT_PATH definido
#   - OBSIDIAN_SESSION_SUMMARY_ENABLED=1
# Anti-recursão (o agente headless também dispara SessionEnd ao terminar):
#   - OBSIDIAN_SESSION_END_CHILD=1 no ambiente do filho → filho nunca re-dispara
#   - cwd dentro do vault → skip (o agente roda com cwd no vault)
#
# Teste: CLAUDE_BIN pode apontar para um stub. Log: /tmp/obsidian-session-end.log

VAULT="${OBSIDIAN_VAULT_PATH:-}"
[[ -z "$VAULT" ]] && exit 0
[[ "${OBSIDIAN_SESSION_SUMMARY_ENABLED:-0}" != "1" ]] && exit 0
[[ "${OBSIDIAN_SESSION_END_CHILD:-0}" == "1" ]] && exit 0

INPUT=$(cat)
TRANSCRIPT=$(printf '%s' "$INPUT" | jq -r '.transcript_path // ""' 2>/dev/null || true)
CWD=$(printf '%s' "$INPUT" | jq -r '.cwd // ""' 2>/dev/null || true)
REASON=$(printf '%s' "$INPUT" | jq -r '.reason // ""' 2>/dev/null || true)
[[ -z "$TRANSCRIPT" || ! -f "$TRANSCRIPT" ]] && exit 0

# Sessão dentro do próprio vault (inclui o agente headless) → não resumir
case "$CWD" in "$VAULT"|"$VAULT"/*) exit 0 ;; esac

# Sessão trivial: menos de 2 prompts reais digitados pelo usuário → não vale nota
USER_COUNT=$(jq -r 'select(.type=="user") | .message.content | if type=="string" then "u" else empty end' "$TRANSCRIPT" 2>/dev/null | wc -l)
[[ "$USER_COUNT" -lt 2 ]] && exit 0

# Digest: mensagens de texto de usuário e assistente, últimos ~30 KB
DIGEST=$(jq -r 'select(.type=="user" or .type=="assistant")
  | (if .type=="user" then "USER: " else "ASSISTANT: " end)
    + (.message.content
       | if type=="string" then .
         elif type=="array" then ([.[] | select(.type=="text") | .text] | join("\n"))
         else "" end)' "$TRANSCRIPT" 2>/dev/null \
  | grep -vE '^(USER|ASSISTANT):[[:space:]]*$' | tail -c 30000)
[[ -z "$DIGEST" ]] && exit 0

TODAY=$(date +%Y-%m-%d)
NOW=$(date +%Hh%M)
PROJECT=$(basename "${CWD:-desconhecido}")

PROMPT_FILE=$(mktemp /tmp/obsidian-session-end-XXXXXX.txt)
cat > "$PROMPT_FILE" << HEADER
You are an autonomous Obsidian vault agent. A Claude Code chat session just ended
(reason: $REASON). Save a session summary to the vault. Run silently.

VAULT: $VAULT
TODAY: $TODAY
TIME: $NOW
PROJECT (from session cwd): $PROJECT ($CWD)

SESSION DIGEST (most recent messages, may be truncated at the start):
HEADER
printf '%s\n\n' "$DIGEST" >> "$PROMPT_FILE"
cat >> "$PROMPT_FILE" << 'INSTRUCTIONS'
INSTRUCTIONS:
1. Read _CLAUDE.md at the vault root first — follow its rules exactly (AI-first
   notes, frontmatter, wikilinks, propagation). Write note CONTENT in the same
   language the vault already uses.
2. Create the session summary note: "Dev Logs/[TODAY] — [PROJECT] sessão [TIME].md"
   with: For future Claude preamble, what was done, problems, decisions, next steps.
   If the digest contains no meaningful work (small talk only), exit silently instead.
3. Propagate (never leave an orphan note):
   - Project note (Projects/[PROJECT].md or closest match): add a Recent Activity
     bullet linking the new note. If no project note exists, do NOT create one.
   - Daily note (Daily/[TODAY].md): add a Work bullet. Create from convention if missing.
   - Operations log (Logs/[TODAY].md): append entries per the vault convention.
   - index.md: register the new note.
4. Before creating anything, search for existing notes — never duplicate. A note for
   this same session (same date+time+project) already existing means: update it.
CONSTRAINTS:
- Filesystem tools only (Read, Write, Edit, Glob, Grep). No questions, no output.
- Do not archive, delete or merge anything — only add or update.
INSTRUCTIONS

CLAUDE_BIN="${CLAUDE_BIN:-claude}"
nohup bash -c '
  cd "$1" || exit 0
  export OBSIDIAN_SESSION_END_CHILD=1
  "$2" --dangerously-skip-permissions -p "$(cat "$3")" >> /tmp/obsidian-session-end.log 2>&1
  rm -f "$3"
' _ "$VAULT" "$CLAUDE_BIN" "$PROMPT_FILE" </dev/null >/dev/null 2>&1 &

exit 0
