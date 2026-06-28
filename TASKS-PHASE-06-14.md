# PHPVM — Task Specification: PHASE-06 to PHASE-14

> ادامه مستقیم از TASKS.md
> فرمت: Execution Task List for Kiro

---

## Quick Reference

| Phase | Release | عنوان | Tasks |
|-------|---------|-------|-------|
| PHASE-06 | v0.7.0 | Download Manager | 0106–0125 |
| PHASE-07 | v0.8.0 | PATH Manager | 0126–0145 |
| PHASE-08 | v0.9.0 | Composer Integration | 0146–0165 |
| PHASE-09 | v0.10.0 | Doctor & Environment | 0166–0185 |
| PHASE-10 | v0.11.0 | Project Integration | 0186–0205 |
| PHASE-11 | v0.12.0 | Plugin System | 0206–0225 |
| PHASE-12 | v0.13.0 | Testing & Quality | 0226–0250 |
| PHASE-13 | v0.14.0 | CI/CD | 0251–0275 |
| PHASE-14 | v1.0.0 | Release & Distribution | 0276–0320 |

---

---

# PHASE-06 — Download Manager

**Release:** `v0.7.0`
**هدف:** دانلود ایمن، قابل اطمینان، و cross-platform فایل‌های PHP از منابع رسمی.

```
TASK-EX-0106  HTTP Client Factory
TASK-EX-0107  Download Session Model
TASK-EX-0108  Download Queue
TASK-EX-0109  PHP Windows Source Driver (windows.php.net)
TASK-EX-0110  PHP Unix Source Driver (php.net/releases)
TASK-EX-0111  Mirror Fallback Strategy
TASK-EX-0112  Progress Tracker
TASK-EX-0113  Progress Renderer (Terminal)
TASK-EX-0114  Resume Support (Range Request)
TASK-EX-0115  Checksum Verifier (SHA256)
TASK-EX-0116  Signature Verifier (GPG — optional)
TASK-EX-0117  Download Cache Manager
TASK-EX-0118  Retry Strategy (Exponential Backoff)
TASK-EX-0119  Timeout Handler
TASK-EX-0120  ZIP Extractor
TASK-EX-0121  TAR.GZ Extractor
TASK-EX-0122  Atomic Extraction (temp → rename)
TASK-EX-0123  Downloader Unit Tests
TASK-EX-0124  Extractor Unit Tests
TASK-EX-0125  Build Verification & Git Tag v0.7.0
```

**Execution Order:** 0106 → 0108 → 0109 → 0110 → 0111 → 0112 → 0113 → 0114 → 0115 → 0116 → 0117 → 0118 → 0119 → 0107 → 0120 → 0121 → 0122 → 0123 → 0124 → 0125

**Files Produced:**
```
internal/downloader/
├── client.go          (TASK-EX-0106)
├── session.go         (TASK-EX-0107)
├── queue.go           (TASK-EX-0108)
├── sources/
│   ├── windows.go     (TASK-EX-0109)
│   └── unix.go        (TASK-EX-0110)
├── mirror.go          (TASK-EX-0111)
├── progress.go        (TASK-EX-0112, 0113)
├── resume.go          (TASK-EX-0114)
├── verify.go          (TASK-EX-0115, 0116)
├── cache.go           (TASK-EX-0117)
├── retry.go           (TASK-EX-0118)
├── timeout.go         (TASK-EX-0119)
└── downloader_test.go (TASK-EX-0123)
internal/archive/
├── zip.go             (TASK-EX-0120)
├── targz.go           (TASK-EX-0121)
├── atomic.go          (TASK-EX-0122)
└── archive_test.go    (TASK-EX-0124)
```

**Definition of Done:**
- [ ] دانلود با progress bar کار می‌کند
- [ ] Resume بعد از قطع شبکه کار می‌کند
- [ ] SHA256 verify می‌شود و در صورت خطا فایل حذف می‌شود
- [ ] Retry تا ۳ بار با exponential backoff
- [ ] ZIP و TAR.GZ روی هر سه پلتفرم استخراج می‌شوند
- [ ] `go test ./internal/downloader/... ./internal/archive/...` پاس می‌شود

---

---

# PHASE-07 — PATH Manager

**Release:** `v0.8.0`
**هدف:** مدیریت ایمن PATH روی هر سه پلتفرم بدون خراب کردن تنظیمات موجود.

```
TASK-EX-0126  PATH Entry Model
TASK-EX-0127  PATH Reader (Cross-platform)
TASK-EX-0128  PATH Writer (Cross-platform)
TASK-EX-0129  PATH Deduplicator
TASK-EX-0130  Windows Registry Reader (HKCU\Environment)
TASK-EX-0131  Windows Registry Writer (HKCU\Environment)
TASK-EX-0132  Windows System PATH Driver (HKLM — elevated)
TASK-EX-0133  Unix Shell Profile Detector (.bashrc / .zshrc / .profile)
TASK-EX-0134  Unix Shell Profile Writer
TASK-EX-0135  Unix Shell Profile Backup
TASK-EX-0136  macOS Shell Profile Driver (zsh default)
TASK-EX-0137  PHPVM Marker Block (begin/end comment block)
TASK-EX-0138  PATH Add Operation
TASK-EX-0139  PATH Remove Operation
TASK-EX-0140  PATH Verify Operation
TASK-EX-0141  Shell Restart Notification
TASK-EX-0142  PATH Rollback on Failure
TASK-EX-0143  PATH Unit Tests (Windows)
TASK-EX-0144  PATH Unit Tests (Unix)
TASK-EX-0145  Build Verification & Git Tag v0.8.0
```

**Execution Order:** 0126 → 0127 → 0128 → 0129 → 0130 → 0131 → 0132 → 0133 → 0134 → 0135 → 0136 → 0137 → 0138 → 0139 → 0140 → 0141 → 0142 → 0143 → 0144 → 0145

