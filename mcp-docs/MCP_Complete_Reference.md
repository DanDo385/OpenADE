# Model Context Protocol (MCP) — Complete Reference Summary
> Compiled from modelcontextprotocol.io and GitHub · March 2026

---

## Table of Contents

1. [What is MCP?](#1-what-is-mcp)
2. [Architecture & Core Concepts](#2-architecture--core-concepts)
3. [Documentation Structure (94+ Pages)](#3-documentation-structure)
4. [Getting Started & Development](#4-getting-started--development)
5. [Extensions](#5-extensions)
6. [Specification (2025-11-25)](#6-specification-2025-11-25)
7. [MCP Registry](#7-mcp-registry)
8. [Community & Governance](#8-community--governance)
9. [GitHub: Official MCP Servers](#9-github-official-mcp-servers)
10. [GitHub: Community MCP Server Lists](#10-github-community-mcp-server-lists)
11. [Categorized MCP Server Directory](#11-categorized-mcp-server-directory)

---

## 1. What is MCP?

**Model Context Protocol (MCP)** is an open-source standard for connecting AI applications to external systems — data sources, tools, APIs, and workflows. Think of it as the "USB-C port for AI": a universal, standardized plug that lets any LLM-powered app connect to any external capability.

MCP is hosted by **The Linux Foundation** and is fully open-source.

**Key use cases:**
- Agents accessing Google Calendar, Notion, or Slack as a personalized AI assistant
- Claude Code generating entire web apps from a Figma design
- Enterprise chatbots querying multiple databases via natural language
- AI models creating 3D designs in Blender and sending them to a 3D printer

**Who benefits:**
- **Developers** — Reduced complexity when building or integrating AI applications
- **AI apps/agents** — Access to a rich ecosystem of tools and data sources
- **End-users** — More capable AI that can act on your data and take real-world actions

**Supported clients:** Claude, ChatGPT, Visual Studio Code (Copilot), Cursor, Goose, Postman, MCPJam, and many others.

---

## 2. Architecture & Core Concepts

### Participants

MCP follows a **client-server architecture**:

| Role | Description |
|---|---|
| **MCP Host** | The AI application (e.g., Claude Desktop, VS Code) that manages MCP clients |
| **MCP Client** | A component that maintains a dedicated connection to one MCP server |
| **MCP Server** | A program that provides context (tools, resources, prompts) to clients |

A single host can maintain connections to multiple servers simultaneously, each through its own dedicated client.

### Two Layers

**Data Layer** — The inner layer. JSON-RPC 2.0 based protocol defining:
- Lifecycle management (initialization, capability negotiation, termination)
- Server primitives (Tools, Resources, Prompts)
- Client primitives (Sampling, Elicitation, Logging)
- Utility features (Notifications, Progress tracking, Tasks)

**Transport Layer** — The outer layer. Two mechanisms:
- **Stdio transport** — Standard input/output for local process-to-process communication. Zero network overhead.
- **Streamable HTTP transport** — HTTP POST for client-to-server, optional Server-Sent Events for streaming. Supports bearer tokens, API keys, OAuth.

### Core Primitives (Server → Client)

| Primitive | Purpose | Example |
|---|---|---|
| **Tools** | Executable functions AI can invoke | File operations, API calls, database queries |
| **Resources** | Data sources for contextual information | File contents, DB records, API responses |
| **Prompts** | Reusable interaction templates | System prompts, few-shot examples |

### Client Primitives (Server can request from Client)

| Primitive | Purpose |
|---|---|
| **Sampling** | Server requests LLM completions from the client's AI app |
| **Elicitation** | Server requests additional input/confirmation from user |
| **Logging** | Server sends debug/monitoring messages to client |

### Experimental: Tasks
Durable execution wrappers for deferred results and status tracking — useful for expensive computations, batch processing, and multi-step workflows.

### Lifecycle Example (JSON-RPC 2.0)

```json
// Client → Server: Initialize
{ "jsonrpc": "2.0", "id": 1, "method": "initialize",
  "params": { "protocolVersion": "2025-06-18", "capabilities": { "elicitation": {} } } }

// Server → Client: Initialized response
{ "jsonrpc": "2.0", "id": 1, "result": {
  "capabilities": { "tools": { "listChanged": true }, "resources": {} } } }

// Client → Server: List tools
{ "jsonrpc": "2.0", "id": 2, "method": "tools/list" }

// Client → Server: Call a tool
{ "jsonrpc": "2.0", "id": 3, "method": "tools/call",
  "params": { "name": "weather_current", "arguments": { "location": "San Francisco" } } }

// Server → Client: Real-time notification (no response expected)
{ "jsonrpc": "2.0", "method": "notifications/tools/list_changed" }
```

---

## 3. Documentation Structure

The full documentation at `modelcontextprotocol.io` contains **94+ pages** organized as follows:

### Getting Started
- What is MCP? (`/docs/getting-started/intro`)

### Learning
- Architecture overview (`/docs/learn/architecture`)
- Understanding MCP clients (`/docs/learn/client-concepts`)
- Understanding MCP servers (`/docs/learn/server-concepts`)

### Development
- Build an MCP client (`/docs/develop/build-client`)
- Build an MCP server (`/docs/develop/build-server`)
- Connect to local MCP servers (`/docs/develop/connect-local-servers`)
- Connect to remote MCP servers (`/docs/develop/connect-remote-servers`)
- Roadmap (`/development/roadmap`)

### SDKs & Tools
- SDKs overview (`/docs/sdk`)
- MCP Inspector (`/docs/tools/inspector`)
- Example clients (`/clients`)
- Example servers (`/examples`)

### Security & Authorization
- Understanding Authorization in MCP
- Security Best Practices
- OAuth Client Credentials
- Enterprise-Managed Authorization
- Authorization Extensions overview

### Extensions
- Extensions Overview
- Extension Support Matrix (client compatibility)
- MCP Apps Overview
- Build an MCP App

### Community & Governance
- Antitrust Policy
- Contributor Communication
- Contributing to MCP
- Design Principles
- Governance and Stewardship
- SDK Tiering System
- Working and Interest Groups
- SEP Guidelines + 25+ active SEP Proposals

### Registry
- About, Authentication, FAQ, GitHub Actions Automation
- Moderation Policy, Package Types, Quickstart
- Aggregators, Remote Servers, Terms of Service, Versioning

### Specification (2025-11-25)
- Architecture, Basic Protocol, Authorization, Lifecycle, Transports
- Utilities: Cancellation, Ping, Progress, Tasks
- Client: Elicitation, Roots, Sampling
- Server: Prompts, Resources, Tools
- Server Utilities: Completion, Logging, Pagination
- Schema Reference, Changelog

---

## 4. Getting Started & Development

### SDKs Available

MCP maintains official SDKs in 10+ languages:

| Language | Repository |
|---|---|
| TypeScript/JS | `modelcontextprotocol/typescript-sdk` |
| Python | `modelcontextprotocol/python-sdk` |
| Go | `modelcontextprotocol/go-sdk` |
| Java | `modelcontextprotocol/java-sdk` |
| Kotlin | `modelcontextprotocol/kotlin-sdk` |
| C# | `modelcontextprotocol/csharp-sdk` |
| Swift | `modelcontextprotocol/swift-sdk` |
| Rust | `modelcontextprotocol/rust-sdk` |
| Ruby | `modelcontextprotocol/ruby-sdk` |
| PHP | `modelcontextprotocol/php-sdk` |

### Building a Server (Quick Overview)

1. Choose your SDK (Python FastMCP or TypeScript MCP SDK are most popular)
2. Define your Tools, Resources, and/or Prompts
3. Set up the transport (stdio for local, Streamable HTTP for remote)
4. Test with the **MCP Inspector** (`@modelcontextprotocol/inspector`)
5. Connect to a host (Claude Desktop, VS Code, etc.)
6. Publish to the MCP Registry

### MCP Inspector
A development tool for interactively testing and debugging MCP servers — inspect tools, resources, prompts, and real-time message flows.

---

## 5. Extensions

### MCP Apps

MCP Apps are **interactive UI applications that render inside MCP host clients** (like Claude Desktop or VS Code). They extend MCP beyond text responses by allowing servers to return interactive HTML interfaces — dashboards, forms, data visualizations — embedded directly in the conversation.

**Why not just build a web app?**
- **Context preservation** — The app lives inside the conversation
- **Bidirectional data flow** — Can call any MCP tool on the server; host can push fresh results
- **Host capability integration** — Can delegate actions to the host's existing connected tools
- **Security** — Sandboxed iframe, cannot escape or access parent page data

**How it works:**
1. Tool description includes `_meta.ui.resourceUri` pointing to a `ui://` resource
2. Host preloads the UI resource (HTML + JS + CSS)
3. Host renders it in a sandboxed iframe
4. App and host communicate via JSON-RPC over `postMessage`

**Supported clients:** Claude, Claude Desktop, VS Code GitHub Copilot, Goose, Postman, MCPJam

**Use cases:**
- Interactive data exploration (maps, charts with drill-down)
- Complex configuration forms
- Rich media viewers (PDF, 3D models, video)
- Real-time monitoring dashboards
- Multi-step approval workflows

**Framework starters available:** React, Vue, Svelte, Preact, Solid, Vanilla JavaScript

**Example app servers:**
- 3D/Visualization: CesiumJS globe, Three.js scenes, ShaderToy effects
- Data: Cohort heatmaps, customer segmentation, wiki explorer
- Business: Scenario modeler, budget allocator
- Media: PDF viewer, video resources, sheet music, text-to-speech
- Utilities: QR codes, system monitor, speech-to-text transcript

### Authorization Extensions
- OAuth Client Credentials flow
- Enterprise-Managed Authorization
- Bearer tokens, API keys, custom headers

### Extension Support Matrix
Not all clients support all extensions. Check `/extensions/client-matrix` for full compatibility.

---

## 6. Specification (2025-11-25)

**Current version:** `2025-11-25` (string-based date format, backwards-compatible incremental changes)

**Version negotiation** happens at initialization. Clients and servers may support multiple versions simultaneously but must agree on one for the session.

### Revision States
- **Draft** — In-progress, not yet ready for use
- **Current** — Ready for production use (`2025-11-25`)
- **Final** — Past, complete, frozen specifications

### Specification Sections

**Architecture** — Overall design, participants, layers

**Basic Protocol:**
- Overview of message format
- Authorization — OAuth and other auth mechanisms
- Lifecycle — Init, negotiation, termination
- Transports — Stdio and Streamable HTTP

**Utilities:**
- Cancellation — Aborting in-flight requests
- Ping — Keep-alive / health checks
- Progress — Tracking long-running operations
- Tasks (Experimental) — Durable async execution

**Client Features:**
- Elicitation — Server asks user for input
- Roots — File system access boundaries
- Sampling — Server requests LLM completion

**Server Features:**
- Prompts — Template management
- Resources — Data access and retrieval
- Tools — Function definition and invocation

**Server Utilities:**
- Completion — Auto-complete suggestions
- Logging — Structured log forwarding
- Pagination — Large result set handling

**Schema:** Full JSON Schema at `/specification/2025-11-25/schema.md`

### Active SEPs (Specification Enhancement Proposals)
25+ active proposals including SEP-414, SEP-932, SEP-973, SEP-985, SEP-986, SEP-990, SEP-991, SEP-994, SEP-1024, SEP-1034, SEP-1036, SEP-1046, SEP-1302, SEP-1303, SEP-1319, SEP-1330, SEP-1577, SEP-1613, SEP-1686, SEP-1699, SEP-1730, SEP-1850, SEP-1865, SEP-2085, SEP-2133

---

## 7. MCP Registry

> **Status:** Currently in preview (launched September 2025). GA release coming.

The MCP Registry is the **official centralized metadata repository** for publicly accessible MCP servers, backed by Anthropic, GitHub, PulseMCP, and Microsoft.

### What It Provides
- A single place to publish server metadata
- Namespace management via DNS verification
- A REST API for discovery by clients and aggregators
- Standardized `server.json` format with: unique name, location (npm/Docker/URL), execution instructions, capabilities

### Server Name Format
Reverse DNS style: `io.github.username/server-name` or `com.example/server-name`

### Ecosystem Relationships
- **Package registries** (npm, PyPI, Docker Hub) — host the actual code; MCP Registry hosts *metadata* pointing to them
- **Downstream aggregators** — consume the Registry API hourly to build marketplaces and curated lists
- **MCP hosts** — should consume aggregators, not the Registry directly

### Publishing Requirements
- Server must be publicly installable (npm, PyPI, Docker Hub) OR publicly accessible (remote URL)
- Private servers are NOT supported (host your own private registry instead)
- Namespace ownership verified via GitHub account or DNS/HTTP challenge

### Security Model
- Namespace authentication prevents spoofing
- Code scanning delegated to package registries and aggregators
- Manual takedown available for spam/malicious servers
- Future: rate limiting, AI-based spam detection, community reporting

### Registry Documentation Pages
- Quickstart, Authentication, FAQ, GitHub Actions Automation
- Package Types, Moderation Policy, Versioning, Terms of Service
- Aggregators guide, Remote Servers guide

---

## 8. Community & Governance

### Project Roles
| Role | Responsibility |
|---|---|
| **Contributors** | File issues, submit PRs, participate in discussions |
| **Maintainers** | Steward specific areas (SDKs, docs, Working Groups) |
| **Core Maintainers** | Guide project direction, review SEPs, oversee specification |

### Contributing
- **Small changes** — Submit a PR directly (typos, docs, examples, minor fixes)
- **Major changes** — Follow the SEP (Specification Enhancement Proposal) process
  1. Validate idea in Discord or Interest Group
  2. Build a prototype
  3. Find a Core Maintainer sponsor
  4. Write and submit the SEP

**AI contributions** are welcome. Just note in your PR how you used AI (drafting, code gen, brainstorming, etc.) and confirm you understand and can stand behind the changes.

### Communication Channels
- **Discord** — Real-time contributor discussion (`discord.gg/6CSzBmMkjX`)
  - Channels: `#typescript-sdk-dev`, `#python-sdk-dev`, `#auth-wg`, `#server-identity-wg`, etc.
- **GitHub Discussions** — Feature requests, questions, roadmap, pre-issue proposals
- **GitHub Issues** — Bug reports, well-defined tasks ready to implement

### Licenses
- Code and specifications: **Apache License 2.0**
- Documentation (excluding specs): **CC-BY 4.0**

### Design Principles
(Documented at `/community/design-principles`)

### SDK Tiering System
MCP maintains a tiering classification for SDK quality and support levels across the 10+ language SDKs.

### Working & Interest Groups
Active working groups focused on specific areas of the protocol (auth, server identity, etc.)

---

## 9. GitHub: Official MCP Servers

**Repository:** [github.com/modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

> These are **reference/educational implementations**, not production-ready solutions.

### Active Reference Servers

| Server | Description |
|---|---|
| **Everything** | Comprehensive reference with prompts, resources, and tools |
| **Fetch** | Web content fetching and conversion optimized for LLM usage |
| **Filesystem** | Secure file operations with configurable access controls |
| **Git** | Tools to read, search, and manipulate Git repositories |
| **Memory** | Knowledge graph-based persistent memory system |
| **Sequential Thinking** | Dynamic problem-solving through structured thought sequences |
| **Time** | Time and timezone conversion capabilities |

### Archived Reference Servers
(Moved to `servers-archived` repo): AWS KB Retrieval, Brave Search, EverArt, GitHub, GitLab, Google Drive, Google Maps, PostgreSQL, Puppeteer, Redis, Sentry, Slack, SQLite

---

## 10. GitHub: Community MCP Server Lists

### Official Curated Lists

| Repository | Focus |
|---|---|
| [punkpeye/awesome-mcp-servers](https://github.com/punkpeye/awesome-mcp-servers) | Large general collection |
| [wong2/awesome-mcp-servers](https://github.com/wong2/awesome-mcp-servers) | Curated with official integrations |
| [appcypher/awesome-mcp-servers](https://github.com/appcypher/awesome-mcp-servers) | Production & experimental focus |
| [rohitg00/awesome-devops-mcp-servers](https://github.com/rohitg00/awesome-devops-mcp-servers) | DevOps-specific servers |
| [ever-works/awesome-mcp-servers](https://github.com/ever-works/awesome-mcp-servers) | Broad production-ready list |
| [patriksimek/awesome-mcp-servers-2](https://github.com/patriksimek/awesome-mcp-servers-2) | Search, social, browser automation |
| [PipedreamHQ/awesome-mcp-servers](https://github.com/PipedreamHQ/awesome-mcp-servers) | Pipedream-focused collection |
| [TensorBlock/awesome-mcp-servers](https://github.com/TensorBlock/awesome-mcp-servers) | Comprehensive collection |
| [tolkonepiu/best-of-mcp-servers](https://github.com/tolkonepiu/best-of-mcp-servers) | Weekly ranked list |
| [esc5221/awesome-awesome-mcp-servers](https://github.com/esc5221/awesome-awesome-mcp-servers) | Meta-index of all awesome lists |

**Web Directory:** [mcpservers.org](https://mcpservers.org) — Search and discover servers and clients

**Scale:** As of mid-2025, the ecosystem tracks **7,260+ MCP servers**

### Novel Architectures
- **[MicroMCP](https://github.com/mabualzait/MicroMCP)** — Composes many single-purpose MCP servers behind a lightweight gateway with security isolation, independent deployability, unified tool/resource/prompt discovery, policy enforcement, and audit logging
- **[Mobile MCP](https://github.com/mobile-next/mobile-mcp)** — Platform-agnostic mobile automation (iOS and Android, emulators and real devices)

---

## 11. Categorized MCP Server Directory

### Infrastructure & Cloud
- AWS CDK, cost analysis, documentation, Bedrock KB retrieval
- Google Cloud Run deployment
- Azure DevOps platform integration
- Cloudflare Workers, KV, R2, D1 management
- Render service deployment and database queries
- Docker operations and container management

### Databases
- **PostgreSQL** — Schema inspection and query execution
- **BigQuery** — Data warehouse queries with schema inspection
- **ClickHouse** — High-performance analytics queries
- **MotherDuck / DuckDB** — In-process analytical queries
- **Milvus** — Vector database operations
- **Neo4j** — Graph database management
- **Qdrant** — Semantic memory and vector search
- **SingleStore** — Multi-model database interaction
- **Neon** — Serverless Postgres with natural language interface
- **Redis** — Key-value store with natural language interface
- **SQLite** — Embedded database operations

### Development Tools
- **GitHub** — Repository, issues, PRs, workflows
- **GitLab** — Repository management
- **Gitee** — Alternative Git platform
- **CircleCI** — Build failure diagnosis
- **Buildkite** — Pipeline management
- **Semgrep** — Code security scanning
- **SonarQube** — Code quality analysis
- **Playwright / Puppeteer** — Browser automation
- **Browserbase** — Cloud browser automation

### AI & LLM Services
- **Perplexity** — Real-time web research
- **ElevenLabs** — Text-to-speech synthesis
- **Comet Opik** — LLM observability and monitoring
- **Apify** — 3,000+ pre-built cloud tools for data extraction

### Web & Search
- **Exa** — AI-optimized search engine
- **Tavily** — Agent search with content extraction
- **Firecrawl** — Web data extraction and scraping
- **Kagi** — Premium search integration
- **Google News** — News feed access
- **Browser MCP (UI-TARS)** — Lightweight LLM browser control via accessibility tree

### Business & Finance
- **Stripe** — Payment processing and management
- **PayPal** — Payment integration
- **Chargebee** — Subscription management
- **Ramp** — Spend analysis and financial insights
- **Alpha Vantage** — Financial market data
- **Mercado Libre** — Marketplace operations
- **Mercado Pago** — Latin American payment processing

### Communication
- **Slack** — Workspace messaging and data
- **Gmail** — Email management
- **Mailgun** — Transactional email API
- **Twilio** — SMS and voice communications
- **Microsoft Teams** — Team messaging integration
- **iMessage** — Secure iMessage database interface
- **WayStation** — Unified connector for Notion, Slack, Monday, Airtable (under 90 seconds to configure)

### Project Management
- **Notion** — Workspace and database management
- **Jira / Atlassian** — Issue tracking and project management
- **Confluence** — Documentation and wiki
- **Linear** — Modern issue tracking
- **Plane** — Project workflow automation
- **Asana** — Task and project management

### Content & Media
- **Mux** — Video API management
- **Contentful** — Headless CMS operations
- **Spotify** — Music platform integration
- **YouTube** — Video content interaction
- **BlueSky** — Social media platform

### E-commerce
- **Shopify** — Store management and product data

### DevOps & Monitoring
- **Grafana** — Dashboard and alerting
- **Sentry** — Error tracking and performance
- **Datadog** — Infrastructure and APM monitoring
- **PagerDuty** — Incident management
- **Portainer** — Container management with natural language
- **k8s-mcp-server** — Kubernetes CLI (kubectl, helm, istioctl, argocd) in Docker sandbox

### Security
- **GitHub Advanced Security** — Vulnerability detection
- **Cycode** — Code security scanning
- **RAD Security** — Kubernetes security
- **pluggedin-mcp-proxy** — Proxy combining multiple MCP servers with visibility, policy, and discovery

### General Purpose / Utilities
- **context-awesome** — Query 8,500+ curated awesome lists (1M+ items)
- **Supabase** — Backend-as-a-service platform
- **Box** — Enterprise content management
- **Neon** — Serverless database management

---

## Quick Reference Links

| Resource | URL |
|---|---|
| MCP Documentation | https://modelcontextprotocol.io/docs/getting-started/intro |
| Full Docs Index | https://modelcontextprotocol.io/llms.txt |
| Specification (Current) | https://modelcontextprotocol.io/specification/2025-11-25 |
| MCP Registry | https://modelcontextprotocol.io/registry/about |
| Community / Contributing | https://modelcontextprotocol.io/community/contributing |
| Extensions Overview | https://modelcontextprotocol.io/extensions/overview |
| MCP Apps Docs | https://apps.extensions.modelcontextprotocol.io |
| Official Servers (GitHub) | https://github.com/modelcontextprotocol/servers |
| GitHub Organization | https://github.com/modelcontextprotocol |
| Python SDK | https://github.com/modelcontextprotocol/python-sdk |
| TypeScript SDK | https://github.com/modelcontextprotocol/typescript-sdk |
| Awesome MCP Servers (wong2) | https://github.com/wong2/awesome-mcp-servers |
| Awesome MCP Servers (punkpeye) | https://github.com/punkpeye/awesome-mcp-servers |
| MCP Server Directory | https://mcpservers.org |
| Discord Community | https://discord.gg/6CSzBmMkjX |

---

*Document compiled March 2026 from modelcontextprotocol.io (94+ pages) and GitHub ecosystem research.*
