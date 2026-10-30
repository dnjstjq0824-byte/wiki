# Treewiki Privacy & Data Protection Specification

## 1. Anonymous User IP Address Handling & Masking

Anonymous contributors are identified by IP address, which constitutes personal data under PIPA and GDPR.

### 1.1 Public Masking Strategy
- **IPv4**: The final octet is masked in the public UI (`198.51.100.42` -> `198.51.100.xxx`).
- **IPv6**: Only the initial 64-bit routing prefix is shown (`2001:0db8:85a3:0000:0000:8a2e:0370:7334` -> `2001:0db8:85a3::/64`).
- **Full IP Retention**: The unmasked IP address is stored encrypted at rest (AES-256-GCM) in `revisions.ip_address_encrypted` and is accessible only to Sysops and Oversight officers for abuse prevention and block enforcement. Unmasked IP logs expire after 90 days.

---

## 2. Media Asset EXIF Sanitization

When images (JPEG, PNG, WebP, TIFF) are uploaded:
1. The media ingestion pipeline parses all EXIF, IPTC, and XMP metadata blocks.
2. Geolocation tags (`GPSLatitude`, `GPSLongitude`, `GPSAltitude`) are irreversibly stripped.
3. Hardware serial numbers, camera model identifiers, and creator author tags are removed.
4. Only technical rendering attributes (dimensions, color space, orientation rotation) are retained.

---

## 3. Revision Suppression & The "Oversight" Protocol

For sensitive leaks, doxxing, defamation, or regulatory violations:
1. **Soft Deletion (`revisions.is_deleted`)**: Standard Sysops can hide a revision from regular readers. The edit still exists in admin logs.
2. **Hard Oversight (`revisions.is_oversight_suppressed`)**:
   - Authorized Oversight officers suppress the revision text, user identity, and edit summary.
   - Suppressed revisions are excluded from public API responses, database read queries, and search engine indices.
   - Requires audit log recording with ticket reference and dual-authorized verification (Four-Eyes principle for high-risk legal expungement).

---

## 4. User Account Deletion & Right to be Forgotten (GDPR Art. 17)

Wiki text contributions are irrevocably licensed under CC BY-SA 4.0 and cannot be retroactively destroyed (to maintain article authorship chain and integrity).
However, personal account data is scrubbed:
- `users.email` -> NULL
- `users.password_hash` -> NULL
- `users.username` -> Renamed to `FormerUser_<UUID>`
- `users.profile_bio` -> NULL
- All tracking cookies and device session tokens are invalidated.

