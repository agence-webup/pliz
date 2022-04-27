all: darwin-amd64 darwin-arm64 linux-amd64 windows-amd64

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o build/pliz_darwin_amd64 main.go

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o build/pliz_darwin_arm64 main.go

linux-amd64:
	GOOS=darwin GOARCH=amd64 go build -o build/pliz_linux_amd64 main.go

windows-amd64:
	GOOS=darwin GOARCH=amd64 go build -o build/pliz_windows_amd64.exe main.go