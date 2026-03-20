# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

`log-sage` is the **AI scaffolding and documentation hub** for the LogSage project. It contains no application source code — only AI collaboration files tracked outside of the main source repository (`.ai/` is excluded from git via `.gitignore`).

LogSage is a **developer CLI tool and analysis engine that parses logs and identifies the most likely root causes of system failures.**

## Session initialization

At the start of every session, invoke the appropriate custom slash command based on
the task type. See the Custom slash commands table below.
Full codebase reference: `.ai/references/repo-map.md`

## User shorthand

| What the user says | What it means |
|---|---|
| "review the handoff" / "review latest handoff" / "review current.md" | Read `.ai/handoffs/current.md` and evaluate the completed work |
| "evaluate the pr review" / "review the pr review" | Read `.ai/handoffs/coderabbit-eval.md` — invoke `/orchestrator` to evaluate |
| "run the prompt" / "execute the prompt" | Read and execute `.ai/handoffs/prompt.md` |

## Custom slash commands

When the user references any of these commands, read the corresponding file and execute it immediately — do not wait to be told to "invoke" it.

| Command | File | Action |
|---|---|---|
| `/orchestrator` | `.claude/commands/orchestrator.md` | Load context and act as Orchestrator |
| `/analyzer` | `.claude/commands/analyzer.md` | Load context and act as Analyzer |
| `/implementor` | `.claude/commands/implementor.md` | Load context and act as Implementor |
| `/implement` | `.claude/commands/implement.md` | Read and execute `.ai/handoffs/prompt.md` |
| `/findings` | `.claude/commands/findings.md` | Read and present `.ai/references/analysis.md` |
