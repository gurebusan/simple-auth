Simple-auth простое приложение авторизации выполненное в рамках тестового задания.
Приложение выдаёт пользователю при авторизации пару токенов (access и refresh), которые используются для дальнейших запросов.
access токен выдается в формате JWT в теле ответа, refresh токен - случайная строка, хэш которой хранится в базе данных.
refresh токен выдается пользователю в куки.
Приложение завернуто в docker-контейнер.
Код требует доработки, но уже в рабочем состоянии.

Комментарии:
Для маршрутов использовал два POST запроса: /auth и /refresh.
При попытке пользователем выполнить авторизации повторно refresh токен обновляется, а access токен выдается заново
Истекшие по времени refresh токены пока что не удаляются из базы данных автоматически, однако при попытке использовать истекший refresh токен он удалится, а
пользователю необходимо будет выполнить повторную авторизацию.
Так как refresh токен это произвольная строка, то у него нет payload, а ip пользователя хранится в базе,
В payload access токена хранится guid пользователя и время жизни токена.
Прошу дать фидбек по коду по ключевым моментам.
