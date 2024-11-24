rm -f ./bin.zip
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./bootstrap -tags lambda.norpc -tags nomsgpack ./cmd/eggserver/main.go
zip -r bin.zip ./bootstrap ./data/conf/* ./data/static/*
rm -f ./bootstrap