**Files Produced:**
```
internal/path/
├── model.go           (TASK-EX-0126)
├── reader.go          (TASK-EX-0127)
├── writer.go          (TASK-EX-0128)
├── dedup.go           (TASK-EX-0129)
├── windows_registry.go (TASK-EX-0130, 0131, 0132)
├── unix_profile.go    (TASK-EX-0133, 0134, 0135)
├── macos_profile.go   (TASK-EX-0136)
├── marker.go          (TASK-EX-0137)
├── operations.go      (TASK-EX-0138, 0139, 0140)
├── notify.go          (TASK-EX-0141)
├── rollback.go        (TASK-EX-0142)
├── path_windows_test.go (TASK-EX-0143)
└── path_unix_test.go  (TASK-EX-0144)
```

**Definition of Done:**
- [ ] PATH موجود بدون تغییر باقی می‌ماند
- [ ] Entry های تکراری حذف می‌شوند
- [ ] PHPVM block با begin/end marker مشخص است و idempotent است
- [ ] Rollback در صورت خطا کار می‌کند
- [ ] روی هر سه پلتفرم تست شده است
- [ ] `go test ./internal/path/...` پاس می‌شود

---

---

# PHASE-08 — Composer Integration

**Release:** `v0.9.0`
**هدف:** مدیریت کامل Composer به صورت per-version و global.

```
TASK-EX-0146  Composer Version Model
TASK-EX-0147  Composer Binary Locator
TASK-EX-0148  Composer Global Detector
TASK-EX-0149  Composer Per-Version Detector
TASK-EX-0150  Composer Installer Script Downloader (getcomposer.org)
TASK-EX-0151  Composer Signature Verifier
TASK-EX-0152  Composer Install (Global)
TASK-EX-0153  Composer Install (Per-Version)
TASK-EX-0154  Composer Version Checker
TASK-EX-0155  Composer Update Logic
TASK-EX-0156  Composer Rollback on Failed Update
TASK-EX-0157  Composer PATH Registration
TASK-EX-0158  Required Extension Checker (json/phar/openssl/mbstring)
TASK-EX-0159  php.ini Compatibility Check (allow_url_fopen)
TASK-EX-0160  Composer Diagnose Report Builder
TASK-EX-0161  Composer Command (install / update / diagnose)
TASK-EX-0162  Composer Unit Tests
TASK-EX-0163  Composer Integration Tests
TASK-EX-0164  Composer Documentation
TASK-EX-0165  Build Verification & Git Tag v0.9.0
```

**Execution Order:** 0146 → 0147 → 0148 → 0149 → 0150 → 0151 → 0152 → 0153 → 0154 → 0155 → 0156 → 0157 → 0158 → 0159 → 0160 → 0161 → 0162 → 0163 → 0164 → 0165

**Files Produced:**
```
internal/composer/
├── model.go           (TASK-EX-0146)
├── locator.go         (TASK-EX-0147, 0148, 0149)
├── installer.go       (TASK-EX-0150, 0151, 0152, 0153)
├── updater.go         (TASK-EX-0154, 0155, 0156)
├── path.go            (TASK-EX-0157)
├── checks.go          (TASK-EX-0158, 0159)
├── diagnose.go        (TASK-EX-0160)
├── composer_test.go   (TASK-EX-0162)
└── integration_test.go (TASK-EX-0163)
cmd/
└── composer.go        (TASK-EX-0161)
```

**Definition of Done:**
- [ ] `phpvm composer install` Composer را دانلود، verify، و نصب می‌کند
- [ ] `phpvm composer update` به آخرین نسخه stable به‌روز می‌کند
- [ ] `phpvm composer diagnose` گزارش کامل با ✓/✗ نمایش می‌دهد
- [ ] Signature verification قبل از اجرای installer
- [ ] `go test ./internal/composer/...` پاس می‌شود

---

---

# PHASE-09 — Doctor & Environment

**Release:** `v0.10.0`
**هدف:** مدیریت php.ini، Extensions، Xdebug، و Doctor جامع محیط.

```
TASK-EX-0166  php.ini Locator (per-version + system)
TASK-EX-0167  php.ini Parser
TASK-EX-0168  php.ini Writer
TASK-EX-0169  php.ini Backup & Restore
TASK-EX-0170  php.ini Diff Viewer
TASK-EX-0171  Extension Model
TASK-EX-0172  Extension List Scanner (loaded + available)
TASK-EX-0173  Extension Enable/Disable (comment toggle in php.ini)
TASK-EX-0174  Extension Downloader (PECL / official builds)
TASK-EX-0175  Extension Version Compatibility Checker
TASK-EX-0176  Xdebug Version Resolver
TASK-EX-0177  Xdebug Installer
TASK-EX-0178  Xdebug Mode Switcher (develop / coverage / profile / off)
TASK-EX-0179  Xdebug php.ini Block Writer
TASK-EX-0180  Doctor Check: PHP Binary
TASK-EX-0181  Doctor Check: PATH & Registry
TASK-EX-0182  Doctor Check: php.ini
TASK-EX-0183  Doctor Check: Extensions
TASK-EX-0184  Doctor Check: Composer
TASK-EX-0185  Doctor Check: Permissions
TASK-EX-0186  Doctor Check: Xdebug
TASK-EX-0187  Doctor Auto-Fix Engine
TASK-EX-0188  Doctor Report Renderer (colored, JSON)
TASK-EX-0189  env / ext / xdebug / doctor Commands
TASK-EX-0190  Environment Unit Tests
TASK-EX-0191  Build Verification & Git Tag v0.10.0
```

