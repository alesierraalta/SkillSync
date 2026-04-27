# Project Intro: Skillsync TUI

Welcome to the **Skillsync TUI** project. This tool is designed to be the central hub for managing, discovering, and testing AI agent skills within your development environment.

## Vision

The Skillsync TUI provides a visual and interactive interface to:
- **Discover**: Automatically find skills defined in `.agents`, `.claude`, `.gemini`, and other convention-based directories.
- **Manage**: Install, update, and organize skills with a clear overview of their metadata and instructions.
- **Sandbox**: Run and test skills in an isolated environment to ensure safety and predictability.
- **Sync**: Keep your `AGENTS.md` and other registry files perfectly synchronized with your actual skill files.

## Core Value

In a world of fragmented AI instructions and local agent configurations, Skillsync TUI brings **order** and **visibility**. It ensures that both human developers and AI agents have a "source of truth" for the capabilities available in the project.

## High-Level Workflow

1. **Scan**: On startup, the TUI scans the project for any `SKILL.md` files.
2. **Interact**: Browse skills, view their documentation, and check their status.
3. **Execute**: Trigger syncs or sandbox tests directly from the interface.
4. **Persist**: Any changes made are persisted back to the filesystem and memory (Engram).
