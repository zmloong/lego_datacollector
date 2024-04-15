@echo "开始编译lego_datacollector"
set GOOS=linux
set GO111MODULE=auto
go build -o ./bin/datacollector ./services/datacollector/main.go
REM pause