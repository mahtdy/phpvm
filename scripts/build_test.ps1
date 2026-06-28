$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\GitHub CLI;" + $env:PATH
$env:CGO_ENABLED = "0"
Write-Host "=== Building ==="
go build ./...
if ($LASTEXITCODE -ne 0) { Write-Host "BUILD FAILED"; exit 1 }
Write-Host "=== Running tests ==="
go test -v ./...
if ($LASTEXITCODE -ne 0) { Write-Host "TESTS FAILED"; exit 1 }
Write-Host "=== All OK ==="
