$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\GitHub CLI;" + $env:PATH
go get golang.org/x/sys@latest
go mod tidy
