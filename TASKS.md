# PHPVM — Project Tasks & Roadmap

> PHP Version Manager | Professional CLI Platform
> Inspired by NVM, Volta, Rustup, and Composer

---

## Project Metadata

| Field | Value |
|---|---|
| Project Name | phpvm |
| Repository | github.com/[your-username]/phpvm |
| Language | Go 1.22+ |
| CLI Framework | Cobra |
| Config | Viper |
| Logger | slog (stdlib) |
| Tests | Go Test + testify |
| CI/CD | GitHub Actions |
| Release | GoReleaser |
| Installer | Shell script + Winget + Scoop |
| License | MIT |
| Platforms | Windows / Linux / macOS |

---

## Conventions

- هر Task کاملاً مستقل و دارای یک مسئولیت (Single Responsibility) است
- هر Task بدون بازنویسی تسک‌های قبلی قابل اجراست
- همه exported functions داکیومنت دارند
- هر تابع عمومی حداقل یک unit test دارد
- خطاها با `fmt.Errorf("context: %w", err)` wrap می‌شوند
- هیچ `panic` در production code مجاز نیست
- همه magic string ها به constant تبدیل می‌شوند

---

## Repository Structure

```
phpvm/
├── cmd/
│   ├── root.go
│   ├── use.go
│   ├── list.go
│   ├── install.go
│   ├── remove.go
│   ├── current.go
│   ├── doctor.go
│   ├── env.go
│   ├── composer.go
│   └── update.go
├── internal/
│   ├── cli/
│   ├── config/
│   ├── downloader/
│   ├── archive/
│   ├── php/
│   ├── composer/
│   ├── extension/
│   ├── ini/
│   ├── path/
│   ├── process/
│   ├── logger/
│   ├── validation/
│   └── version/
├── pkg/
├── configs/
├── scripts/
│   ├── install.sh
│   └── install.ps1
├── docs/
├── tests/
├── .github/
│   └── workflows/
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yaml
├── LICENSE
└── README.md
```

---

## GitHub Setup Checklist

قبل از شروع کدنویسی این موارد باید انجام شود:

- [ ] ساخت repo با نام `phpvm` روی GitHub (public)
- [ ] اضافه کردن توضیح: `PHP Version Manager — Cross-platform CLI tool`
- [ ] اضافه کردن Topics: `php`, `version-manager`, `cli`, `golang`, `developer-tools`
- [ ] Branch protection روی `main` (require PR + 1 review)
- [ ] ساخت branch `develop` به عنوان branch توسعه
- [ ] اضافه کردن Secrets برای GitHub Actions:
  - `GITHUB_TOKEN` (اتوماتیک)
  - `GORELEASER_KEY` (اگر از Pro استفاده می‌شود)
- [ ] اضافه کردن LICENSE فایل (MIT)
- [ ] git remote add origin و push اولیه

---

---

# PHASE 1 — Core CLI Foundation

**هدف فاز:** ایجاد پایه‌های محکم پروژه. بعد از این فاز باید بتوان `phpvm version` و `phpvm help` را اجرا کرد.

**خروجی فاز:** یک binary کامپایل‌شده که روی هر سه پلتفرم اجرا می‌شود.

---

## TASK-001 — Repository Initialization

**هدف:** ساخت ساختار حرفه‌ای repository از صفر.

**Acceptance Criteria:**
- [ ] `go.mod` با module path `github.com/[username]/phpvm` ساخته شده
- [ ] Cobra و Viper به عنوان dependency اضافه شده‌اند
- [ ] ساختار کامل پوشه‌ها طبق معماری بالا ایجاد شده
- [ ] `Makefile` با targetsهای `build`, `test`, `lint`, `clean`, `release` ساخته شده
- [ ] `LICENSE` فایل MIT ساخته شده
- [ ] `README.md` اولیه با badge و توضیح پروژه ساخته شده
- [ ] `.gitignore` مخصوص Go ساخته شده
- [ ] `.goreleaser.yaml` برای build سه پلتفرم تنظیم شده
- [ ] `pkg/version/version.go` با متغیرهای `Version`, `Commit`, `Date` ساخته شده

**فایل‌های مرتبط:**
```
go.mod
go.sum
Makefile
LICENSE
README.md
.gitignore
.goreleaser.yaml
pkg/version/version.go
```

**Dependencies:** هیچ (اولین تسک)

**Definition of Done:**
- `go build ./...` بدون خطا اجرا می‌شود
- `go test ./...` بدون خطا اجرا می‌شود
- تمام فایل‌های فهرست‌شده وجود دارند

---

## TASK-002 — Configuration System

**هدف:** سیستم load و save تنظیمات با استفاده از Viper.

**Config Schema:**
```json
{
  "root": "~/.phpvm",
  "current": "8.3.0",
  "default_arch": "x64",
  "telemetry": false
}
```

**Acceptance Criteria:**
- [ ] `internal/config/config.go` ساختار `Config` را تعریف می‌کند
- [ ] Config از فایل `~/.phpvm/config.json` خوانده می‌شود
- [ ] اگر فایل وجود نداشت، مقادیر default استفاده می‌شوند
- [ ] `SaveConfig()` تغییرات را persist می‌کند
- [ ] متغیرهای محیطی با prefix `PHPVM_` تنظیمات را override می‌کنند
- [ ] `PHPVM_ROOT` مسیر root را تغییر می‌دهد
- [ ] Unit tests برای load, save, و default values

**فایل‌های مرتبط:**
```
internal/config/config.go
internal/config/config_test.go
```

**Dependencies:** TASK-001

**Definition of Done:**
- `go test ./internal/config/...` پاس می‌شود
- Config در هر سه پلتفرم به درستی load می‌شود

---

## TASK-003 — CLI Bootstrap

**هدف:** راه‌اندازی Cobra و دستورات پایه.

**دستورات:**
```
phpvm            # نمایش help
phpvm --help     # help کامل
phpvm version    # نمایش نسخه binary
phpvm help       # help
```

