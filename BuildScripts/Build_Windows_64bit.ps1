$env:GOOS = "windows"
$env:GOARCH="amd64"
go build -o "./bin/" -ldflags "-s -w" ..
#Reset to defaults
$env:GOOS=''
$env:GOARCH=''
Write-Output 'Windows(amd64) Finished.'
