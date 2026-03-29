# JANE Autonomous Execution Research

This document outlines high-momentum Go-based repositories identified to empower JANE with active, autonomous task execution, moving from a passive "answering" system to an active "doing" agent.

## 1. Advanced Sandboxing & Execution Isolation
To execute untrusted or complex scripts dynamically without compromising the host environment, JANE requires robust sandboxing.
*   **[gVisor](https://github.com/google/gvisor)** (17.9K ⭐): Provides deep container-level isolation by intercepting system calls. This would allow JANE to safely spin up complete, isolated execution environments for running unknown code or complex multi-step build pipelines.
*   **[go-landlock](https://github.com/landlock-lsm/go-landlock)**: A Go wrapper for the Linux Landlock LSM. This is a lightweight, kernel-level sandboxing feature that can be directly integrated into JANE's existing `exec` and `shell` tools to restrict file system access on a granular level, far more securely than naive path checking.

## 2. High-level Automation & CI Orchestration
For JANE to perform multi-step, system-level tasks automatically, it needs an orchestration layer.
*   **[Dagger](https://github.com/dagger/dagger)** (15.5K ⭐): Integrating Dagger's Go SDK would allow JANE to natively write and execute directed acyclic graphs (DAGs) of containerized CI/CD tasks. This provides an incredible leap in autonomy, letting JANE handle complex build, test, and deployment workflows natively.

## 3. Web Automation & Interaction
*   **[go-rod](https://github.com/go-rod/rod)** (6.8K ⭐): A robust Chrome DevTools Protocol driver. Compared to the current Playwright-based `browser_action` tool, `go-rod` is more lightweight and Go-native. It can be integrated to allow JANE to write continuous, sandboxed automation scripts rather than invoking single-step click/type commands, dramatically reducing round-trip latency to the LLM during complex web tasks.

## 4. Enhanced Contextual Scripting (Implemented)
*   **[Yaegi](https://github.com/traefik/yaegi)** (5K ⭐): Already utilized by JANE's `go_eval` tool, we have enhanced this integration by injecting contextual bindings (`Workspace`, `HTTPClient`, `BrowserActionTool`, and `Send` messaging callbacks). This fulfills the "Autonomous Skill Acquisition" directive by allowing JANE to write one complex, sandboxed script that orchestrates multiple internal systems simultaneously.