**Execution Order:** 0166 → 0167 → 0168 → 0169 → 0170 → 0171 → 0172 → 0173 → 0174 → 0175 → 0176 → 0177 → 0178 → 0179 → 0180 → 0181 → 0182 → 0183 → 0184 → 0185 → 0186 → 0187 → 0188 → 0189 → 0190 → 0191

**Files Produced:**
```
internal/ini/
├── locator.go         (TASK-EX-0166)
├── parser.go          (TASK-EX-0167)
├── writer.go          (TASK-EX-0168)
├── backup.go          (TASK-EX-0169)
├── diff.go            (TASK-EX-0170)
└── ini_test.go
internal/extension/
├── model.go           (TASK-EX-0171)
├── scanner.go         (TASK-EX-0172)
├── toggle.go          (TASK-EX-0173)
├── downloader.go      (TASK-EX-0174)
├── compat.go          (TASK-EX-0175)
├── xdebug.go          (TASK-EX-0176, 0177, 0178, 0179)
└── extension_test.go
internal/php/
└── doctor.go          (TASK-EX-0180–0188)
cmd/
├── env.go             (TASK-EX-0189)
├── ext.go             (TASK-EX-0189)
├── xdebug.go          (TASK-EX-0189)
└── doctor.go          (TASK-EX-0189)
```

**Definition of Done:**
- [ ] `phpvm env set memory_limit 512M` php.ini را به‌روز می‌کند
- [ ] `phpvm ext enable redis` extension را فعال می‌کند
- [ ] `phpvm xdebug switch coverage` mode را تغییر می‌دهد
- [ ] `phpvm doctor` گزارش کامل با ✓/✗/⚠ نمایش می‌دهد
- [ ] `phpvm doctor --fix` مشکلات ساده را خودکار رفع می‌کند
- [ ] `phpvm doctor --json` خروجی JSON می‌دهد
- [ ] `go test ./internal/ini/... ./internal/extension/...` پاس می‌شود

---

---

# PHASE-10 — Project Integration

**Release:** `v0.11.0`
**هدف:** تجربه developer-first با `.phpvmrc`، shell hooks، و framework detection.

```
TASK-EX-0192  .phpvmrc File Model
TASK-EX-0193  .phpvmrc Parser (version + options)
TASK-EX-0194  .phpvmrc Writer
TASK-EX-0195  .phpvmrc Search (current dir → root, like git)
TASK-EX-0196  phpvm local Command
TASK-EX-0197  phpvm auto Command
TASK-EX-0198  Auto-Switch on Directory Change (Hook Output)
TASK-EX-0199  Framework Detector: Laravel
TASK-EX-0200  Framework Detector: Symfony
TASK-EX-0201  Framework Detector: WordPress
TASK-EX-0202  Framework Detector: Magento
TASK-EX-0203  Framework Detector: CodeIgniter
TASK-EX-0204  Framework Detector: Yii
TASK-EX-0205  Recommended PHP Version per Framework
TASK-EX-0206  phpvm detect Command
TASK-EX-0207  Shell Hook: Bash/Zsh (cd override)
TASK-EX-0208  Shell Hook: PowerShell (prompt function)
TASK-EX-0209  Shell Hook: Fish (chpwd event)
TASK-EX-0210  phpvm shell Command (install / uninstall / status)
TASK-EX-0211  Project Integration Unit Tests
TASK-EX-0212  Build Verification & Git Tag v0.11.0
```

**Execution Order:** 0192 → 0193 → 0194 → 0195 → 0196 → 0197 → 0198 → 0199 → 0200 → 0201 → 0202 → 0203 → 0204 → 0205 → 0206 → 0207 → 0208 → 0209 → 0210 → 0211 → 0212

**Files Produced:**
```
internal/php/
├── local.go           (TASK-EX-0192, 0193, 0194, 0195)
├── auto.go            (TASK-EX-0197, 0198)
└── detector.go        (TASK-EX-0199–0205)
internal/cli/
├── shell_bash.go      (TASK-EX-0207)
├── shell_pwsh.go      (TASK-EX-0208)
└── shell_fish.go      (TASK-EX-0209)
cmd/
├── local.go           (TASK-EX-0196)
├── auto.go            (TASK-EX-0197)
├── detect.go          (TASK-EX-0206)
└── shell.go           (TASK-EX-0210)
scripts/
├── phpvm.sh           (TASK-EX-0207)
├── phpvm.ps1          (TASK-EX-0208)
└── phpvm.fish         (TASK-EX-0209)
```

**Definition of Done:**
- [ ] `phpvm local 8.4` فایل `.phpvmrc` می‌سازد
- [ ] `phpvm auto` نسخه را بر اساس `.phpvmrc` سوئیچ می‌کند
- [ ] `phpvm detect` در پروژه Laravel فریم‌ورک را تشخیص می‌دهد
- [ ] `phpvm shell install` hook را به shell config اضافه می‌کند
- [ ] ورود به پوشه با `.phpvmrc` نسخه را خودکار سوئیچ می‌کند
- [ ] `go test ./internal/php/...` پاس می‌شود

---

---

# PHASE-11 — Plugin System

**Release:** `v0.12.0`
**هدف:** زیرساخت plugin قابل توسعه برای community و enterprise.

```
TASK-EX-0213  Plugin Interface Definition
TASK-EX-0214  Plugin Manifest Model (plugin.json)
TASK-EX-0215  Plugin Registry (local index)
TASK-EX-0216  Plugin Loader
TASK-EX-0217  Plugin Executor (subprocess isolation)
TASK-EX-0218  Plugin Hook System (before/after install, use, etc.)
TASK-EX-0219  Plugin Install (from URL / GitHub)
TASK-EX-0220  Plugin Uninstall
TASK-EX-0221  Plugin Update
TASK-EX-0222  Plugin List
TASK-EX-0223  Plugin Version Compatibility Check
TASK-EX-0224  Plugin Sandbox (read-only access to phpvm internals)
TASK-EX-0225  Plugin Communication Protocol (stdin/stdout JSON)
TASK-EX-0226  Example Plugin: php-switcher-notify
TASK-EX-0227  Example Plugin: laravel-optimizer
TASK-EX-0228  phpvm plugin Command
TASK-EX-0229  Plugin Security Model Documentation
TASK-EX-0230  Plugin Development Guide
TASK-EX-0231  Plugin Unit Tests
TASK-EX-0232  Build Verification & Git Tag v0.12.0
```

