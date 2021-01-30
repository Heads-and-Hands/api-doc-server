GOOS=linux GOARCH=amd64 go build -o api-doc-server-2
scp api-doc-server-2 root@apidoc.handh.ru:/