**Acceptance Criteria:**
- [ ] `cmd/root.go` با Cobra راه‌اندازی شده
- [ ] `phpvm version` نمایش می‌دهد: `phpvm v0.1.0 (commit: abc123, built: 2025-01-01)`
- [ ] `phpvm --help` لیست تمام دستورات را نشان می‌دهد
- [ ] `main.go` فقط `cmd.Execute()` را صدا می‌زند
- [ ] Global flags: `--debug`, `--no-color`, `--config`
- [ ] Exit code 0 برای موفقیت، 1 برای خطا

**فایل‌های مرتبط:**
```
main.go
cmd/root.go
cmd/version.go
```

**Dependencies:** TASK-001, TASK-002

**Definition of Done:**
- `./phpvm version` خروجی صحیح می‌دهد
- `./phpvm --help` بدون خطا اجرا می‌شود

---

## TASK-004 — Logger System

**هدف:** سیستم logging یکپارچه با استفاده از `slog`.

**Log Levels:**
```
DEBUG   # اطلاعات توسعه‌دهنده
INFO    # عملیات عادی
WARN    # هشدارها
ERROR   # خطاهای قابل بازیابی
```

**Acceptance Criteria:**
- [ ] `internal/logger/logger.go` با wrapper بر slog ساخته شده
- [ ] فرمت human-readable برای terminal (رنگی)
- [ ] فرمت JSON برای فایل log
- [ ] Log level با `--debug` flag یا `PHPVM_LOG_LEVEL=debug` تغییر می‌کند
- [ ] رنگ‌ها با `--no-color` غیرفعال می‌شوند
- [ ] Logger به صورت singleton در دسترس کل برنامه است
- [ ] Unit tests برای هر level

**فایل‌های مرتبط:**
```
internal/logger/logger.go
internal/logger/logger_test.go
```

**Dependencies:** TASK-001

**Definition of Done:**
- `go test ./internal/logger/...` پاس می‌شود
- Log های رنگی در terminal نمایش داده می‌شوند

---

## TASK-005 — Error Handler

**هدف:** سیستم یکپارچه مدیریت خطا در کل برنامه.

**انواع خطا:**
```go
ErrPermission   // دسترسی ناکافی
ErrNetwork      // مشکل اتصال
ErrFileNotFound // فایل یافت نشد
ErrArchive      // خطای استخراج
ErrPathConflict // تداخل در PATH
ErrRegistry     // خطای Windows Registry
ErrValidation   // ورودی نامعتبر
ErrNotInstalled // نسخه نصب نشده
```

**Acceptance Criteria:**
- [ ] `internal/cli/errors.go` تمام error types را تعریف می‌کند
- [ ] هر خطا پیام کاربرپسند به فارسی/انگلیسی دارد
- [ ] هر خطا exit code مناسب دارد
- [ ] `errors.Is()` و `errors.As()` پشتیبانی می‌شود
- [ ] خطاها stack trace در debug mode نمایش می‌دهند
- [ ] Unit tests برای هر error type

**فایل‌های مرتبط:**
```
internal/cli/errors.go
internal/cli/errors_test.go
```

**Dependencies:** TASK-004

**Definition of Done:**
- `go test ./internal/cli/...` پاس می‌شود
- خطاها پیام واضح و exit code صحیح دارند

---

---

# PHASE 2 — PHP Version Management

**هدف فاز:** مدیریت نسخه‌های PHP که از قبل نصب شده‌اند. بعد از این فاز `phpvm list`, `phpvm current`, `phpvm use` کار می‌کنند.

**خروجی فاز:** کاربر می‌تواند بین نسخه‌های نصب‌شده سوئیچ کند.

---

## TASK-006 — Scan Installed Versions

**هدف:** اسکن پوشه root و لیست کردن نسخه‌های نصب‌شده.

**دستور:**
```
phpvm list
```

**خروجی نمونه:**
```
Installed PHP versions:
  8.2.20
  8.3.10
→ 8.4.1  (current)
  8.5.0
```

**Acceptance Criteria:**
- [ ] `internal/php/scanner.go` پوشه `~/.phpvm/versions/` را اسکن می‌کند
- [ ] هر subdirectory که شامل `php` یا `php.exe` باشد به عنوان نسخه شناخته می‌شود
- [ ] نسخه current با `→` مشخص می‌شود
- [ ] اگر هیچ نسخه‌ای نصب نباشد پیام راهنما نمایش می‌دهد
- [ ] خروجی قابل sort بر اساس semver است
- [ ] `cmd/list.go` این functionality را expose می‌کند
- [ ] Unit tests با mock filesystem

**فایل‌های مرتبط:**
```
internal/php/scanner.go
internal/php/scanner_test.go
cmd/list.go
```

**Dependencies:** TASK-003, TASK-005

**Definition of Done:**
- `phpvm list` لیست صحیح نمایش می‌دهد
- `go test ./internal/php/...` پاس می‌شود

---

## TASK-007 — Current Version Detection

**هدف:** تشخیص نسخه فعال PHP.

**دستور:**
```
phpvm current
```

**خروجی نمونه:**
```
8.4.1
```

**Acceptance Criteria:**
- [ ] `internal/php/current.go` نسخه فعال را از config و PATH تشخیص می‌دهد
- [ ] اگر `php` در PATH نباشد خطای مناسب می‌دهد
- [ ] نسخه از خروجی `php --version` هم verify می‌شود
- [ ] `cmd/current.go` این functionality را expose می‌کند
- [ ] Unit tests با mock

**فایل‌های مرتبط:**
```
internal/php/current.go
internal/php/current_test.go
cmd/current.go
```

**Dependencies:** TASK-006

**Definition of Done:**
- `phpvm current` نسخه صحیح را نمایش می‌دهد

---

## TASK-008 — Version Validation

**هدف:** اعتبارسنجی ورودی نسخه PHP.

