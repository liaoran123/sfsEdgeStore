@echo off
set EDGEX_DB_PATH=./edgex_data_test
echo Starting sfsDb EdgeX adapter...
go run main.go
pause