**Execution Order:** 0213 → 0214 → 0215 → 0216 → 0217 → 0218 → 0219 → 0220 → 0221 → 0222 → 0223 → 0224 → 0225 → 0226 → 0227 → 0228 → 0229 → 0230 → 0231 → 0232

**Files Produced:**
```
internal/plugin/
├── interface.go       (TASK-EX-0213)
├── manifest.go        (TASK-EX-0214)
├── registry.go        (TASK-EX-0215)
├── loader.go          (TASK-EX-0216)
├── executor.go        (TASK-EX-0217)
├── hooks.go           (TASK-EX-0218)
├── installer.go       (TASK-EX-0219)
├── remover.go         (TASK-EX-0220)
├── updater.go         (TASK-EX-0221)
├── compat.go          (TASK-EX-0223)
├── sandbox.go         (TASK-EX-0224)
├── protocol.go        (TASK-EX-0225)
└── plugin_test.go     (TASK-EX-0231)
cmd/
└── plugin.go          (TASK-EX-0228)
examples/
├── php-switcher-notify/ (TASK-EX-0226)
└── laravel-optimizer/   (TASK-EX-0227)
docs/
├── plugin-security.md   (TASK-EX-0229)
└── plugin-development.md (TASK-EX-0230)
```

**Definition of Done:**
- [ ] Plugin interface مستند و stable است
- [ ] `phpvm plugin install https://github.com/x/y` کار می‌کند
- [ ] Plugin در subprocess ایزوله اجرا می‌شود
- [ ] هر دو example plugin اجرا می‌شوند
- [ ] `go test ./internal/plugin/...` پاس می‌شود

---

---

# PHASE-12 — Testing & Quality

**Release:** `v0.13.0`
**هدف:** پوشش تست کامل، linting، benchmark، و تضمین کیفیت برای v1.0.0.

```
TASK-EX-0233  Test Infrastructure Setup (testify, mock framework)
TASK-EX-0234  Mock: Filesystem (afero)
TASK-EX-0235  Mock: HTTP Client
TASK-EX-0236  Mock: Process Executor
TASK-EX-0237  Mock: PHP Runtime
TASK-EX-0238  Unit Tests: Config Package
TASK-EX-0239  Unit Tests: Version Resolver
TASK-EX-0240  Unit Tests: Downloader
TASK-EX-0241  Unit Tests: Archive Extractor
TASK-EX-0242  Unit Tests: PATH Manager
TASK-EX-0243  Unit Tests: php.ini Parser/Writer
TASK-EX-0244  Unit Tests: Extension Manager
TASK-EX-0245  Unit Tests: Composer
TASK-EX-0246  Unit Tests: Plugin System
TASK-EX-0247  Integration Tests: Install Flow
TASK-EX-0248  Integration Tests: Use/Switch Flow
TASK-EX-0249  Integration Tests: Composer Flow
TASK-EX-0250  Integration Tests: Doctor Flow
TASK-EX-0251  End-to-End Test: Full Install → Use → Remove Cycle
TASK-EX-0252  Benchmark Tests: Version Scanner
TASK-EX-0253  Benchmark Tests: Downloader
TASK-EX-0254  golangci-lint Configuration (.golangci.yml)
TASK-EX-0255  golangci-lint Zero Warnings Pass
TASK-EX-0256  Coverage Report (minimum 80%)
TASK-EX-0257  gosec Security Scan
TASK-EX-0258  Build Verification & Git Tag v0.13.0
```

**Execution Order:** 0233 → 0234 → 0235 → 0236 → 0237 → 0238 → 0239 → 0240 → 0241 → 0242 → 0243 → 0244 → 0245 → 0246 → 0247 → 0248 → 0249 → 0250 → 0251 → 0252 → 0253 → 0254 → 0255 → 0256 → 0257 → 0258

**Files Produced:**
```
tests/
├── mocks/
│   ├── filesystem.go  (TASK-EX-0234)
│   ├── httpclient.go  (TASK-EX-0235)
│   ├── process.go     (TASK-EX-0236)
│   └── phpruntime.go  (TASK-EX-0237)
├── integration/
│   ├── install_test.go  (TASK-EX-0247)
│   ├── use_test.go      (TASK-EX-0248)
│   ├── composer_test.go (TASK-EX-0249)
│   └── doctor_test.go   (TASK-EX-0250)
└── e2e/
    └── full_cycle_test.go (TASK-EX-0251)
.golangci.yml              (TASK-EX-0254)
```

**Definition of Done:**
- [ ] `go test -race -coverprofile=coverage.out ./...` پاس می‌شود
- [ ] Coverage بالای 80%
- [ ] `golangci-lint run` بدون warning
- [ ] `gosec ./...` بدون critical issue
- [ ] Benchmark ها documented هستند
- [ ] هیچ data race در `-race` mode

---

---

# PHASE-13 — CI/CD

**Release:** `v0.14.0`
**هدف:** Pipeline کامل GitHub Actions برای هر PR، push، و release.

