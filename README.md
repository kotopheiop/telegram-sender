# Telegram-Sender

## Описание проекта
Telegram-Sender - это учебный проект, реализованный в рамках изучения языка Go. Это сервис, который запускается в Docker и принимает POST-запросы для отправки сообщений в Telegram.

## Как это работает
Сервис принимает POST-запросы, в заголовках которых должен быть `ChatId` (обязательно) и `MessageThreadID` (опционально).

В теле запроса должно быть сообщение, которое необходимо отправить.

Сервис может хранить сообщения в очереди и при неудачной отправке например из-за ошибки с кодом 429 (Too many request) от API Telegram переслать это сообщение позже, тем самым гарантируя отправку сообщения.

## Подготовка к запуску
Скачайте проект командой
```bash
git clone https://github.com/kotopheiop/telegram-sender.git
```

Перед запуском в файл `docker-compose.yml` нужно прописать токен бота:
```yaml
version: '3.3'
services:
  telegram-sender:
    build:
      context: .
      dockerfile: Dockerfile
    image: kotopheiop/telegram-sender
    container_name: <project-name>-app # Вместо <project-name> укажите имя своего проекта
    ports:
      - "8095:8080"
    restart: unless-stopped
    environment:
      - BOT_TOKEN=<your_bot_token> # Укажите здесь токен своего бота
    healthcheck:
      test: [ "CMD", "curl", "--fail", "http://localhost:8080/health", "||", "exit", "1" ]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Запуск проекта
Проект запускается с помощью команды 
```bash
docker-compose up --build
```

## Пример использования
После запуска сервиса вы можете отправить POST-запрос на `http://localhost:port/send`, где `port` - это порт, на котором работает ваш сервис, по умолчанию в `docker-compose.yml` прописан 8095. 

В заголовках запроса укажите `ChatId` и, если необходимо, `MessageThreadID`. В теле запроса укажите сообщение, которое вы хотите отправить.

Примеры запросов:
<details> <summary>PHP - cURL</summary>

```php
        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_URL => 'localhost:8095/api/send',
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_ENCODING => '',
            CURLOPT_MAXREDIRS => 10,
            CURLOPT_TIMEOUT => 0,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_HTTP_VERSION => CURL_HTTP_VERSION_1_1,
            CURLOPT_CUSTOMREQUEST => 'POST',
            CURLOPT_POSTFIELDS => 'Тест',
            CURLOPT_HTTPHEADER => array(
                'ChatId: you-chat-id',
                'MessageThreadID: message-thread-id',
                'Content-Type: text/plain'
            ),
        ));

        $response = curl_exec($curl);

        curl_close($curl);
        echo $response;
```
</details>

<details> <summary>PHP - Guzzle</summary>

```php
        $client = new Client();
        $headers = [
            'ChatId' => 'you-chat-id',
            'MessageThreadID' => 'message-thread-id',
            'Content-Type' => 'text/plain'
        ];
        $body = 'Тест';
        $request = new Request('POST', 'localhost:8095/api/send', $headers, $body);
        $res = $client->sendAsync($request)->wait();
        echo $res->getBody();
```
</details>

<details> <summary>NodeJS - Native</summary>

```js
        var https = require('follow-redirects').https;
        var fs = require('fs');
        
        var options = {
            'method': 'POST',
            'hostname': 'localhost',
            'port': 8095,
            'path': '/api/send',
            'headers': {
                'ChatId': 'you-chat-id',
                'MessageThreadID': 'message-thread-id',
                'Content-Type': 'text/plain'
            },
            'maxRedirects': 20
        };
        
        var req = https.request(options, function (res) {
            var chunks = [];
        
            res.on("data", function (chunk) {
                chunks.push(chunk);
            });
        
            res.on("end", function (chunk) {
                var body = Buffer.concat(chunks);
                console.log(body.toString());
            });
        
            res.on("error", function (error) {
                console.error(error);
            });
        });
        
        var postData =  "Тест";
        
        req.write(postData);
        
        req.end();
```
</details>

<details> <summary>NodeJS - Axios</summary>

```js
        const axios = require('axios');
        let data = 'Тест';
        
        let config = {
            method: 'post',
            maxBodyLength: Infinity,
            url: 'localhost:8095/api/send',
            headers: {
                'ChatId': 'you-chat-id',
                'MessageThreadID': 'message-thread-id',
                'Content-Type': 'text/plain'
            },
            data : data
        };
        
        axios.request(config)
            .then((response) => {
                console.log(JSON.stringify(response.data));
            })
            .catch((error) => {
                console.log(error);
            });
```
</details>

## Проверка работоспособности сервиса
Вы можете проверить работоспособность сервиса, отправив GET-запрос на `http://localhost:port/health`. В ответе вернётся текущая дата и время.

## Важно
Убедитесь, что у вас установлен Docker и docker-compose перед запуском этого проекта.