**Acceptance Criteria:**
- [ ] `internal/validation/version.go` ورودی را validate می‌کند
- [ ] فرمت‌های معتبر: `8.3`, `8.3.10`, `8.3.10-nts`, `8.3.10-ts`
- [ ] فرمت‌های نامعتبر پیام خطای واضح می‌دهند
- [ ] Semver comparison: `8.3 < 8.4 < 8.5`
- [ ] تشخیص نسخه‌های پشتیبانی‌شده PHP (8.0+)
- [ ] Unit tests برای حالت‌های edge case

**فایل‌های مرتبط:**
```
internal/validation/version.go
internal/validation/version_test.go
```

**Dependencies:** TASK-005

**Definition of Done:**
- `go test ./internal/validation/...` پاس می‌شود

---

## TASK-009 — Switch Version (use)

**هدف:** سوئیچ بین نسخه‌های نصب‌شده.

**دستور:**
```
phpvm use 8.4
phpvm use 8.4.1
```

**Acceptance Criteria:**
- [ ] `internal/php/switcher.go` سوئیچ نسخه را انجام می‌دهد
- [ ] اگر نسخه نصب نباشد پیام پیشنهاد install می‌دهد
- [ ] Config به‌روز می‌شود
- [ ] PATH به‌روز می‌شود (تسک بعدی)
- [ ] Symlink یا PATH update انجام می‌شود
- [ ] `cmd/use.go` این functionality را expose می‌کند
- [ ] پیام موفقیت: `Switched to PHP 8.4.1`

**فایل‌های مرتبط:**
```
internal/php/switcher.go
internal/php/switcher_test.go
cmd/use.go
```

**Dependencies:** TASK-007, TASK-008

**Definition of Done:**
- `phpvm use 8.4` بدون خطا اجرا می‌شود
- `php --version` بعد از سوئیچ نسخه درست را نشان می‌دهد

---

## TASK-010 — PATH Manager

**هدف:** مدیریت PATH بدون خراب کردن تنظیمات موجود.

**Acceptance Criteria:**
- [ ] `internal/path/manager.go` مدیریت PATH را انجام می‌دهد
- [ ] **Windows:** Registry `HKCU\Environment` به‌روز می‌شود
- [ ] **Linux/macOS:** به `.bashrc`, `.zshrc`, `.profile` اضافه می‌شود
- [ ] entry قدیمی phpvm از PATH حذف می‌شود قبل از اضافه کردن جدید
- [ ] PATH های دیگر دست نخورده می‌مانند
- [ ] `AddToPath(dir string)` و `RemoveFromPath(dir string)` پیاده‌سازی شده‌اند
- [ ] پیام راهنما برای restart shell
- [ ] Unit tests با mock برای هر پلتفرم

**فایل‌های مرتبط:**
```
internal/path/manager.go
internal/path/manager_windows.go
internal/path/manager_unix.go
internal/path/manager_test.go
```

**Dependencies:** TASK-009

**Definition of Done:**
- PATH بعد از `phpvm use` به درستی به‌روز می‌شود
- PATH های موجود دست نخورده‌اند

---

---

# PHASE 3 — Downloader & Installer

**هدف فاز:** دانلود و نصب نسخه‌های PHP از اینترنت. بعد از این فاز `phpvm install 8.5` کار می‌کند.

**خروجی فاز:** کاربر می‌تواند هر نسخه PHP را با یک دستور نصب کند.

---

## TASK-011 — PHP Release Resolver

**هدف:** دریافت لیست نسخه‌های موجود از php.net و ساخت download URL.

**منابع:**
- Windows: `https://windows.php.net/downloads/releases/`
- Linux/macOS: `https://www.php.net/releases/index.php?json&version=8.x`

**Acceptance Criteria:**
- [ ] `internal/version/resolver.go` لیست نسخه‌های موجود را fetch می‌کند
- [ ] Cache نتایج برای 1 ساعت در `~/.phpvm/cache/`
- [ ] تشخیص پلتفرم و arch (x64, arm64)
- [ ] پشتیبانی از هر دو نوع **TS** (Thread Safe) و **NTS** (Non-Thread Safe)
- [ ] ساخت download URL صحیح برای هر ترکیب (OS + arch + TS/NTS + version)
- [ ] خطای مناسب اگر نسخه موجود نباشد
- [ ] Unit tests با mock HTTP

**فایل‌های مرتبط:**
```
internal/version/resolver.go
internal/version/resolver_test.go
internal/version/urls.go
```

**Dependencies:** TASK-005

**Definition of Done:**
- URL صحیح برای هر ترکیب پلتفرم/نسخه برگردانده می‌شود
- Cache کار می‌کند

---

## TASK-012 — Downloader

**هدف:** دانلود فایل با progress bar، retry، و checksum verification.

**Acceptance Criteria:**
- [ ] `internal/downloader/downloader.go` دانلود را مدیریت می‌کند
- [ ] Progress bar در terminal نمایش داده می‌شود (با سرعت و ETA)
- [ ] دانلود قابل resume است (اگر فایل ناقص بود)
- [ ] حداکثر 3 بار retry در صورت خطای شبکه
- [ ] SHA256 checksum verification بعد از دانلود
- [ ] دانلود در `~/.phpvm/downloads/` ذخیره می‌شود
- [ ] Timeout configurable (default: 30 دقیقه)
- [ ] Unit tests با mock HTTP server

**فایل‌های مرتبط:**
```
internal/downloader/downloader.go
internal/downloader/downloader_test.go
internal/downloader/progress.go
```

**Dependencies:** TASK-011

**Definition of Done:**
- دانلود با progress bar کار می‌کند
- Checksum verify می‌شود
- Retry در صورت خطا کار می‌کند

---

## TASK-013 — Archive Extractor

**هدف:** استخراج فایل‌های zip (Windows) و tar.gz (Linux/macOS).

