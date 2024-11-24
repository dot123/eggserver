@echo off

cd ./bin

chcp 65001
redis-cli --raw -h 127.0.0.1 -p 6379
