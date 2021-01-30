### API-DOC-SERVER
Простенький http-сервер для отображение статики, адрес которой берется из складывания переменной окружение ROOT_PATH
и HTTP-заголовка host.
Например, мы запустили его как ```ROOT_PATH="/var/www/common-dir/" ./api-doc-server``` и обратились к нему
по адресу test.server.handh.ru:8383 (порт 8383 по умолчанию зашит в код).
Сервер вернет нам файл по адресу /var/www/common-dir/test.server.handh.ru/index.html

### Сборка
GOOS=linux GOARCH=amd64 go build -o api-doc-server