**Acceptance Criteria:**
- [ ] `internal/archive/extractor.go` استخراج را مدیریت می‌کند
- [ ] پشتیبانی از `.zip` برای Windows
- [ ] پشتیبانی از `.tar.gz` برای Linux/macOS
- [ ] Progress نمایش داده می‌شود
- [ ] استخراج به مسیر `~/.phpvm/versions/[version]/`
- [ ] فایل‌های موجود overwrite می‌شوند
- [ ] Atomic extraction (ابتدا به temp، سپس rename)
- [ ] Unit tests با test archives

**فایل‌های مرتبط:**
```
internal/archive/extractor.go
internal/archive/extractor_test.go
internal/archive/zip.go
internal/archive/targz.go
```

**Dependencies:** TASK-012

**Definition of Done:**
- `go test ./internal/archive/...` پاس می‌شود
- استخراج روی هر سه پلتفرم کار می‌کند

---

## TASK-014 — Install Command

**هدف:** دستور `phpvm install` که تمام مراحل را ترکیب می‌کند.

**دستورات:**
```
phpvm install 8.5
phpvm install 8.5.0
phpvm install 8.5.0 --ts        # Thread Safe
phpvm install 8.5.0 --nts       # Non-Thread Safe (default)
phpvm install 8.5.0 --arm64     # ARM architecture
phpvm install latest            # آخرین نسخه stable
```

**Acceptance Criteria:**
- [ ] `cmd/install.go` دستور install را پیاده‌سازی می‌کند
- [ ] مراحل: Resolve → Download → Verify → Extract → Configure
- [ ] اگر نسخه از قبل نصب باشد پیام مناسب می‌دهد
- [ ] Flag `--force` برای نصب مجدد
- [ ] بعد از نصب پیش‌نهاد `phpvm use [version]` می‌دهد
- [ ] نصب به صورت atomic (rollback در صورت خطا)
- [ ] Integration test

**فایل‌های مرتبط:**
```
cmd/install.go
internal/php/installer.go
internal/php/installer_test.go
```

**Dependencies:** TASK-013, TASK-010

**Definition of Done:**
- `phpvm install 8.4` نسخه را دانلود و نصب می‌کند
- `phpvm use 8.4` بعد از نصب کار می‌کند

---

## TASK-015 — Remove Command

**هدف:** حذف نسخه‌های نصب‌شده PHP.

**دستور:**
```
phpvm remove 8.3
phpvm remove 8.3.10
```

**Acceptance Criteria:**
- [ ] `cmd/remove.go` دستور remove را پیاده‌سازی می‌کند
- [ ] تأیید از کاربر قبل از حذف (مگر با `--yes`)
- [ ] نمی‌توان نسخه current را حذف کرد
- [ ] پوشه نسخه به طور کامل حذف می‌شود
- [ ] Config به‌روز می‌شود
- [ ] Unit tests

**فایل‌های مرتبط:**
```
cmd/remove.go
internal/php/remover.go
internal/php/remover_test.go
```

**Dependencies:** TASK-014

**Definition of Done:**
- `phpvm remove 8.3` نسخه را حذف می‌کند
- نسخه current قابل حذف نیست

---

## TASK-016 — Update Command

**هدف:** به‌روزرسانی نسخه‌های نصب‌شده به آخرین patch version.

**دستورات:**
```
phpvm update         # به‌روزرسانی نسخه current
phpvm update 8.3     # به‌روزرسانی یک نسخه خاص
phpvm update --all   # به‌روزرسانی همه نسخه‌ها
```

**Acceptance Criteria:**
- [ ] `cmd/update.go` دستور update را پیاده‌سازی می‌کند
- [ ] بررسی آخرین patch version از php.net
- [ ] نصب نسخه جدید و حذف قدیمی (با تأیید)
- [ ] اگر نسخه‌ای آخرین باشد پیام مناسب می‌دهد
- [ ] Unit tests

**فایل‌های مرتبط:**
```
cmd/update.go
internal/php/updater.go
internal/php/updater_test.go
```

**Dependencies:** TASK-014, TASK-015

**Definition of Done:**
- `phpvm update` نسخه current را به آخرین patch به‌روز می‌کند

---

---

# PHASE 4 — Composer Integration

**هدف فاز:** مدیریت کامل Composer. بعد از این فاز کاربر می‌تواند Composer را نصب، به‌روز، و تشخیص دهد.

**خروجی فاز:** `phpvm composer install` و `phpvm composer update` کار می‌کنند.

---

## TASK-017 — Composer Detection

**هدف:** تشخیص وجود و نسخه Composer در سیستم.

**Acceptance Criteria:**
- [ ] `internal/composer/detector.go` Composer را در PATH پیدا می‌کند
- [ ] نسخه Composer از `composer --version` خوانده می‌شود
- [ ] تشخیص Composer per-PHP-version (در پوشه PHP نصب‌شده)
- [ ] تشخیص Composer global (در PATH سیستم)
- [ ] Unit tests با mock

**فایل‌های مرتبط:**
```
internal/composer/detector.go
internal/composer/detector_test.go
```

**Dependencies:** TASK-007

**Definition of Done:**
- `go test ./internal/composer/...` پاس می‌شود

---

## TASK-018 — Composer Installer

**هدف:** دانلود و نصب آخرین نسخه Composer.

**دستور:**
```
phpvm composer install
```

**Acceptance Criteria:**
- [ ] `internal/composer/installer.go` Composer را نصب می‌کند
- [ ] دانلود از `https://getcomposer.org/installer`
- [ ] Verify signature قبل از اجرا
- [ ] نصب در کنار PHP version فعال
- [ ] `composer` در PATH قرار می‌گیرد
- [ ] پشتیبانی از نصب global یا per-version
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/composer/installer.go
internal/composer/installer_test.go
cmd/composer.go
```

**Dependencies:** TASK-017, TASK-012

**Definition of Done:**
- `phpvm composer install` Composer را نصب می‌کند
- `composer --version` بعد از نصب کار می‌کند

---

## TASK-019 — Composer Update

**هدف:** به‌روزرسانی Composer به آخرین نسخه.

**دستور:**
```
phpvm composer update
```

**Acceptance Criteria:**
- [ ] `internal/composer/updater.go` Composer را به‌روز می‌کند
- [ ] بررسی آخرین نسخه از getcomposer.org
- [ ] اگر آخرین نسخه نصب باشد پیام مناسب می‌دهد
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/composer/updater.go
internal/composer/updater_test.go
```