```
TASK-EX-0259  Workflow: CI (lint + test + build)
TASK-EX-0260  Workflow Matrix: ubuntu-latest
TASK-EX-0261  Workflow Matrix: windows-latest
TASK-EX-0262  Workflow Matrix: macos-latest
TASK-EX-0263  Workflow: Security Scan (gosec + govulncheck)
TASK-EX-0264  Workflow: Coverage Upload (Codecov)
TASK-EX-0265  Workflow: PR Labeler
TASK-EX-0266  Workflow: Stale Issue Bot
TASK-EX-0267  Workflow: Dependency Review (Dependabot)
TASK-EX-0268  Dependabot Configuration
TASK-EX-0269  Workflow: Release (GoReleaser on tag push)
TASK-EX-0270  GoReleaser: Windows amd64 / arm64
TASK-EX-0271  GoReleaser: Linux amd64 / arm64
TASK-EX-0272  GoReleaser: macOS amd64 / arm64 (Apple Silicon)
TASK-EX-0273  GoReleaser: SHA256 Checksums File
TASK-EX-0274  GoReleaser: CHANGELOG from Conventional Commits
TASK-EX-0275  GoReleaser: GitHub Release with Assets
TASK-EX-0276  GoReleaser: Homebrew Formula (tap)
TASK-EX-0277  GoReleaser: Scoop Manifest
TASK-EX-0278  Branch Protection Rules Documentation
TASK-EX-0279  Build Verification & Git Tag v0.14.0
```

**Execution Order:** 0259 → 0260 → 0261 → 0262 → 0263 → 0264 → 0265 → 0266 → 0267 → 0268 → 0269 → 0270 → 0271 → 0272 → 0273 → 0274 → 0275 → 0276 → 0277 → 0278 → 0279

**Files Produced:**
```
.github/
├── workflows/
│   ├── ci.yml           (TASK-EX-0259–0264)
│   ├── labeler.yml      (TASK-EX-0265)
│   ├── stale.yml        (TASK-EX-0266)
│   ├── dependency-review.yml (TASK-EX-0267)
│   └── release.yml      (TASK-EX-0269–0277)
├── dependabot.yml       (TASK-EX-0268)
└── labeler.yml          (TASK-EX-0265)
.goreleaser.yaml         (TASK-EX-0270–0277)
```

**Definition of Done:**
- [ ] هر PR بدون pass شدن CI قابل merge نیست
- [ ] Build روی هر سه پلتفرم و هر دو arch موفق است
- [ ] `git tag v0.14.0 && git push --tags` یک Release کامل می‌سازد
- [ ] Homebrew و Scoop manifest تولید می‌شوند
- [ ] SHA256 checksums در Release assets موجود است

---

---

# PHASE-14 — Release & Distribution

**Release:** `v1.0.0`
**هدف:** انتشار رسمی v1.0.0 با installer، documentation کامل، و حضور در package manager ها.

```
TASK-EX-0280  Install Script: install.sh (Linux / macOS)
TASK-EX-0281  Install Script: install.ps1 (Windows PowerShell)
TASK-EX-0282  Uninstall Script: uninstall.sh
TASK-EX-0283  Uninstall Script: uninstall.ps1
TASK-EX-0284  Installer: Architecture Detection (amd64 / arm64)
TASK-EX-0285  Installer: Checksum Verification
TASK-EX-0286  Installer: PATH Configuration Post-Install
TASK-EX-0287  Installer: Welcome Message & Quick Start Guide
TASK-EX-0288  Winget Manifest Package
TASK-EX-0289  Scoop Manifest Package
TASK-EX-0290  Homebrew Tap Repository Setup
TASK-EX-0291  Homebrew Formula Final
TASK-EX-0292  README: Full Rewrite for v1.0.0
TASK-EX-0293  README: Installation Section (all methods)
TASK-EX-0294  README: Quick Start Guide
TASK-EX-0295  README: Command Reference Table
TASK-EX-0296  README: Configuration Reference
TASK-EX-0297  README: Badges (CI / Coverage / Version / License)
TASK-EX-0298  docs/: Getting Started
TASK-EX-0299  docs/: Configuration Guide
TASK-EX-0300  docs/: Command Reference
TASK-EX-0301  docs/: Platform Notes (Windows / Linux / macOS)
TASK-EX-0302  docs/: Plugin Development Guide
TASK-EX-0303  docs/: Contributing Guide
TASK-EX-0304  docs/: Privacy Policy
TASK-EX-0305  CHANGELOG: v1.0.0 Final Entry
TASK-EX-0306  Security Audit: Final gosec + govulncheck Pass
TASK-EX-0307  Performance Audit: Cold Start < 100ms
TASK-EX-0308  Binary Size Audit: < 15MB
TASK-EX-0309  Memory Usage Audit: < 50MB
TASK-EX-0310  Cross-Platform Final Smoke Test: Windows
TASK-EX-0311  Cross-Platform Final Smoke Test: Ubuntu
TASK-EX-0312  Cross-Platform Final Smoke Test: macOS
TASK-EX-0313  v1.0.0 Release Notes Draft
TASK-EX-0314  v1.0.0 GitHub Release
TASK-EX-0315  Announce: GitHub Discussions
TASK-EX-0316  Announce: README Social Links
TASK-EX-0317  Post-Release: Issue Templates
TASK-EX-0318  Post-Release: PR Template
TASK-EX-0319  Post-Release: v1.1.0 Milestone Creation
TASK-EX-0320  Post-Release: Roadmap Update (Vision 2.0)
```

**Execution Order:**
```
0280–0287 (Installers)
→ 0288–0291 (Package Managers)
→ 0292–0305 (Documentation)
→ 0306–0312 (Audits & Smoke Tests)
→ 0313–0316 (Release)
→ 0317–0320 (Post-Release)
```

