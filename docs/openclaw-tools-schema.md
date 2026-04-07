# OpenClaw Tools Schema Reference

> 从 OpenClaw 客户端实际请求中提取，用于 sanitize 映射维护。
> 最后更新: 2026-04-07

## 工具列表总览

| 原始名称 | 映射名称 | 类型 |
|---|---|---|
| `read` | `Kx7` | 交叉 (Claude Code 也有) |
| `edit` | `Rq9` | 交叉 |
| `write` | `Mv3` | 交叉 |
| `exec` | `Tn4` | 交叉 |
| `web_search` | `Jb6` | 交叉 |
| `web_fetch` | `Wp2` | 交叉 |
| `process` | `Uf8` | 非交叉 |
| `canvas` | `Zd1` | 非交叉 |
| `nodes` | `Hy5` | 非交叉 |
| `cron` | `Oa0` | 非交叉 |
| `message` | `Lc3` | 非交叉 |
| `tts` | `Bg7` | 非交叉 |
| `gateway` | `Xe9` | 非交叉 |
| `agents_list` | `Fs2` | 非交叉 |
| `sessions_list` | `Qi4` | 非交叉 |
| `sessions_history` | `Nv6` | 非交叉 |
| `sessions_send` | `Pw1` | 非交叉 |
| `sessions_yield` | `Ek8` | 非交叉 |
| `sessions_spawn` | `Gm5` | 非交叉 |
| `subagents` | `Dr0` | 非交叉 |
| `session_status` | `Aj7` | 非交叉 |
| `browser` | `Vc3` | 非交叉 |
| `memory_search` | `Yl6` | 非交叉 |
| `memory_get` | `Sh9` | 非交叉 |

> 映射名称为随机无意义代号，与原始功能无语义关联。未命中映射表的工具自动加 `Zx` 前缀。

---

## 各工具详细 Schema

### read

```json
{
  "description": "Read the contents of a file. Supports text files and images (jpg, png, gif, webp). Images are sent as attachments. For text files, output is truncated to 2000 lines or 50KB (whichever is hit first). Use offset/limit for large files. When you need the full file, continue with offset until complete.",
  "name": "read",
  "parametersJsonSchema": {
    "properties": {
      "file": { "description": "Path to the file to read (relative or absolute)", "type": "string" },
      "filePath": { "description": "Path to the file to read (relative or absolute)", "type": "string" },
      "file_path": { "description": "Path to the file to read (relative or absolute)", "type": "string" },
      "limit": { "description": "Maximum number of lines to read", "type": "number" },
      "offset": { "description": "Line number to start reading from (1-indexed)", "type": "number" },
      "path": { "description": "Path to the file to read (relative or absolute)", "type": "string" }
    },
    "required": [],
    "type": "object"
  }
}
```

### edit

```json
{
  "description": "Edit a single file using exact text replacement. Every edits[].oldText must match a unique, non-overlapping region of the original file. If two changes affect the same block or nearby lines, merge them into one edit instead of emitting overlapping edits. Do not include large unchanged regions just to connect distant changes.",
  "name": "edit",
  "parametersJsonSchema": {
    "additionalProperties": false,
    "properties": {
      "edits": {
        "description": "One or more targeted replacements. Each edit is matched against the original file, not incrementally. Do not include overlapping or nested edits. If two changes touch the same block or nearby lines, merge them into one edit instead.",
        "items": {
          "additionalProperties": false,
          "properties": {
            "newText": { "description": "Replacement text for this targeted edit.", "type": "string" },
            "oldText": { "description": "Exact text for one targeted replacement. It must be unique in the original file and must not overlap with any other edits[].oldText in the same call.", "type": "string" }
          },
          "required": ["oldText", "newText"],
          "type": "object"
        },
        "type": "array"
      },
      "file": { "description": "Path to the file to edit (relative or absolute)", "type": "string" },
      "filePath": { "description": "Path to the file to edit (relative or absolute)", "type": "string" },
      "file_path": { "description": "Path to the file to edit (relative or absolute)", "type": "string" },
      "path": { "description": "Path to the file to edit (relative or absolute)", "type": "string" }
    },
    "required": ["edits"],
    "type": "object"
  }
}
```

### write