**Dependencies:** TASK-018

**Definition of Done:**
- `phpvm composer update` Composer را به‌روز می‌کند

---

## TASK-020 — Composer Diagnose

**هدف:** بررسی سلامت Composer و PHP.

**دستور:**
```
phpvm composer diagnose
```

**Acceptance Criteria:**
- [ ] بررسی: Composer نصب است؟
- [ ] بررسی: PHP version سازگار با Composer است؟
- [ ] بررسی: Extensions موردنیاز Composer موجود است؟ (json, phar, filter, hash, iconv, mbstring, openssl)
- [ ] بررسی: `allow_url_fopen` در php.ini فعال است؟
- [ ] گزارش رنگی با ✓ و ✗
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/composer/diagnose.go
internal/composer/diagnose_test.go
```

**Dependencies:** TASK-019

**Definition of Done:**
- `phpvm composer diagnose` گزارش کامل نمایش می‌دهد

---

---

# PHASE 5 — PHP Environment Management

**هدف فاز:** مدیریت کامل محیط PHP شامل php.ini، extensions، و Xdebug.

**خروجی فاز:** `phpvm doctor` محیط کامل PHP را بررسی می‌کند.

---

## TASK-021 — php.ini Manager

**هدف:** مدیریت تنظیمات php.ini به صورت برنامه‌نویسی.

**دستورات:**
```
phpvm env get memory_limit
phpvm env set memory_limit 512M
phpvm env list
phpvm env reset
```

**Acceptance Criteria:**
- [ ] `internal/ini/parser.go` فایل php.ini را parse می‌کند
- [ ] خواندن، نوشتن، و reset مقادیر پشتیبانی می‌شود
- [ ] نسخه‌بندی php.ini (backup قبل از تغییر)
- [ ] تشخیص php.ini هر نسخه PHP
- [ ] `cmd/env.go` این functionality را expose می‌کند
- [ ] Unit tests با test ini files

**فایل‌های مرتبط:**
```
internal/ini/parser.go
internal/ini/parser_test.go
internal/ini/writer.go
cmd/env.go
```

**Dependencies:** TASK-009

**Definition of Done:**
- `phpvm env set memory_limit 512M` php.ini را به‌روز می‌کند
- `go test ./internal/ini/...` پاس می‌شود

---

## TASK-022 — Extension Manager

**هدف:** فعال/غیرفعال کردن PHP extensions.

**دستورات:**
```
phpvm ext list              # لیست همه extensions
phpvm ext enable redis
phpvm ext disable xdebug
phpvm ext install imagick   # دانلود و نصب
phpvm ext status openssl    # وضعیت یک extension
```

**Extensions پشتیبانی‌شده (اولویت):**
```
openssl, curl, intl, redis, imagick,
pdo_mysql, pdo_pgsql, mbstring, zip,
gd, soap, xml, json, bcmath
```

**Acceptance Criteria:**
- [ ] `internal/extension/manager.go` مدیریت extensions را انجام می‌دهد
- [ ] تشخیص extensions فعال از php.ini
- [ ] فعال/غیرفعال با تغییر php.ini (کامنت/آن‌کامنت)
- [ ] دانلود extension از PECL یا منبع رسمی
- [ ] تطابق نسخه extension با نسخه PHP
- [ ] `cmd/ext.go` (یا sub-command) این functionality را expose می‌کند
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/extension/manager.go
internal/extension/manager_test.go
internal/extension/installer.go
cmd/ext.go
```

**Dependencies:** TASK-021

**Definition of Done:**
- `phpvm ext enable redis` php.ini را به‌روز می‌کند
- `phpvm ext list` لیست صحیح نمایش می‌دهد

---

## TASK-023 — Xdebug Manager

**هدف:** مدیریت Xdebug به صورت مستقل.

**دستورات:**
```
phpvm xdebug enable
phpvm xdebug disable
phpvm xdebug status
phpvm xdebug install
phpvm xdebug switch develop   # حالت develop
phpvm xdebug switch coverage  # حالت coverage
phpvm xdebug switch off       # خاموش
```

**Acceptance Criteria:**
- [ ] `internal/extension/xdebug.go` مدیریت Xdebug را انجام می‌دهد
- [ ] نصب نسخه سازگار Xdebug با PHP version فعال
- [ ] پشتیبانی از حالت‌های: develop, coverage, profile, trace
- [ ] تنظیم خودکار `xdebug.mode` در php.ini
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/extension/xdebug.go
internal/extension/xdebug_test.go
cmd/xdebug.go
```

**Dependencies:** TASK-022

**Definition of Done:**
- `phpvm xdebug enable` Xdebug را فعال می‌کند
- `php -m | grep xdebug` نتیجه صحیح نشان می‌دهد

---

## TASK-024 — Doctor Command

**هدف:** بررسی کامل سلامت محیط PHP در یک دستور.

**دستور:**
```
phpvm doctor
```

**خروجی نمونه:**
```
phpvm Doctor — Environment Check
─────────────────────────────────
✓ PHP 8.4.1 installed and active
✓ PATH configured correctly
✓ php.ini found and readable
✓ Composer 2.7.0 installed
✗ Xdebug not installed
⚠ memory_limit is 128M (recommended: 256M+)
✓ openssl extension enabled
✓ mbstring extension enabled
✗ redis extension not found

