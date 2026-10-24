# Treewiki (🌲) — Global Collaborative Knowledge Platform

> **A structured collaborative knowledge and reference platform designed to organize human knowledge into clear, interconnected topic trees.**  
> Developed for the Booking Network portfolio. Released under **CC BY-SA 4.0**.

---

## 1. System Overview

Treewiki combines the best principles of traditional wiki encyclopedias (MediaWiki/나무위키) with modern structured taxonomy (Directed Acyclic Graph topic trees) and real-time conflict-free collaboration.

* **Core Backend**: Go 1.22+ Modular Monolith (PostgreSQL 16, Redis 7.2, MinIO S3, Meilisearch)
* **Frontend Web Client**: Vue 3 + Nuxt 3 (SSR/SEO) + Tailwind CSS + Pinia
* **Diff Engine**: Tokenized Myers Diff algorithm with word/line granularity
* **Wikitext & Markdown Engine**: Hybrid parser with automated Table of Contents (TOC) and footnote citation generation
* **Governance & Security**: 9-tier RBAC, IP anonymization (PIPA/GDPR), and real-time abuse filtering

---

## 2. Repository Layout

```
.
├── apps/
│   └── web/                   # Nuxt 3 SSR Web Client (Vue 3, Tailwind CSS)
├── database/
│   └── migrations/
│       └── 001_initial_schema.sql  # 25+ Tables PostgreSQL 16 Master DDL
├── docs/
│   ├── treewiki-feature-inventory.csv   # 50+ item feature breakdown
│   ├── feature-parity-matrix.csv        # Parity vs Wikipedia & Namuwiki
│   ├── rbac-matrix.yaml                 # 9 user roles & permission matrix
│   ├── screen-registry.yaml             # Complete UI route & screen registry
│   ├── event-catalog.yaml               # Event-driven architecture audit catalog
│   ├── policy-spec.md                   # NPOV, BLP & CC BY-SA 4.0 governance
│   ├── privacy-data-protection.md       # IP masking & Oversight suppression
│   ├── threat-model.md                  # STRIDE threat model & CSP v3
│   ├── design-system.md                 # Typographic scale & color tokens
│   ├── architecture-overview.md         # C4 system diagrams & parser design
│   └── roadmap-wbs.md                   # 100-step master development roadmap
├── infrastructure/
│   └── docker-compose.yml     # PostgreSQL 16, Redis 7.2, MinIO, Meilisearch
├── pkg/
│   ├── diff/                  # Myers Diff algorithm (tested)
│   └── parser/                # Hybrid Wikitext/Markdown parser (tested)
├── proto/
│   └── treewiki/v1/wiki.proto # gRPC & Protobuf contract definitions
└── services/
    └── gateway/               # Go API Gateway & Health/Preview router
```

---

## 3. Quickstart

### 3.1 Start Backing Infrastructure
```bash
cd infrastructure
docker compose up -d
```

### 3.2 Run Go Backend Unit Tests
```bash
go test -v ./...
```

### 3.3 Run Go API Gateway
```bash
go run services/gateway/cmd/main.go
# Health check: http://localhost:8080/healthz
```

### 3.4 Run Web Client
```bash
pnpm install
pnpm dev
# Web Portal: http://localhost:3000
```