**Files Produced:**
```
scripts/
├── install.sh         (TASK-EX-0280)
├── install.ps1        (TASK-EX-0281)
├── uninstall.sh       (TASK-EX-0282)
└── uninstall.ps1      (TASK-EX-0283)
packages/
├── winget/
│   └── manifest.yaml  (TASK-EX-0288)
└── scoop/
    └── phpvm.json     (TASK-EX-0289)
homebrew-phpvm/        (جداگانه — TASK-EX-0290, 0291)
README.md              (TASK-EX-0292–0297)
docs/
├── getting-started.md (TASK-EX-0298)
├── configuration.md   (TASK-EX-0299)
├── commands.md        (TASK-EX-0300)
├── platforms.md       (TASK-EX-0301)
├── plugins.md         (TASK-EX-0302)
├── contributing.md    (TASK-EX-0303)
└── privacy.md         (TASK-EX-0304)
CHANGELOG.md           (TASK-EX-0305)
.github/
├── ISSUE_TEMPLATE/    (TASK-EX-0317)
└── PULL_REQUEST_TEMPLATE.md (TASK-EX-0318)
```

**Definition of Done:**
- [ ] `curl -fsSL https://raw.githubusercontent.com/[user]/phpvm/main/scripts/install.sh | sh` کار می‌کند
- [ ] `irm https://raw.githubusercontent.com/[user]/phpvm/main/scripts/install.ps1 | iex` کار می‌کند
- [ ] Binary size زیر 15MB
- [ ] Cold start زیر 100ms
- [ ] Coverage بالای 80%
- [ ] هیچ critical security issue در gosec
- [ ] Smoke test روی هر سه پلتفرم پاس
- [ ] GitHub Release v1.0.0 با تمام assets منتشر شده است

---

---

# Complete Task Index (PHASE-06 to PHASE-14)

