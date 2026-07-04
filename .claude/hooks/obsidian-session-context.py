#!/usr/bin/env python3
"""SessionStart hook: injeta contexto do vault Obsidian ao abrir o chat.

Complementa o load_vault_context.py (que só atua DENTRO do vault): este roda
para sessões de projeto (cwd fora do vault), localiza a nota do projeto em
Projects/<basename(cwd)>.md e injeta um bloco compacto no contexto — preâmbulo
"For future Claude", últimas entradas de Recent Activity e Key Decisions.

Silencioso quando: sem OBSIDIAN_VAULT_PATH, cwd dentro do vault, ou projeto
sem nota no vault. Saída limitada (~3 KB) para não inflar o contexto.
"""
from __future__ import annotations

import json
import os
import sys
from pathlib import Path

MAX_OUT = 3000


def section(text: str, header: str, max_lines: int) -> list[str]:
    lines = text.splitlines()
    out: list[str] = []
    inside = False
    for ln in lines:
        if ln.strip().lower().startswith(f"## {header}".lower()):
            inside = True
            continue
        if inside:
            if ln.startswith("## "):
                break
            if ln.strip():
                out.append(ln.rstrip())
            if len(out) >= max_lines:
                break
    return out


def main() -> int:
    vault = os.environ.get("OBSIDIAN_VAULT_PATH", "")
    if not vault or not os.path.isdir(vault):
        return 0
    try:
        data = json.load(sys.stdin)
    except Exception:
        data = {}
    cwd = data.get("cwd") or os.getcwd()
    vault_norm = os.path.normpath(vault)
    if os.path.normpath(cwd).startswith(vault_norm):
        return 0  # sessão no vault: load_vault_context.py cuida

    project = os.path.basename(os.path.normpath(cwd))
    projects_dir = Path(vault) / "Projects"
    if not projects_dir.is_dir():
        return 0
    note = None
    for p in projects_dir.glob("*.md"):
        if p.stem.lower() == project.lower():
            note = p
            break
    if note is None:
        return 0

    try:
        text = note.read_text(encoding="utf-8", errors="replace")
    except Exception:
        return 0

    preamble = section(text, "For future Claude", 6)
    activity = section(text, "Recent Activity", 5)
    decisions = section(text, "Key Decisions", 5)

    parts = [f"# [Obsidian] Contexto do vault para o projeto {note.stem}",
             f"(fonte: {note}; atualizado pelos hooks de sessão)"]
    if preamble:
        parts += ["", *preamble]
    if activity:
        parts += ["", "## Recent Activity (mais recentes)", *activity]
    if decisions:
        parts += ["", "## Key Decisions", *decisions]
    out = "\n".join(parts)
    if len(out) > MAX_OUT:
        out = out[:MAX_OUT] + "\n[...truncado]"
    print(out)
    return 0


if __name__ == "__main__":
    sys.exit(main())