Issues found: 2
Run 'phpvm doctor --fix' to auto-fix issues
```

**Acceptance Criteria:**
- [ ] `internal/php/doctor.go` تمام چک‌ها را انجام می‌دهد
- [ ] چک‌ها: PHP, PATH, Registry(Win), php.ini, Extensions, Composer, Permissions
- [ ] خروجی رنگی با ✓ ✗ ⚠
- [ ] Flag `--fix` برای رفع خودکار مشکلات قابل‌رفع
- [ ] Flag `--json` برای خروجی JSON
- [ ] `cmd/doctor.go` این functionality را expose می‌کند
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/php/doctor.go
internal/php/doctor_test.go
cmd/doctor.go
```

**Dependencies:** TASK-021, TASK-022, TASK-020

**Definition of Done:**
- `phpvm doctor` گزارش کامل نمایش می‌دهد
- `phpvm doctor --fix` مشکلات قابل‌رفع را اصلاح می‌کند

---

---

# PHASE 6 — Developer Experience

**هدف فاز:** قابلیت‌هایی که PHPVM را از یک tool ساده به یک platform تبدیل می‌کند.

**خروجی فاز:** تجربه developer-first کامل با auto-detection و shell integration.

---

## TASK-025 — Project Local Version (.phpvmrc)

**هدف:** پشتیبانی از نسخه‌بندی per-project شبیه به `.nvmrc`.

**مثال:**
```
# .phpvmrc
8.4
```

**دستور:**
```
phpvm auto              # تشخیص و سوئیچ خودکار
phpvm local 8.4         # ست کردن نسخه پروژه
phpvm local --unset     # حذف .phpvmrc
```

**Acceptance Criteria:**
- [ ] `internal/php/local.go` فایل `.phpvmrc` را جستجو و parse می‌کند
- [ ] جستجو از پوشه فعلی به سمت root (مثل git)
- [ ] `phpvm auto` نسخه را بر اساس `.phpvmrc` سوئیچ می‌کند
- [ ] اگر `.phpvmrc` پیدا نشد به نسخه default برمی‌گردد
- [ ] `phpvm local` فایل `.phpvmrc` را می‌سازد
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/php/local.go
internal/php/local_test.go
cmd/auto.go
cmd/local.go
```

**Dependencies:** TASK-009

**Definition of Done:**
- `phpvm auto` در پوشه‌ای با `.phpvmrc` نسخه را سوئیچ می‌کند

---

## TASK-026 — Framework Detection

**هدف:** تشخیص خودکار فریم‌ورک PHP پروژه.

**دستور:**
```
phpvm detect
```

**خروجی نمونه:**
```
Detected: Laravel 11.x
Recommended PHP: 8.2+
Current PHP: 8.4.1 ✓
```

**فریم‌ورک‌های پشتیبانی‌شده:**
```
Laravel    → composer.json: require.laravel/framework
Symfony    → composer.json: require.symfony/framework-bundle
WordPress  → wp-config.php یا wp-includes/
Magento    → app/etc/di.xml
CodeIgniter → system/CodeIgniter.php
Yii        → yii
```

**Acceptance Criteria:**
- [ ] `internal/php/detector.go` فریم‌ورک را تشخیص می‌دهد
- [ ] بررسی `composer.json` برای framework-specific packages
- [ ] بررسی فایل‌های signature پروژه
- [ ] نمایش PHP version مناسب برای فریم‌ورک تشخیص‌داده‌شده
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/php/detector.go
internal/php/detector_test.go
cmd/detect.go
```

**Dependencies:** TASK-025

**Definition of Done:**
- `phpvm detect` در پروژه Laravel نتیجه صحیح نمایش می‌دهد

---

## TASK-027 — Shell Integration

**هدف:** یکپارچه‌سازی با shell برای سوئیچ خودکار نسخه.

**Shells:**
```
PowerShell (Windows)
CMD (Windows)
Bash (Linux/macOS)
Zsh (macOS/Linux)
Fish
Git Bash (Windows)
```

**دستور:**
```
phpvm shell install     # نصب integration برای shell فعلی
phpvm shell uninstall   # حذف integration
phpvm shell status      # وضعیت integration
```

**Acceptance Criteria:**
- [ ] `internal/cli/shell.go` shell integration را مدیریت می‌کند
- [ ] تشخیص shell فعلی
- [ ] اضافه کردن hook به shell config (چک `.phpvmrc` هر بار `cd`)
- [ ] **PowerShell:** تغییر `$PROFILE`
- [ ] **Bash/Zsh:** تغییر `.bashrc` / `.zshrc`
- [ ] نمایش دستورالعمل نصب دستی اگر automation ممکن نبود
- [ ] Unit tests

**فایل‌های مرتبط:**
```
internal/cli/shell.go
internal/cli/shell_test.go
scripts/phpvm.ps1
scripts/phpvm.sh
cmd/shell.go
```

**Dependencies:** TASK-025

**Definition of Done:**
- بعد از نصب، ورود به پوشه با `.phpvmrc` نسخه را خودکار سوئیچ می‌کند

---

## TASK-028 — Shell Autocomplete

**هدف:** پشتیبانی از TAB completion در همه shell‌ها.

**دستور:**
```
phpvm completion bash    # خروجی script برای bash
phpvm completion zsh     # خروجی script برای zsh
phpvm completion fish    # خروجی script برای fish
phpvm completion powershell  # خروجی script برای PowerShell
```

**Acceptance Criteria:**
- [ ] Cobra completion برای تمام sub-commands
- [ ] Dynamic completion برای نسخه‌های نصب‌شده در `phpvm use [TAB]`
- [ ] Dynamic completion برای نسخه‌های موجود در `phpvm install [TAB]`
- [ ] دستورالعمل نصب در `phpvm completion --help`
- [ ] Unit tests

**فایل‌های مرتبط:**
```
cmd/completion.go
```

**Dependencies:** TASK-003

**Definition of Done:**
- TAB completion در bash و zsh کار می‌کند
- نسخه‌های نصب‌شده در completion نمایش داده می‌شوند

---

---

# PHASE 7 — Release & Enterprise

**هدف فاز:** آماده‌سازی برای انتشار عمومی روی GitHub و قابلیت‌های enterprise.