```json
{
  "description": "Write content to a file. Creates the file if it doesn't exist, overwrites if it does. Automatically creates parent directories.",
  "name": "write",
  "parametersJsonSchema": {
    "properties": {
      "content": { "description": "Content to write to the file", "type": "string" },
      "file": { "description": "Path to the file to write (relative or absolute)", "type": "string" },
      "filePath": { "description": "Path to the file to write (relative or absolute)", "type": "string" },
      "file_path": { "description": "Path to the file to write (relative or absolute)", "type": "string" },
      "path": { "description": "Path to the file to write (relative or absolute)", "type": "string" }
    },
    "required": ["content"],
    "type": "object"
  }
}
```

### exec

```json
{
  "description": "Execute shell commands with background continuation. Use yieldMs/background to continue later via process tool. Use pty=true for TTY-required commands (terminal UIs, coding agents).",
  "name": "exec",
  "parametersJsonSchema": {
    "properties": {
      "ask": { "description": "Exec ask mode (off|on-miss|always).", "type": "string" },
      "background": { "description": "Run in background immediately", "type": "boolean" },
      "command": { "description": "Shell command to execute", "type": "string" },
      "elevated": { "description": "Run on the host with elevated permissions (if allowed)", "type": "boolean" },
      "env": { "patternProperties": { "^(.*)$": { "type": "string" } }, "type": "object" },
      "host": { "description": "Exec host/target (auto|sandbox|gateway|node).", "type": "string" },
      "node": { "description": "Node id/name for host=node.", "type": "string" },
      "pty": { "description": "Run in a pseudo-terminal (PTY) when available (TTY-required CLIs, coding agents)", "type": "boolean" },
      "security": { "description": "Exec security mode (deny|allowlist|full).", "type": "string" },
      "timeout": { "description": "Timeout in seconds (optional, kills process on expiry)", "type": "number" },
      "workdir": { "description": "Working directory (defaults to cwd)", "type": "string" },
      "yieldMs": { "description": "Milliseconds to wait before backgrounding (default 10000)", "type": "number" }
    },
    "required": ["command"],
    "type": "object"
  }
}
```

### process

```json
{
  "description": "Manage running exec sessions: list, poll, log, write, send-keys, submit, paste, kill.",
  "name": "process",
  "parametersJsonSchema": {
    "properties": {
      "action": { "description": "Process action", "type": "string" },
      "bracketed": { "description": "Wrap paste in bracketed mode", "type": "boolean" },
      "data": { "description": "Data to write for write", "type": "string" },
      "eof": { "description": "Close stdin after write", "type": "boolean" },
      "hex": { "description": "Hex bytes to send for send-keys", "items": { "type": "string" }, "type": "array" },
      "keys": { "description": "Key tokens to send for send-keys", "items": { "type": "string" }, "type": "array" },
      "limit": { "description": "Log length", "type": "number" },
      "literal": { "description": "Literal string for send-keys", "type": "string" },
      "offset": { "description": "Log offset", "type": "number" },
      "sessionId": { "description": "Session id for actions other than list", "type": "string" },
      "text": { "description": "Text to paste for paste", "type": "string" },
      "timeout": { "description": "For poll: wait up to this many milliseconds before returning", "minimum": 0, "type": "number" }
    },
    "required": ["action"],
    "type": "object"
  }
}
```

### canvas

```json
{
  "description": "Control node canvases (present/hide/navigate/eval/snapshot/A2UI). Use snapshot to capture the rendered UI.",
  "name": "canvas",
  "parametersJsonSchema": {
    "properties": {
      "action": { "enum": ["present", "hide", "navigate", "eval", "snapshot", "a2ui_push", "a2ui_reset"], "type": "string" },
      "delayMs": { "type": "number" },
      "gatewayToken": { "type": "string" },
      "gatewayUrl": { "type": "string" },
      "height": { "type": "number" },
      "javaScript": { "type": "string" },
      "jsonl": { "type": "string" },
      "jsonlPath": { "type": "string" },
      "maxWidth": { "type": "number" },
      "node": { "type": "string" },
      "outputFormat": { "enum": ["png", "jpg", "jpeg"], "type": "string" },
      "quality": { "type": "number" },
      "target": { "type": "string" },
      "timeoutMs": { "type": "number" },
      "url": { "type": "string" },
      "width": { "type": "number" },
      "x": { "type": "number" },
      "y": { "type": "number" }
    },
    "required": ["action"],
    "type": "object"
  }
}
```

