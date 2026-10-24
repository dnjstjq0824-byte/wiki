# Treewiki Threat Model & Security Architecture (STRIDE)

## 1. Threat Classification & Mitigation Matrix

| Category | Specific Threat Vector | Impact | Mitigation Mechanism |
|---|---|---|---|
| **Spoofing** | IP address spoofing via manipulated `X-Forwarded-For` HTTP headers | High | Strip untrusted proxy headers at Cloudflare/Traefik reverse proxy; only trust configured upstream edge CIDRs. |
| **Spoofing** | Impersonating bot accounts or sysop usernames | High | Strict username regex forbidding unicode homoglyph attacks; automated bot flag requirement (`is_bot = true`). |
| **Tampering** | Stored XSS via malicious Wikitext / Markdown AST parsing | Critical | Two-stage HTML sanitization: Server-side DOMPurify/Bluemonday AST filter; strict Content Security Policy (`script-src 'self'`). |
| **Tampering** | Race condition / Lost updates during concurrent page editing | High | Optimistic concurrency control via revision hash/ETag checking; 3-way text merge with visual conflict resolution. |
| **Tampering** | Malicious SVG uploads containing embedded JavaScript | Critical | Strict SVG XML parser rejecting `<script>`, `onload`, `xlink:href` to external resources, and XML entity expansion (XXE). |
| **Repudiation** | Vandal denying defamatory edits or edit history manipulation | Medium | Cryptographic SHA-256 chaining of all revision records; tamper-evident append-only PostgreSQL ledger with immutable timestamps. |
| **Info Disclosure** | Exposing PII, passwords, or deleted revisions through search API | High | Database Row-Level Security (RLS) and query filters strictly excluding `is_deleted` and `is_oversight_suppressed` records from general search indices. |
| **Denial of Service** | Catastrophic ReDoS in Wikitext regex parser | High | Non-backtracking linear time parser engine; strict recursion depth limits (max 40 nested templates); 500ms parser execution timeout. |
| **Denial of Service** | Recursive template expansion bomb (`{{TemplateA}} -> {{TemplateB}} -> {{TemplateA}}`) | High | Template call stack depth tracker (max 10 levels) and loop detection preventing circular template expansion. |
| **Elevation of Privilege**| Unauthorized editing of semi-protected pages or administrative namespaces | High | Server-side RBAC middleware validating page protection level and namespace ACL against user claims before accepting revision payload. |

---

## 2. Content Security Policy (CSP v3)

```http
Content-Security-Policy: default-src 'self';
  script-src 'self' 'nonce-{RANDOM}' https://cdn.jsdelivr.net;
  style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
  font-src 'self' https://fonts.gstatic.com;
  img-src 'self' data: https://storage.treewiki.org;
  connect-src 'self' wss://treewiki.org;
  frame-ancestors 'none';
  base-uri 'self';
  form-action 'self';
```