**خروجی فاز:** Release v1.0.0 روی GitHub با installer برای هر سه پلتفرم.

---

## TASK-029 — GitHub Actions CI

**هدف:** Pipeline کامل CI برای هر Pull Request.

**Jobs:**
```
lint     → golangci-lint
test     → go test -race -cover
build    → go build برای هر سه پلتفرم
security → gosec
```

**Acceptance Criteria:**
- [ ] `.github/workflows/ci.yml` ساخته شده
- [ ] روی هر PR و push به `main` و `develop` اجرا می‌شود
- [ ] Matrix build: `ubuntu-latest`, `windows-latest`, `macos-latest`
- [ ] Test coverage report آپلود می‌شود
- [ ] Build artifacts ذخیره می‌شوند
- [ ] Badge‌های CI در README

**فایل‌های مرتبط:**
```
.github/workflows/ci.yml
.github/workflows/security.yml
```

**Dependencies:** TASK-005 (همه تست‌ها باید وجود داشته باشند)

**Definition of Done:**
- PR بدون pass شدن CI merge نمی‌شود
- Build روی هر سه پلتفرم موفق است

---

## TASK-030 — GoReleaser Pipeline

**هدف:** ساخت خودکار Release برای هر سه پلتفرم با یک Git tag.

**Targets:**
```
Windows  amd64   → phpvm_windows_amd64.zip
Windows  arm64   → phpvm_windows_arm64.zip
Linux    amd64   → phpvm_linux_amd64.tar.gz
Linux    arm64   → phpvm_linux_arm64.tar.gz
macOS    amd64   → phpvm_darwin_amd64.tar.gz
macOS    arm64   → phpvm_darwin_arm64.tar.gz  (Apple Silicon)
```

**Acceptance Criteria:**
- [ ] `.goreleaser.yaml` برای تمام targets تنظیم شده
- [ ] `.github/workflows/release.yml` روی tag push اجرا می‌شود
- [ ] Checksums file (`SHA256SUMS`) تولید می‌شود
- [ ] CHANGELOG خودکار از Conventional Commits تولید می‌شود
- [ ] GitHub Release با assets و release notes ایجاد می‌شود
- [ ] Homebrew formula تولید می‌شود (برای macOS/Linux)
- [ ] Scoop manifest تولید می‌شود (برای Windows)

**فایل‌های مرتبط:**
```
.goreleaser.yaml
.github/workflows/release.yml
```

**Dependencies:** TASK-029

**Definition of Done:**
- `git tag v0.1.0 && git push --tags` یک Release کامل می‌سازد

---

## TASK-031 — Installers

**هدف:** ساخت installer برای هر پلتفرم که کاربر با یک دستور PHPVM را نصب کند.

**دستورات نصب:**
```bash
# Linux/macOS
curl -fsSL https://phpvm.dev/install.sh | sh

# Windows PowerShell
irm https://phpvm.dev/install.ps1 | iex

# Winget
winget install phpvm

# Scoop
scoop install phpvm

# Homebrew
brew install phpvm
```

**Acceptance Criteria:**
- [ ] `scripts/install.sh` برای Linux/macOS (pure bash)
- [ ] `scripts/install.ps1` برای Windows PowerShell
- [ ] هر دو script: دانلود binary مناسب، verify checksum، نصب در PATH
- [ ] پیام خوشامدگویی و دستورالعمل شروع سریع بعد از نصب
- [ ] Uninstall نیز پشتیبانی می‌شود

**فایل‌های مرتبط:**
```
scripts/install.sh
scripts/install.ps1
scripts/uninstall.sh
scripts/uninstall.ps1
```

**Dependencies:** TASK-030

**Definition of Done:**
- Install script با یک دستور PHPVM را نصب می‌کند
- `phpvm version` بعد از نصب کار می‌کند

---

## TASK-032 — Plugin System (Foundation)

**هدف:** زیرساخت اولیه برای plugin system.

**Acceptance Criteria:**
- [ ] `internal/plugin/loader.go` plugin interface را تعریف می‌کند
- [ ] Plugins به صورت executable در `~/.phpvm/plugins/` قرار می‌گیرند
- [ ] `phpvm plugin list` لیست plugins را نمایش می‌دهد
- [ ] `phpvm plugin install [url]` یک plugin را نصب می‌کند
- [ ] Plugin lifecycle: `init`, `run`, `cleanup`
- [ ] Documentation کامل برای plugin developers

**فایل‌های مرتبط:**
```
internal/plugin/loader.go
internal/plugin/loader_test.go
cmd/plugin.go
docs/plugin-development.md
```

**Dependencies:** TASK-003

**Definition of Done:**
- یک sample plugin نوشته و test شده است
- `phpvm plugin list` کار می‌کند

---

## TASK-033 — Telemetry (Optional, Opt-in)

**هدف:** جمع‌آوری اختیاری آمار استفاده (به‌صورت پیش‌فرض غیرفعال).

**Acceptance Criteria:**
- [ ] **به‌صورت پیش‌فرض کاملاً غیرفعال است**
- [ ] کاربر باید صریحاً opt-in کند: `phpvm telemetry enable`
- [ ] `phpvm telemetry status` وضعیت را نشان می‌دهد
- [ ] `phpvm telemetry disable` غیرفعال می‌کند
- [ ] فقط: OS, command name, و نسخه phpvm ارسال می‌شود
- [ ] هیچ PII ارسال نمی‌شود
- [ ] Privacy policy در docs

**فایل‌های مرتبط:**
```
internal/telemetry/telemetry.go
cmd/telemetry.go
docs/privacy.md
```

**Dependencies:** TASK-003

**Definition of Done:**
- Telemetry به‌صورت پیش‌فرض OFF است
- Opt-in کار می‌کند

---

---

# PHASE 0 — GitHub Setup (قبل از هر چیز)

> این فاز باید قبل از شروع Phase 1 انجام شود.

---

## TASK-000 — GitHub Repository Setup

