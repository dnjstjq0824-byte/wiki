# Treewiki Editorial Policy & Legal Governance Specification

## 1. Core Editorial Pillars

### 1.1 Neutral Point of View (NPOV)
All articles on Treewiki must be written from a neutral point of view, representing fairly, proportionately, and as far as possible without editorial bias, all significant views that have been published by reliable sources.
- No soapboxing, advocacy, or promotional framing.
- Contested assertions must be attributed to verifiable sources rather than stated as absolute facts.

### 1.2 Verifiability & Reliable Sources (RS)
Readers must be able to verify information against authoritative third-party publications.
- Self-published sources (personal blogs, unverified social media) are prohibited for factual claims.
- Academic peer-reviewed journals, recognized news organizations, books by reputable publishing houses, and official government statistics serve as Tier 1 sources.

### 1.3 Biographies of Living Persons (BLP / 인물 정보 보호)
Articles concerning living individuals require exceptional sensitivity and strict adherence to verifiability.
- Unsourced or poorly sourced contentious material about living persons must be removed immediately without waiting for discussion.
- Adheres to Korean Information and Communications Network Act (정보통신망법 제44조의2) regarding temporary post suspension (임시조치/게시중단) upon formal claim of rights infringement or defamation.

---

## 2. Licensing and Intellectual Property (CC BY-SA 4.0)

1. **Text Contributions**:
   - All text contributed to Treewiki is released under the **Creative Commons Attribution-ShareAlike 4.0 International License (CC BY-SA 4.0)** and GNU Free Documentation License (GFDL) dual-licensing option.
   - Contributors grant Treewiki and downstream reusers perpetual, irrevocable permission to redistribute and adapt content, provided attribution and copyleft reciprocity are preserved.

2. **Media and Assets**:
   - Uploaded images, audio, and video must specify a valid open license (CC0 / Public Domain, CC BY, CC BY-SA) or meet strict Fair Use (공정이용) criteria under Korean Copyright Act (저작권법 제35조의5) and US Copyright Act 17 U.S.C. § 107.

---

## 3. Korean & Global Statutory Compliance

| Regulation | Compliance Mechanism |
|---|---|
| **Personal Information Protection Act (개인정보보호법 PIPA)** | Automated EXIF scrubber on image uploads; IP address truncation/hashing in public logs; strict revision oversight/suppression mechanism for residential identification numbers (RRN), phone numbers, and private addresses. |
| **Information & Communications Network Act (정보통신망법 제44조의2)** | Digital Rights Violation Center (권리침해신고센터) workflow: 30-day temporary blackout flag (`is_temporarily_blinded = true`) upon receipt of formal prima facie defamation complaint. |
| **EU GDPR (Regulation 2016/679)** | Article 17 Right to Erasure ("Right to be Forgotten"): Permanent cryptographic scrubbing of user identifiers and private history via Oversight role. |
| **US COPPA / Children's Online Privacy** | Age gate restriction (under 14 in KR, under 13 in US/EU) preventing unauthorized collection of personal telemetry. |

---

## 4. Dispute Resolution Ladder

```
[Level 1: Article Talk Discussion]
       │
       ▼ (If deadlock exceeds 72h)
[Level 2: Request for Comments (RFC) & Community Polling]
       │
       ▼ (If disruptive edit war occurs)
[Level 3: Page Protection (Autoconfirmed / Sysop Only)]
       │
       ▼ (If bad-faith vandalism or conduct violation)
[Level 4: Sysop Dispute Arbitration & Account Block]
```

