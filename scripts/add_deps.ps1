$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\GitHub CLI;" + $env:PATH
# progress bar library
go get github.com/schollz/progressbar/v3@v3.14.4
go mod tidy