**هدف:** آماده‌سازی repository روی GitHub برای کار تیمی و CI/CD.

**مراحل:**

### ۱. ساخت Repository
```
Name: phpvm
Description: PHP Version Manager — Cross-platform CLI tool for managing PHP versions
Visibility: Public
Initialize: بدون README (خودمان می‌سازیم)
License: MIT
```

### ۲. Branch Strategy
```
main      → فقط Release و hotfix
develop   → توسعه روزانه
feature/* → هر Task یک branch جداگانه
```

### ۳. Branch Protection (main)
```
- Require pull request reviews: 1
- Require status checks to pass
- Require branches to be up to date
- Do not allow force pushes
```

### ۴. دسترسی‌های مورد نیاز برای Kiro
برای اینکه بتوانم کدها را push کنم:
```
Option A: GitHub CLI
  gh auth login
  (دادن token با scope: repo, workflow)

Option B: SSH Key
  کپی کردن public key به GitHub Settings > SSH Keys

Option C: Personal Access Token
  GitHub Settings > Developer Settings > PAT > Classic
  Scopes: repo (full), workflow
```

### ۵. Git Config محلی
```bash
git config --global user.name "Your Name"
git config --global user.email "your@email.com"
```

**Definition of Done:**
- [ ] Repository روی GitHub ساخته شده
- [ ] Branch protection فعال است
- [ ] Kiro می‌تواند push کند (یکی از سه گزینه بالا)
- [ ] `git clone git@github.com:[username]/phpvm.git` کار می‌کند

---

---

# Task Summary

| Task | Phase | عنوان | Dependencies |
|------|-------|-------|-------------|
| TASK-000 | 0 | GitHub Repository Setup | — |
| TASK-001 | 1 | Repository Initialization | TASK-000 |
| TASK-002 | 1 | Configuration System | TASK-001 |
| TASK-003 | 1 | CLI Bootstrap | TASK-001, TASK-002 |
| TASK-004 | 1 | Logger System | TASK-001 |
| TASK-005 | 1 | Error Handler | TASK-004 |
| TASK-006 | 2 | Scan Installed Versions | TASK-003, TASK-005 |
| TASK-007 | 2 | Current Version Detection | TASK-006 |
| TASK-008 | 2 | Version Validation | TASK-005 |
| TASK-009 | 2 | Switch Version (use) | TASK-007, TASK-008 |
| TASK-010 | 2 | PATH Manager | TASK-009 |
| TASK-011 | 3 | PHP Release Resolver | TASK-005 |
| TASK-012 | 3 | Downloader | TASK-011 |
| TASK-013 | 3 | Archive Extractor | TASK-012 |
| TASK-014 | 3 | Install Command | TASK-013, TASK-010 |
| TASK-015 | 3 | Remove Command | TASK-014 |
| TASK-016 | 3 | Update Command | TASK-014, TASK-015 |
| TASK-017 | 4 | Composer Detection | TASK-007 |
| TASK-018 | 4 | Composer Installer | TASK-017, TASK-012 |
| TASK-019 | 4 | Composer Update | TASK-018 |
| TASK-020 | 4 | Composer Diagnose | TASK-019 |
| TASK-021 | 5 | php.ini Manager | TASK-009 |
| TASK-022 | 5 | Extension Manager | TASK-021 |
| TASK-023 | 5 | Xdebug Manager | TASK-022 |
| TASK-024 | 5 | Doctor Command | TASK-021, TASK-022, TASK-020 |
| TASK-025 | 6 | Project Local Version (.phpvmrc) | TASK-009 |
| TASK-026 | 6 | Framework Detection | TASK-025 |
| TASK-027 | 6 | Shell Integration | TASK-025 |
| TASK-028 | 6 | Shell Autocomplete | TASK-003 |
| TASK-029 | 7 | GitHub Actions CI | همه تست‌ها |
| TASK-030 | 7 | GoReleaser Pipeline | TASK-029 |
| TASK-031 | 7 | Installers | TASK-030 |
| TASK-032 | 7 | Plugin System (Foundation) | TASK-003 |
| TASK-033 | 7 | Telemetry (Opt-in) | TASK-003 |

**مجموع: 34 Task در 8 فاز**

---

# Vision 2.0

بعد از انتشار v1.0.0 stable:

| Feature | توضیح |
|---------|-------|
| FPM Manager | مدیریت PHP-FPM pools |
| Web Server Manager | Apache و Nginx configuration |
| Laravel Installer | نصب و مدیریت Laravel installer |
| Laravel Herd Compatibility | سازگاری با Herd |
| Docker Integration | php-fpm container management |
| SSL Certificate Manager | Local SSL با mkcert |
| PHP Build from Source | کامپایل PHP از source |
| Extension Marketplace | مارکت‌پلیس extension |
| Plugin Marketplace | مارکت‌پلیس plugin |
| GUI Desktop | رابط گرافیکی با Wails + Go |

---

# Commit Convention

همه commit ها از [Conventional Commits](https://www.conventionalcommits.org/) پیروی می‌کنند:

```
feat(install): add PHP 8.5 support
fix(path): prevent duplicate PATH entries on Windows
docs(readme): add installation instructions
test(downloader): add retry mechanism tests
chore(deps): update cobra to v1.9.0
refactor(config): simplify config loading
```

---

# Definition of "Production Ready"

قبل از Release v1.0.0:

- [ ] Test coverage بالای 80%
- [ ] هیچ `panic` در production code
- [ ] همه exported functions documented
- [ ] golangci-lint بدون warning
- [ ] Binary size زیر 15MB
- [ ] Memory usage زیر 50MB در عادی‌ترین حالت
- [ ] Cold start زیر 100ms
- [ ] README کامل با examples
- [ ] CHANGELOG کامل
- [ ] تست‌شده روی Windows 10/11، Ubuntu 22.04، macOS 13+

---

*آخرین به‌روزرسانی: بر اساس توافق اولیه پروژه*
*نسخه سند: 1.0*