### message

```json
{
  "description": "Send, delete, and manage messages via channel plugins. Supports actions: send, broadcast.",
  "name": "message",
  "parametersJsonSchema": {
    "properties": {
      "accountId": { "type": "string" },
      "action": { "enum": ["send", "broadcast"], "type": "string" },
      "channel": { "type": "string" },
      "message": { "type": "string" },
      "target": { "description": "Target channel/user id or name.", "type": "string" },
      "targets": { "items": { "type": "string" }, "type": "array" },
      "replyTo": { "type": "string" },
      "silent": { "type": "boolean" }
    },
    "required": ["action"],
    "type": "object"
  }
}
```

> Note: message tool 的完整 schema 参数非常多（Discord/Telegram/Lark 等全量参数），
> 这里只列出核心参数，完整版见原始请求 dump。

### tts

```json
{
  "description": "Convert text to speech. Audio is delivered automatically from the tool result — reply with NO_REPLY after a successful call to avoid duplicate messages.",
  "name": "tts",
  "parametersJsonSchema": {
    "properties": {
      "channel": { "description": "Optional channel id to pick output format (e.g. telegram).", "type": "string" },
      "text": { "description": "Text to convert to speech.", "type": "string" }
    },
    "required": ["text"],
    "type": "object"
  }
}
```

### agents_list

```json
{
  "description": "List OpenClaw agent ids you can target with sessions_spawn when runtime=\"subagent\" (based on subagent allowlists).",
  "name": "agents_list",
  "parametersJsonSchema": { "properties": {}, "type": "object" }
}
```

### sessions_list

```json
{
  "description": "List sessions with optional filters and last messages.",
  "name": "sessions_list",
  "parametersJsonSchema": {
    "properties": {
      "activeMinutes": { "minimum": 1, "type": "number" },
      "kinds": { "items": { "type": "string" }, "type": "array" },
      "limit": { "minimum": 1, "type": "number" },
      "messageLimit": { "minimum": 0, "type": "number" }
    },
    "type": "object"
  }
}
```

### sessions_history

```json
{
  "description": "Fetch message history for a session.",
  "name": "sessions_history",
  "parametersJsonSchema": {
    "properties": {
      "includeTools": { "type": "boolean" },
      "limit": { "minimum": 1, "type": "number" },
      "sessionKey": { "type": "string" }
    },
    "required": ["sessionKey"],
    "type": "object"
  }
}
```

### sessions_send

```json
{
  "description": "Send a message into another session. Use sessionKey or label to identify the target.",
  "name": "sessions_send",
  "parametersJsonSchema": {
    "properties": {
      "agentId": { "maxLength": 64, "minLength": 1, "type": "string" },
      "label": { "maxLength": 512, "minLength": 1, "type": "string" },
      "message": { "type": "string" },
      "sessionKey": { "type": "string" },
      "timeoutSeconds": { "minimum": 0, "type": "number" }
    },
    "required": ["message"],
    "type": "object"
  }
}
```

### sessions_yield

```json
{
  "description": "End your current turn. Use after spawning subagents to receive their results as the next message.",
  "name": "sessions_yield",
  "parametersJsonSchema": {
    "properties": {
      "message": { "type": "string" }
    },
    "type": "object"
  }
}
```

### sessions_spawn

```json
{
  "description": "Spawn an isolated session (runtime=\"subagent\" or runtime=\"acp\"). mode=\"run\" is one-shot and mode=\"session\" is persistent/thread-bound. Subagents inherit the parent workspace directory automatically.",
  "name": "sessions_spawn",
  "parametersJsonSchema": {
    "properties": {
      "agentId": { "type": "string" },
      "cleanup": { "enum": ["delete", "keep"], "type": "string" },
      "cwd": { "type": "string" },
      "label": { "type": "string" },
      "mode": { "enum": ["run", "session"], "type": "string" },
      "model": { "type": "string" },
      "runtime": { "enum": ["subagent", "acp"], "type": "string" },
      "task": { "type": "string" },
      "timeoutSeconds": { "minimum": 0, "type": "number" }
    },
    "required": ["task"],
    "type": "object"
  }
}
```

> Note: sessions_spawn 完整 schema 包含 attachments/attachAs/thinking 等更多参数，这里列出核心参数。

