@echo off
chcp 65001

taskkill /F /IM redis-server.exe

cd ./bin

del /f /s /q *.log

start redis-server.exe .\redis.conf