| Task | Phase | عنوان |
|------|-------|-------|
| TASK-EX-0106 | 06 | HTTP Client Factory |
| TASK-EX-0107 | 06 | Download Session Model |
| TASK-EX-0108 | 06 | Download Queue |
| TASK-EX-0109 | 06 | PHP Windows Source Driver |
| TASK-EX-0110 | 06 | PHP Unix Source Driver |
| TASK-EX-0111 | 06 | Mirror Fallback Strategy |
| TASK-EX-0112 | 06 | Progress Tracker |
| TASK-EX-0113 | 06 | Progress Renderer |
| TASK-EX-0114 | 06 | Resume Support |
| TASK-EX-0115 | 06 | Checksum Verifier |
| TASK-EX-0116 | 06 | Signature Verifier |
| TASK-EX-0117 | 06 | Download Cache Manager |
| TASK-EX-0118 | 06 | Retry Strategy |
| TASK-EX-0119 | 06 | Timeout Handler |
| TASK-EX-0120 | 06 | ZIP Extractor |
| TASK-EX-0121 | 06 | TAR.GZ Extractor |
| TASK-EX-0122 | 06 | Atomic Extraction |
| TASK-EX-0123 | 06 | Downloader Unit Tests |
| TASK-EX-0124 | 06 | Extractor Unit Tests |
| TASK-EX-0125 | 06 | Build Verification & Tag v0.7.0 |
| TASK-EX-0126 | 07 | PATH Entry Model |
| TASK-EX-0127 | 07 | PATH Reader |
| TASK-EX-0128 | 07 | PATH Writer |
| TASK-EX-0129 | 07 | PATH Deduplicator |
| TASK-EX-0130 | 07 | Windows Registry Reader |
| TASK-EX-0131 | 07 | Windows Registry Writer |
| TASK-EX-0132 | 07 | Windows System PATH Driver |
| TASK-EX-0133 | 07 | Unix Shell Profile Detector |
| TASK-EX-0134 | 07 | Unix Shell Profile Writer |
| TASK-EX-0135 | 07 | Unix Shell Profile Backup |
| TASK-EX-0136 | 07 | macOS Shell Profile Driver |
| TASK-EX-0137 | 07 | PHPVM Marker Block |
| TASK-EX-0138 | 07 | PATH Add Operation |
| TASK-EX-0139 | 07 | PATH Remove Operation |
| TASK-EX-0140 | 07 | PATH Verify Operation |
| TASK-EX-0141 | 07 | Shell Restart Notification |
| TASK-EX-0142 | 07 | PATH Rollback on Failure |
| TASK-EX-0143 | 07 | PATH Unit Tests (Windows) |
| TASK-EX-0144 | 07 | PATH Unit Tests (Unix) |
| TASK-EX-0145 | 07 | Build Verification & Tag v0.8.0 |
| TASK-EX-0146 | 08 | Composer Version Model |
| TASK-EX-0147 | 08 | Composer Binary Locator |
| TASK-EX-0148 | 08 | Composer Global Detector |
| TASK-EX-0149 | 08 | Composer Per-Version Detector |
| TASK-EX-0150 | 08 | Composer Installer Script Downloader |
| TASK-EX-0151 | 08 | Composer Signature Verifier |
| TASK-EX-0152 | 08 | Composer Install (Global) |
| TASK-EX-0153 | 08 | Composer Install (Per-Version) |
| TASK-EX-0154 | 08 | Composer Version Checker |
| TASK-EX-0155 | 08 | Composer Update Logic |
| TASK-EX-0156 | 08 | Composer Rollback on Failed Update |
| TASK-EX-0157 | 08 | Composer PATH Registration |
| TASK-EX-0158 | 08 | Required Extension Checker |
| TASK-EX-0159 | 08 | php.ini Compatibility Check |
| TASK-EX-0160 | 08 | Composer Diagnose Report Builder |
| TASK-EX-0161 | 08 | Composer Command |
| TASK-EX-0162 | 08 | Composer Unit Tests |
| TASK-EX-0163 | 08 | Composer Integration Tests |
| TASK-EX-0164 | 08 | Composer Documentation |
| TASK-EX-0165 | 08 | Build Verification & Tag v0.9.0 |
| TASK-EX-0166 | 09 | php.ini Locator |
| TASK-EX-0167 | 09 | php.ini Parser |
| TASK-EX-0168 | 09 | php.ini Writer |
| TASK-EX-0169 | 09 | php.ini Backup & Restore |
| TASK-EX-0170 | 09 | php.ini Diff Viewer |
| TASK-EX-0171 | 09 | Extension Model |
| TASK-EX-0172 | 09 | Extension List Scanner |
| TASK-EX-0173 | 09 | Extension Enable/Disable |
| TASK-EX-0174 | 09 | Extension Downloader |
| TASK-EX-0175 | 09 | Extension Version Compatibility |
| TASK-EX-0176 | 09 | Xdebug Version Resolver |
| TASK-EX-0177 | 09 | Xdebug Installer |
| TASK-EX-0178 | 09 | Xdebug Mode Switcher |
| TASK-EX-0179 | 09 | Xdebug php.ini Block Writer |
| TASK-EX-0180 | 09 | Doctor Check: PHP Binary |
| TASK-EX-0181 | 09 | Doctor Check: PATH & Registry |
| TASK-EX-0182 | 09 | Doctor Check: php.ini |
| TASK-EX-0183 | 09 | Doctor Check: Extensions |
| TASK-EX-0184 | 09 | Doctor Check: Composer |
| TASK-EX-0185 | 09 | Doctor Check: Permissions |
| TASK-EX-0186 | 09 | Doctor Check: Xdebug |
| TASK-EX-0187 | 09 | Doctor Auto-Fix Engine |
| TASK-EX-0188 | 09 | Doctor Report Renderer |
| TASK-EX-0189 | 09 | env / ext / xdebug / doctor Commands |
| TASK-EX-0190 | 09 | Environment Unit Tests |
| TASK-EX-0191 | 09 | Build Verification & Tag v0.10.0 |
| TASK-EX-0192 | 10 | .phpvmrc File Model |
| TASK-EX-0193 | 10 | .phpvmrc Parser |
| TASK-EX-0194 | 10 | .phpvmrc Writer |
| TASK-EX-0195 | 10 | .phpvmrc Search |
| TASK-EX-0196 | 10 | phpvm local Command |
| TASK-EX-0197 | 10 | phpvm auto Command |
| TASK-EX-0198 | 10 | Auto-Switch on Directory Change |
| TASK-EX-0199 | 10 | Framework Detector: Laravel |
| TASK-EX-0200 | 10 | Framework Detector: Symfony |
| TASK-EX-0201 | 10 | Framework Detector: WordPress |
| TASK-EX-0202 | 10 | Framework Detector: Magento |
| TASK-EX-0203 | 10 | Framework Detector: CodeIgniter |
| TASK-EX-0204 | 10 | Framework Detector: Yii |
| TASK-EX-0205 | 10 | Recommended PHP Version per Framework |
| TASK-EX-0206 | 10 | phpvm detect Command |
| TASK-EX-0207 | 10 | Shell Hook: Bash/Zsh |
| TASK-EX-0208 | 10 | Shell Hook: PowerShell |
| TASK-EX-0209 | 10 | Shell Hook: Fish |
| TASK-EX-0210 | 10 | phpvm shell Command |
| TASK-EX-0211 | 10 | Project Integration Unit Tests |
| TASK-EX-0212 | 10 | Build Verification & Tag v0.11.0 |
| TASK-EX-0213 | 11 | Plugin Interface Definition |
| TASK-EX-0214 | 11 | Plugin Manifest Model |
| TASK-EX-0215 | 11 | Plugin Registry |
| TASK-EX-0216 | 11 | Plugin Loader |
| TASK-EX-0217 | 11 | Plugin Executor |
| TASK-EX-0218 | 11 | Plugin Hook System |
| TASK-EX-0219 | 11 | Plugin Install |
| TASK-EX-0220 | 11 | Plugin Uninstall |
| TASK-EX-0221 | 11 | Plugin Update |
| TASK-EX-0222 | 11 | Plugin List |
| TASK-EX-0223 | 11 | Plugin Version Compatibility |
| TASK-EX-0224 | 11 | Plugin Sandbox |
| TASK-EX-0225 | 11 | Plugin Communication Protocol |
| TASK-EX-0226 | 11 | Example Plugin: php-switcher-notify |
| TASK-EX-0227 | 11 | Example Plugin: laravel-optimizer |
| TASK-EX-0228 | 11 | phpvm plugin Command |
| TASK-EX-0229 | 11 | Plugin Security Model Docs |
| TASK-EX-0230 | 11 | Plugin Development Guide |
| TASK-EX-0231 | 11 | Plugin Unit Tests |
| TASK-EX-0232 | 11 | Build Verification & Tag v0.12.0 |
| TASK-EX-0233 | 12 | Test Infrastructure Setup |
| TASK-EX-0234 | 12 | Mock: Filesystem |
| TASK-EX-0235 | 12 | Mock: HTTP Client |
| TASK-EX-0236 | 12 | Mock: Process Executor |
| TASK-EX-0237 | 12 | Mock: PHP Runtime |
| TASK-EX-0238 | 12 | Unit Tests: Config |
| TASK-EX-0239 | 12 | Unit Tests: Version Resolver |
| TASK-EX-0240 | 12 | Unit Tests: Downloader |
| TASK-EX-0241 | 12 | Unit Tests: Archive |
| TASK-EX-0242 | 12 | Unit Tests: PATH Manager |
| TASK-EX-0243 | 12 | Unit Tests: php.ini |
| TASK-EX-0244 | 12 | Unit Tests: Extension Manager |
| TASK-EX-0245 | 12 | Unit Tests: Composer |
| TASK-EX-0246 | 12 | Unit Tests: Plugin System |
| TASK-EX-0247 | 12 | Integration Tests: Install Flow |
| TASK-EX-0248 | 12 | Integration Tests: Use/Switch Flow |
| TASK-EX-0249 | 12 | Integration Tests: Composer Flow |
| TASK-EX-0250 | 12 | Integration Tests: Doctor Flow |
| TASK-EX-0251 | 12 | E2E: Full Install → Use → Remove |
| TASK-EX-0252 | 12 | Benchmark: Version Scanner |
| TASK-EX-0253 | 12 | Benchmark: Downloader |
| TASK-EX-0254 | 12 | golangci-lint Configuration |
| TASK-EX-0255 | 12 | golangci-lint Zero Warnings Pass |
| TASK-EX-0256 | 12 | Coverage Report ≥ 80% |
| TASK-EX-0257 | 12 | gosec Security Scan |
| TASK-EX-0258 | 12 | Build Verification & Tag v0.13.0 |
| TASK-EX-0259 | 13 | Workflow: CI |
| TASK-EX-0260 | 13 | Workflow Matrix: ubuntu |
| TASK-EX-0261 | 13 | Workflow Matrix: windows |
| TASK-EX-0262 | 13 | Workflow Matrix: macos |
| TASK-EX-0263 | 13 | Workflow: Security Scan |
| TASK-EX-0264 | 13 | Workflow: Coverage Upload |
| TASK-EX-0265 | 13 | Workflow: PR Labeler |
| TASK-EX-0266 | 13 | Workflow: Stale Bot |
| TASK-EX-0267 | 13 | Workflow: Dependency Review |
| TASK-EX-0268 | 13 | Dependabot Configuration |
| TASK-EX-0269 | 13 | Workflow: Release |
| TASK-EX-0270 | 13 | GoReleaser: Windows |
| TASK-EX-0271 | 13 | GoReleaser: Linux |
| TASK-EX-0272 | 13 | GoReleaser: macOS |
| TASK-EX-0273 | 13 | GoReleaser: SHA256 Checksums |
| TASK-EX-0274 | 13 | GoReleaser: CHANGELOG |
| TASK-EX-0275 | 13 | GoReleaser: GitHub Release |
| TASK-EX-0276 | 13 | GoReleaser: Homebrew Formula |
| TASK-EX-0277 | 13 | GoReleaser: Scoop Manifest |
| TASK-EX-0278 | 13 | Branch Protection Docs |
| TASK-EX-0279 | 13 | Build Verification & Tag v0.14.0 |
| TASK-EX-0280 | 14 | Install Script: install.sh |
| TASK-EX-0281 | 14 | Install Script: install.ps1 |
| TASK-EX-0282 | 14 | Uninstall Script: uninstall.sh |
| TASK-EX-0283 | 14 | Uninstall Script: uninstall.ps1 |
| TASK-EX-0284 | 14 | Installer: Architecture Detection |
| TASK-EX-0285 | 14 | Installer: Checksum Verification |
| TASK-EX-0286 | 14 | Installer: PATH Configuration |
| TASK-EX-0287 | 14 | Installer: Welcome Message |
| TASK-EX-0288 | 14 | Winget Manifest |
| TASK-EX-0289 | 14 | Scoop Manifest |
| TASK-EX-0290 | 14 | Homebrew Tap Setup |
| TASK-EX-0291 | 14 | Homebrew Formula Final |
| TASK-EX-0292 | 14 | README Full Rewrite |
| TASK-EX-0293 | 14 | README: Installation Section |
| TASK-EX-0294 | 14 | README: Quick Start Guide |
| TASK-EX-0295 | 14 | README: Command Reference |
| TASK-EX-0296 | 14 | README: Configuration Reference |
| TASK-EX-0297 | 14 | README: Badges |
| TASK-EX-0298 | 14 | docs: Getting Started |
| TASK-EX-0299 | 14 | docs: Configuration Guide |
| TASK-EX-0300 | 14 | docs: Command Reference |
| TASK-EX-0301 | 14 | docs: Platform Notes |
| TASK-EX-0302 | 14 | docs: Plugin Development |
| TASK-EX-0303 | 14 | docs: Contributing Guide |
| TASK-EX-0304 | 14 | docs: Privacy Policy |
| TASK-EX-0305 | 14 | CHANGELOG v1.0.0 |
| TASK-EX-0306 | 14 | Security Audit Final |
| TASK-EX-0307 | 14 | Performance Audit |
| TASK-EX-0308 | 14 | Binary Size Audit |
| TASK-EX-0309 | 14 | Memory Usage Audit |
| TASK-EX-0310 | 14 | Smoke Test: Windows |
| TASK-EX-0311 | 14 | Smoke Test: Ubuntu |
| TASK-EX-0312 | 14 | Smoke Test: macOS |
| TASK-EX-0313 | 14 | v1.0.0 Release Notes |
| TASK-EX-0314 | 14 | v1.0.0 GitHub Release |
| TASK-EX-0315 | 14 | Announce: GitHub Discussions |
| TASK-EX-0316 | 14 | Announce: README Social Links |
| TASK-EX-0317 | 14 | Post-Release: Issue Templates |
| TASK-EX-0318 | 14 | Post-Release: PR Template |
| TASK-EX-0319 | 14 | Post-Release: v1.1.0 Milestone |
| TASK-EX-0320 | 14 | Post-Release: Roadmap Update |

---

## Grand Total

| | Count |
|--|--|
| Phases (06–14) | 9 |
| Tasks (EX-0106–0320) | **215** |
| Tasks (EX-0001–0105, از فایل قبلی) | **105** |
| **Total Tasks** | **320** |

---

*نسخه سند: 1.0 | ادامه از TASKS.md*
