# Model Context Protocol — GitHub Organization Deep Dive
> github.com/modelcontextprotocol · Compiled March 2026

---

## Table of Contents

1. [Organization Overview](#1-organization-overview)
2. [Official SDKs](#2-official-sdks)
3. [Core Tools & Infrastructure](#3-core-tools--infrastructure)
4. [Specification & Documentation](#4-specification--documentation)
5. [Extensions & Experimental](#5-extensions--experimental)
6. [Working Groups & Governance](#6-working-groups--governance)
7. [Official Reference Servers](#7-official-reference-servers)
8. [Community MCP Server Ecosystem](#8-community-mcp-server-ecosystem)
9. [MCP Servers by Category](#9-mcp-servers-by-category)

---

## 1. Organization Overview

**GitHub URL:** https://github.com/modelcontextprotocol
**Total Repositories:** 37+
**License:** Apache 2.0 (code) / CC-BY 4.0 (docs)
**Hosted by:** Linux Foundation
**Maintained by:** Anthropic + community contributors

The `modelcontextprotocol` GitHub org is the canonical home for all things MCP — official SDKs in 10 languages, the protocol specification, reference servers, the server registry, tooling, extensions, and working groups. The `servers` repo alone has over 80,000 GitHub stars, making it one of the fastest-growing AI developer repositories in history.

---

## 2. Official SDKs

All SDKs implement the `2025-11-25` specification (current stable). They provide both server-side and client-side libraries for building MCP integrations.

### Primary SDKs

| Repository | Language | Stars | Maintained With | Key Notes |
|---|---|---|---|---|
| [typescript-sdk](https://github.com/modelcontextprotocol/typescript-sdk) | TypeScript | 11,800 | Anthropic | Most widely used; runs on Node.js, Bun, Deno; v2 in progress for Q1 2026 |
| [python-sdk](https://github.com/modelcontextprotocol/python-sdk) | Python | 22,000 | Anthropic | Highest star count among SDKs; includes FastMCP for rapid server building |
| [go-sdk](https://github.com/modelcontextprotocol/go-sdk) | Go | 4,100 | Google | Primary APIs for constructing MCP clients and servers |
| [csharp-sdk](https://github.com/modelcontextprotocol/csharp-sdk) | C# | 4,000 | Microsoft | Full client + server implementations |
| [java-sdk](https://github.com/modelcontextprotocol/java-sdk) | Java | 3,200 | Spring AI | Enterprise Java integration |
| [rust-sdk](https://github.com/modelcontextprotocol/rust-sdk) | Rust | 3,100 | Anthropic | High-performance, low-overhead implementation |
| [kotlin-sdk](https://github.com/modelcontextprotocol/kotlin-sdk) | Kotlin | 1,300 | JetBrains | IntelliJ/JVM ecosystem; umbrella SDK with client-only + server-only options |
| [swift-sdk](https://github.com/modelcontextprotocol/swift-sdk) | Swift | 1,300 | Anthropic | Apple ecosystem; implements full 2025-11-25 spec |
| [php-sdk](https://github.com/modelcontextprotocol/php-sdk) | PHP | 1,400 | PHP Foundation | Laravel/Symfony ecosystem |
| [ruby-sdk](https://github.com/modelcontextprotocol/ruby-sdk) | Ruby | 745 | Anthropic | Rails ecosystem |

### SDK Feature Matrix

All SDKs implement:
- MCP server and client libraries
- stdio transport (local process communication)
- Streamable HTTP transport (remote connections)
- Tool, Resource, and Prompt primitives
- JSON-RPC 2.0 message format

Most SDKs additionally support:
- OAuth 2.0 / Bearer token authorization
- Server-Sent Events (SSE) for streaming
- Capability negotiation
- Progress notifications
- Pagination

---

## 3. Core Tools & Infrastructure

### Inspector
**Repo:** [inspector](https://github.com/modelcontextprotocol/inspector) | 9,000 stars | TypeScript

The official visual testing tool for MCP servers. Provides an interactive UI to:
- Connect to any MCP server (stdio or HTTP)
- Browse available tools, resources, and prompts
- Execute tools and inspect real-time message flows
- Debug JSON-RPC messages in full detail
- Test authentication flows

Run instantly: `npx @modelcontextprotocol/inspector`

### Registry
**Repo:** [registry](https://github.com/modelcontextprotocol/registry) | 6,500 stars | Go

The official centralized metadata repository for publicly accessible MCP servers. Features:
- REST API for server discovery by clients and aggregators
- Namespace management via DNS verification (reverse-DNS format: `io.github.user/server`)
- Standardized `server.json` format
- Authentication via GitHub account or DNS/HTTP challenge
- Backed by Anthropic, GitHub, PulseMCP, and Microsoft
- Currently in preview (launched September 2025)

### mcpb (Desktop Extensions)
**Repo:** [mcpb](https://github.com/modelcontextprotocol/mcpb) | 1,800 stars | TypeScript

One-click MCP server installation for desktop users. Allows non-technical users to install and manage MCP servers without touching config files.

### Conformance Tests
**Repo:** [conformance](https://github.com/modelcontextprotocol/conformance) | 43 stars | TypeScript

Official conformance test suite for verifying that MCP server and client implementations correctly implement the protocol specification.

### Access (IaC)
**Repo:** [access](https://github.com/modelcontextprotocol/access) | 32 stars | TypeScript

Infrastructure as Code for MCP access management — defines who can access what across the MCP ecosystem.

### DNS (IaC)
**Repo:** [dns](https://github.com/modelcontextprotocol/dns) | 9 stars | TypeScript

Infrastructure as Code for MCP domain and DNS management — manages the `modelcontextprotocol.io` domain infrastructure.

---

## 4. Specification & Documentation

### modelcontextprotocol (Spec Repo)
**Repo:** [modelcontextprotocol](https://github.com/modelcontextprotocol/modelcontextprotocol) | 7,400 stars | TypeScript/MDX

The source of truth for the MCP specification and documentation. Built with Mintlify and deployed at `modelcontextprotocol.io`. Contains:
- Full protocol specification (current: `2025-11-25`)
- Architecture documentation
- Developer guides (build server, build client, connect local/remote)
- Security and authorization documentation
- SDK reference
- Community governance documents
- Extension specifications (MCP Apps, Auth extensions)

**Specification Releases:**
- `2025-06-18` — Initial stable release
- `2025-11-25` — Current stable release (major additions: Tasks, Elicitation, enhanced auth)
- Draft proposals via SEP process (25+ active SEPs)

### Quickstart Resources
**Repo:** [quickstart-resources](https://github.com/modelcontextprotocol/quickstart-resources) | 1,000 stars | Go

Tutorial companion repo with complete working examples of servers and clients from the official MCP getting-started guides.

### Example Remote Server
**Repo:** [example-remote-server](https://github.com/modelcontextprotocol/example-remote-server) | 70 stars | TypeScript

A hosted demonstration server showing how to build and deploy a production-ready remote MCP server with Streamable HTTP transport.

---

## 5. Extensions & Experimental

### ext-apps (MCP Apps)
**Repo:** [ext-apps](https://github.com/modelcontextprotocol/ext-apps) | 1,800 stars | TypeScript

Official spec and SDK for the MCP Apps protocol extension — enables MCP servers to return interactive HTML/JS/CSS UI applications that render inside host clients (Claude Desktop, VS Code, etc.) as sandboxed iframes. Supports React, Vue, Svelte, Vanilla JS starters.

### ext-auth (Authorization Extensions)
**Repo:** [ext-auth](https://github.com/modelcontextprotocol/ext-auth) | 62 stars | MDX

Specification for authorization extensions beyond the core protocol:
- OAuth Client Credentials flow
- Enterprise-Managed Authorization
- Custom auth schemes

### experimental-ext-skills
**Repo:** [experimental-ext-skills](https://github.com/modelcontextprotocol/experimental-ext-skills) | 46 stars

Exploration of skills discovery and distribution through MCP primitives. Maintained by the Skills Interest Group. Not yet production-ready.

### experimental-ext-grouping
**Repo:** [experimental-ext-grouping](https://github.com/modelcontextprotocol/experimental-ext-grouping) | 5 stars | JavaScript

Experimental exploration of primitive organization/grouping within MCP.

### experimental-ext-variants
**Repo:** [experimental-ext-variants](https://github.com/modelcontextprotocol/experimental-ext-variants) | 5 stars | Go

Multi-language variants reference implementation.

### experimental-ext-interceptors
**Repo:** [experimental-ext-interceptors](https://github.com/modelcontextprotocol/experimental-ext-interceptors) | 8 stars | C#

Interceptor extension reference implementation — allows middleware-style processing of MCP messages.

---

## 6. Working Groups & Governance

### financial-services-interest-group
**Repo:** [financial-services-interest-group](https://github.com/modelcontextprotocol/financial-services-interest-group) | 47 stars

Industry-specific interest group focused on MCP use cases in financial services — data standards, compliance, audit requirements, and integration patterns for trading, banking, and wealth management.

### agents-wg (Agents Working Group)
**Repo:** [agents-wg](https://github.com/modelcontextprotocol/agents-wg) | 1 star

Staging grounds for the Agents Working Group — focused on multi-agent coordination patterns, agent-to-agent communication via MCP, and standardizing how AI agents discover and delegate to each other.

### transports-wg (Transports Working Group)
**Repo:** [transports-wg](https://github.com/modelcontextprotocol/transports-wg) | 12 stars

Working group focused on transport layer evolution — evaluating new transport mechanisms beyond stdio and Streamable HTTP.

### .github (Org Discussions)
**Repo:** [.github](https://github.com/modelcontextprotocol/.github) | 69 stars

Organization-level README and discussions forum for cross-repo conversations, RFCs, and governance.

---

## 7. Official Reference Servers

**Repo:** [servers](https://github.com/modelcontextprotocol/servers) | **80,500 stars** | TypeScript

The most-starred repo in the org. Contains official reference server implementations — educational demonstrations of MCP features, NOT production-ready deployments.

### Active Reference Servers

| Server | Purpose | Key Tools |
|---|---|---|
| **everything** | Comprehensive reference with all primitive types | All tool/resource/prompt examples |
| **fetch** | Web content retrieval optimized for LLM use | `fetch` — get URL contents as markdown |
| **filesystem** | Secure file operations with configurable access | `read_file`, `write_file`, `list_directory`, `search_files`, `get_file_info` |
| **git** | Git repository operations | `git_log`, `git_diff`, `git_status`, `git_commit`, `git_branch` |
| **memory** | Knowledge graph persistent memory | `create_entities`, `create_relations`, `search_nodes`, `open_nodes` |
| **sequential-thinking** | Dynamic problem-solving through thought sequences | `sequentialthinking` — structured reasoning chains |
| **time** | Time and timezone operations | `get_current_time`, `convert_time` |

### Archived Reference Servers
(Moved to `servers-archived` repo — still usable but no longer actively maintained)

AWS KB Retrieval, Brave Search, EverArt, GitHub, GitLab, Google Drive, Google Maps, PostgreSQL, Puppeteer, Redis, Sentry, Slack, SQLite

---

## 8. Community MCP Server Ecosystem

As of early 2026, the community ecosystem has exploded to **7,260+ documented MCP servers** tracked across multiple curated lists.

### Top Curated Lists

| Repository | Focus | URL |
|---|---|---|
| punkpeye/awesome-mcp-servers | Large general collection | https://github.com/punkpeye/awesome-mcp-servers |
| wong2/awesome-mcp-servers | Curated with quality filtering | https://github.com/wong2/awesome-mcp-servers |
| appcypher/awesome-mcp-servers | Production & experimental focus | https://github.com/appcypher/awesome-mcp-servers |
| rohitg00/awesome-devops-mcp-servers | DevOps-specific | https://github.com/rohitg00/awesome-devops-mcp-servers |
| ever-works/awesome-mcp-servers | Broad production-ready | https://github.com/ever-works/awesome-mcp-servers |
| patriksimek/awesome-mcp-servers-2 | Search, social, browser automation | https://github.com/patriksimek/awesome-mcp-servers-2 |
| PipedreamHQ/awesome-mcp-servers | Pipedream integrations | https://github.com/PipedreamHQ/awesome-mcp-servers |
| TensorBlock/awesome-mcp-servers | Comprehensive collection | https://github.com/TensorBlock/awesome-mcp-servers |
| tolkonepiu/best-of-mcp-servers | Weekly ranked list | https://github.com/tolkonepiu/best-of-mcp-servers |
| esc5221/awesome-awesome-mcp-servers | Meta-index of all lists | https://github.com/esc5221/awesome-awesome-mcp-servers |
| habitoai/awesome-mcp-servers | Tools & discovery focus | https://github.com/habitoai/awesome-mcp-servers |
| MobinX/awesome-mcp-list | Concise curated list | https://github.com/MobinX/awesome-mcp-list |

**Web Directory:** [mcpservers.org](https://mcpservers.org) — Search and browse the full ecosystem

---

## 9. MCP Servers by Category

### Cloud Infrastructure & DevOps

| Server | Description | GitHub |
|---|---|---|
| aws-kb-retrieval | AWS Bedrock Knowledge Base retrieval | archived in modelcontextprotocol/servers |
| aws-cdk | Infrastructure as Code with AWS CDK | community |
| google-cloud-run | Deploy and manage Cloud Run services | community |
| azure-devops | Azure DevOps platform integration | community |
| cloudflare | Workers, KV, R2, D1, Queues management | community |
| render | Service deployment and database queries | community |
| docker | Container and Compose stack management | community |
| k8s-mcp-server | Kubernetes CLI (kubectl, helm, istioctl, argocd) in Docker sandbox | rohitg00 |
| portainer-mcp | Portainer container management via natural language | community |
| grafana | Dashboards and alerting | community |
| datadog | Infrastructure and APM monitoring | community |
| sentry | Error tracking and performance monitoring | archived |
| pagerduty | Incident management | community |

### Databases

| Server | Description | Auth |
|---|---|---|
| postgres | Schema inspection + query execution | API key |
| bigquery | Google BigQuery data warehouse queries | API key |
| clickhouse | High-performance analytics queries | API key |
| motherduck / duckdb | In-process analytical queries | API key |
| milvus | Vector database operations | API key |
| neo4j | Graph database management | API key |
| qdrant | Semantic memory and vector search | API key |
| singlestore | Multi-model database interaction | API key |
| neon | Serverless Postgres with NL interface | API key |
| redis | Key-value store with NL interface | API key |
| sqlite | Embedded database operations | local |
| supabase | Backend-as-a-service platform | API key |
| mysql | Multi-database MySQL management | API key |

### Development Tools

| Server | Description |
|---|---|
| github | Repos, issues, PRs, workflows via GitHub API |
| gitlab | GitLab repository management |
| gitee | Alternative Git platform |
| circleci | Build failure diagnosis and pipeline management |
| buildkite | CI/CD pipeline management |
| semgrep | Code security scanning |
| sonarqube | Code quality analysis |
| playwright | Browser automation for testing and scraping |
| puppeteer | Headless browser control |
| browserbase | Cloud browser automation |
| browser-mcp (UI-TARS) | Lightweight LLM browser control via accessibility tree |

### AI & LLM Services

| Server | Description |
|---|---|
| perplexity | Real-time web research |
| elevenlabs | Text-to-speech synthesis |
| comet-opik | LLM observability and monitoring |
| apify | 3,000+ pre-built cloud tools for web data extraction |
| context-awesome | Query 8,500+ curated awesome lists (1M+ items) |

### Web & Search

| Server | Description |
|---|---|
| exa | AI-optimized search engine |
| tavily | Agent search with content extraction |
| firecrawl | Web data extraction and scraping |
| kagi | Premium search integration |
| google-news | News feed access |
| brave-search | Privacy-preserving web search |

### Business & Finance

| Server | Description |
|---|---|
| stripe | Payment processing and management |
| paypal | Payment integration |
| chargebee | Subscription management |
| ramp | Spend analysis and financial insights |
| alpha-vantage | Financial market data |
| mercado-libre | Marketplace operations |
| mercado-pago | Latin American payment processing |
| quickbooks | Accounting and invoicing |

### Communication & Messaging

| Server | Description |
|---|---|
| slack | Workspace messaging, search, file access |
| gmail | Email management and search |
| mailgun | Transactional email API |
| twilio | SMS and voice communications |
| microsoft-teams | Team messaging integration |
| imessage | Secure iMessage database interface |
| waystation | Unified connector: Notion, Slack, Monday, Airtable |
| discord | Discord bot and server management |

### Project Management

| Server | Description |
|---|---|
| notion | Workspace, databases, pages |
| jira | Issue tracking and project management |
| confluence | Documentation and wiki |
| linear | Modern issue tracking |
| plane | Project workflow automation |
| asana | Task and project management |
| trello | Kanban board management |
| monday | Work OS integrations |

### Content & Media

| Server | Description |
|---|---|
| mux | Video API management |
| contentful | Headless CMS operations |
| spotify | Music platform integration |
| youtube | Video content interaction and search |
| bluesky | Social media platform |
| twitter-x | X/Twitter integration |

### E-Commerce

| Server | Description |
|---|---|
| shopify | Store management, product data, orders |
| woocommerce | WordPress e-commerce |
| amazon-sp | Amazon Selling Partner API |

### Security

| Server | Description |
|---|---|
| github-advanced-security | Vulnerability detection and code scanning |
| cycode | Application security scanning |
| rad-security | Kubernetes security |
| pluggedin-mcp-proxy | Multi-server proxy with visibility, policy, and discovery |

### Data Science & Analytics

| Server | Description |
|---|---|
| jupyter | Jupyter notebook execution and management |
| databricks | Spark clusters and ML pipelines |
| snowflake | Cloud data warehouse queries |
| tableau | BI dashboard data access |
| looker | Looker analytics integration |

### Location & Maps

| Server | Description |
|---|---|
| google-maps | Places, directions, geocoding |
| mapbox | Custom maps and geospatial data |
| goplaces | Google Places search |

### Novel Architectures

| Server | Description |
|---|---|
| MicroMCP | Single-purpose server composition behind a gateway with security isolation, unified discovery, policy enforcement, and audit logging |
| Mobile MCP | Platform-agnostic mobile automation — iOS + Android, emulators and real devices |
| pluggedin-mcp-proxy | Combines multiple MCP servers into one interface with extensive visibility across all connected servers |
| mcp-remote | Bridge: exposes stdio MCP servers as remote HTTP endpoints, and vice versa |

---

## Quick Reference

| Resource | URL |
|---|---|
| GitHub Organization | https://github.com/modelcontextprotocol |
| All Repositories | https://github.com/orgs/modelcontextprotocol/repositories |
| Servers Repo (80K stars) | https://github.com/modelcontextprotocol/servers |
| Python SDK (22K stars) | https://github.com/modelcontextprotocol/python-sdk |
| TypeScript SDK (11.8K) | https://github.com/modelcontextprotocol/typescript-sdk |
| Inspector Tool | https://github.com/modelcontextprotocol/inspector |
| Registry | https://github.com/modelcontextprotocol/registry |
| Specification Repo | https://github.com/modelcontextprotocol/modelcontextprotocol |
| MCP Apps Extension | https://github.com/modelcontextprotocol/ext-apps |
| Awesome Servers (punkpeye) | https://github.com/punkpeye/awesome-mcp-servers |
| Awesome Servers (wong2) | https://github.com/wong2/awesome-mcp-servers |
| MCP Server Directory | https://mcpservers.org |

---

*Compiled March 2026 from github.com/modelcontextprotocol (37 repos) and community ecosystem research. 7,260+ total community MCP servers documented.*