### subagents

```json
{
  "description": "List, kill, or steer spawned sub-agents for this requester session. Use this for sub-agent orchestration.",
  "name": "subagents",
  "parametersJsonSchema": {
    "properties": {
      "action": { "enum": ["list", "kill", "steer"], "type": "string" },
      "message": { "type": "string" },
      "recentMinutes": { "minimum": 1, "type": "number" },
      "target": { "type": "string" }
    },
    "type": "object"
  }
}
```

### session_status

```json
{
  "description": "Show a /status-equivalent session status card (usage + time + cost when available), including linked background task context when present. Use for model-use questions. Optional: set per-session model override (model=default resets overrides).",
  "name": "session_status",
  "parametersJsonSchema": {
    "properties": {
      "model": { "type": "string" },
      "sessionKey": { "type": "string" }
    },
    "type": "object"
  }
}
```

### web_search

```json
{
  "description": "Search the web using DuckDuckGo. Returns titles, URLs, and snippets with no API key required.",
  "name": "web_search",
  "parametersJsonSchema": {
    "additionalProperties": false,
    "properties": {
      "count": { "description": "Number of results to return (1-10).", "maximum": 10, "minimum": 1, "type": "number" },
      "query": { "description": "Search query string.", "type": "string" },
      "region": { "description": "Optional DuckDuckGo region code such as us-en, uk-en, or de-de.", "type": "string" },
      "safeSearch": { "description": "SafeSearch level: strict, moderate, or off.", "type": "string" }
    },
    "required": ["query"],
    "type": "object"
  }
}
```

### web_fetch

```json
{
  "description": "Fetch and extract readable content from a URL (HTML → markdown/text). Use for lightweight page access without browser automation.",
  "name": "web_fetch",
  "parametersJsonSchema": {
    "properties": {
      "extractMode": { "default": "markdown", "description": "Extraction mode (\"markdown\" or \"text\").", "enum": ["markdown", "text"], "type": "string" },
      "maxChars": { "description": "Maximum characters to return (truncates when exceeded).", "minimum": 100, "type": "number" },
      "url": { "description": "HTTP or HTTPS URL to fetch.", "type": "string" }
    },
    "required": ["url"],
    "type": "object"
  }
}
```

### browser

```json
{
  "description": "Control the browser via OpenClaw's browser control server (status/start/stop/profiles/tabs/open/snapshot/screenshot/actions).",
  "name": "browser",
  "parametersJsonSchema": {
    "properties": {
      "action": { "enum": ["status", "start", "stop", "profiles", "tabs", "open", "focus", "close", "snapshot", "screenshot", "navigate", "console", "pdf", "upload", "dialog", "act"], "type": "string" },
      "compact": { "type": "boolean" },
      "element": { "type": "string" },
      "frame": { "type": "string" },
      "fullPage": { "type": "boolean" },
      "kind": { "enum": ["click", "type", "press", "hover", "drag", "select", "fill", "resize", "wait", "evaluate", "close"], "type": "string" },
      "labels": { "type": "boolean" },
      "ref": { "type": "string" },
      "selector": { "type": "string" },
      "url": { "type": "string" }
    },
    "required": ["action"],
    "type": "object"
  }
}
```

> Note: browser 完整 schema 参数非常多，这里列出核心参数。

### memory_search

```json
{
  "description": "Mandatory recall step: semantically search MEMORY.md + memory/*.md (and optional session transcripts) before answering questions about prior work, decisions, dates, people, preferences, or todos; returns top snippets with path + lines. If response has disabled=true, memory retrieval is unavailable and should be surfaced to the user.",
  "name": "memory_search",
  "parametersJsonSchema": {
    "properties": {
      "maxResults": { "type": "number" },
      "minScore": { "type": "number" },
      "query": { "type": "string" }
    },
    "required": ["query"],
    "type": "object"
  }
}
```

### memory_get

```json
{
  "description": "Safe snippet read from MEMORY.md or memory/*.md with optional from/lines; use after memory_search to pull only the needed lines and keep context small.",
  "name": "memory_get",
  "parametersJsonSchema": {
    "properties": {
      "from": { "type": "number" },
      "lines": { "type": "number" },
      "path": { "type": "string" }
    },
    "required": ["path"],
    "type": "object"
  }
}
```
