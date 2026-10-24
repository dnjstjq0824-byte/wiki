# Treewiki Design System & Typographic Guidelines

## 1. Design Philosophy
Treewiki is engineered for deep reading, scholarly research, and effortless collaborative curation.
The visual interface emphasizes crisp typography, high informational density with comfortable breathing room, and instantaneous dark/light theme switching.

---

## 2. Color Palette & Semantic Tokens

### 2.1 Brand Tokens
- **Brand Primary (Tree Forest)**: `#0F766E` (Light) / `#14B8A6` (Dark)
- **Brand Accent (Leaf Bright)**: `#10B981` (Light) / `#34D399` (Dark)
- **Knowledge Slate**: `#0B192C` (Deep Slate background for Dark mode)
- **Neutral Light**: `#F8FAFC` (Canvas background for Light mode)

### 2.2 Semantic & Editorial Colors
- **Wiki Hyperlink (Internal)**: `#0284C7` (Light) / `#38BDF8` (Dark)
- **Red Link (Non-existent Page)**: `#DC2626` (Light) / `#F87171` (Dark)
- **External Link**: `#6366F1` (Light) / `#818CF8` (Dark)
- **Diff Addition**: Background `#DCFCE7`, Text `#15803D`, Inline `#86EFAC`
- **Diff Deletion**: Background `#FEE2E2`, Text `#B91C1C`, Inline `#FCA5A5`
- **Talk Consensus Green**: `#16A34A`
- **Talk Closed Gray**: `#64748B`

---

## 3. Typographic Hierarchy (Reading Engine)

- **Primary Reading Font**: `Pretendard`, `-apple-system`, `BlinkMacSystemFont`, `"Segoe UI"`, `Roboto`, sans-serif
- **Code & Syntax Font**: `"JetBrains Mono"`, `"Fira Code"`, monospace
- **Body Text Size**: `16px` (`1rem`)
- **Body Line Height**: `1.75` (optimal reading cadence for encyclopedic long-form articles)
- **Max Content Width**: `78ch` (approximately `760px` to `820px` for optimal eye tracking)

---

## 4. Standard Wiki Components

### 4.1 Infobox (Semantic Entity Summary)
- Floating right container (`width: 320px`, `margin: 0 0 1.5rem 1.5rem`).
- Table layout with 1px border (`border-slate-200 dark:border-slate-700`).
- Header with entity title, primary thumbnail, and key-value metadata rows.

### 4.2 Table of Contents (TOC)
- Collapsible floating sidebar on desktop (width 260px) and inline expandable card on mobile.
- Active scroll-spy highlighting the currently visible section heading.

### 4.3 Footnotes & Reference Accordion
- Hover tooltip showing citation title, author, and external link.
- Clicking reference indicator jumps to bottom bibliography with backlink return indicator (`^`).
