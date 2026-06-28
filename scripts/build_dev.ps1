$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\GitHub CLI;" + $env:PATH

$commit = (git rev-parse --short HEAD 2>$null)
if (-not $commit) { $commit = "none" }
$date = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
$module = "github.com/mahtdy/phpvm/pkg/version"

$ldflags = "-s -w -X `"${module}.Version=0.1.0`" -X `"${module}.Commit=${commit}`" -X `"${module}.Date=${date}`" -X `"${module}.BuiltBy=make`""

New-Item -ItemType Directory -Path "dist" -Force | Out-Null
go build -ldflags $ldflags -o dist/phpvm.exe .
