$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\GitHub CLI;" + $env:PATH
$env:CGO_ENABLED = "0"
go test -v ./... 2>&1
