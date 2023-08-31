# Segment-api

## Запуск

Для запуска приложения без таблиц в корне репозитория необходимо выполнить команду:
```
make start
```
Чтобы запустить с автоматически созданными таблицами :
```
make start-with-migrations
```
Остановка приложения:
```
make stop
```

## Описание

### Что использовал при разработке:
1) `go-chi` - HTTP роутер
2) `pgx` - для работы с PostgreSQL
3) `cleanenv` - для работы с переменными окружения
4) `uber-go/zap` - для логирования
5) `migrate` - для работы с миграциями
6) `testcontainers-go` - для запуска PostgreSQL с интеграционными тестами
7) `testfixtures` -  для написания интеграционных тестов.
8) `testify` - для unit тестов
9) `gomock` - для генерации mock-структур
10) `squirrel` - для написания SQL запросов

## Схема базы данных 
![image](https://github.com/VrMolodyakov/segment-api/assets/99216816/602f6a51-0c33-430d-8ac6-da87941ea482)


# Swagger
[swagger файл](docs/swagger.yaml)


## API Reference

### Созание пользователя

```
  POST http://localhost:8080/api/v1/users
```

Тело запроса

```
{
  "firsName": "John", //required
  "lastName": "Doe",  //required
  "email": "example@example.com" //required
}
```
Ответ
```
{
    "userID": 1,
    "firsName": "John",
    "lastName": "Doe",
    "email": "example@example.com"
}

```
Email должен быть уникальным

Возможная ошибка
```
{"ok":false,"message":"User already exists"}
```

### Созание Сегмента

Принимает название и процент автоматического попадание в этот сегмент. Если hitPercentage не установлен , то автоматически в него не попасть

```
  POST http://localhost:8080/api/v1/segments
```

Тело запроса

```
{
  "name": "test_segment_2", //required
  "hitPercentage": 10
}
```
Ответ
```
{
    "segmentID": 1,
    "name": "test_segment",
    "hitPercentage": 10
}
```
Возможная ошибка
```
{"ok":false,"message":"Segment already exists"}
```


### Добавление/удаление сегментов у пользователя

Принимает id пользоватедя и списки на обновление/удаление. Если ttl не задан , то сегмент будет закреплен за пользователем до удаления.
Обязательно либо update, либо delete не должны быть пустыми.

```
  POST http://localhost:8080/api/v1/membership/update
```

Тело запроса

```
{
    "userID": 1,
    "update": [
        {
            "name": "test_name_1", //required
            "ttl": 3600
        },
        {
            "name": "test_name_2" //required
        }
    ],
    "delete": [
        {
            "name": "test_name_3" //required
        },
        {
            "name": "test_name_4" //required
        }
    ]
}
```
Ответ
```
200 OK
```
Возможная ошибки

```
{"ok":false,"message":"Data for update and delete cannot be empty at the same time"}
```
```
{"ok":false,"message":"Attempt to add segments that the user already belongs to"}
```
```
{"ok":false,"message":"Attempt to update the data of a non-existent user"}
```
```
{"ok":false,"message":"Attempt to add and remove the same segment"}
```
```
{"ok":false,"message":"Attempt to delete a segment unassigned to the user"}
```
```
{"ok":false,"message":"Not all segments with the specified names were found or adding/removing one segment multiple times"}
```


### Получение сегментов пользовтеля

```
  Get http://localhost:8080/api/v1/users/{userID}
```

Ответ
```
{
    "memberships": [
        {
            "userID": 1,
            "segmentName": "test_name_1",
            "expiredAt": "2023-08-31T18:43:33.262977+03:00"
        },
        {
            "userID": 1,
            "segmentName": "test_name_2",
            "expiredAt": "9999-01-01T04:59:59+03:00"
        }
    ]
}
```
Возможная ошибка
```
{"ok":false,"message":"No data was found for the specified user"}
```

### Удаление сегмента

```
  DELETE http://localhost:8080/api/v1/segments/{segmentName}
```

Ответ
```
200 OK
```
Возможная ошибка
```
{"ok":false,"message":"Segment with the specified name wasn't found"}
```


### Создание ссылки на историю сегментов

```
  POST http://localhost:8080/api/v1/history/link
```
Тело запроса

```
{
  "year": 2023,
  "month": 8
}

```

Ответ
```
{
    "link": "http://localhost:8080/api/v1/history/download/2023/8"
}
```

Возможная ошибка
```
{"ok":false,"message":"Incorrect date, history for dates before 2007 year is not available"}
```

### Скачать данные по ссылке

```
GET http://localhost:8080/api/v1/history/download/{year}/{month}
```
Пример 
```
ID,UserID,Segment,Operation,Time
1,test_name_1,added,2023-08-31 17:43:33
1,test_name_2,added,2023-08-31 17:43:33

```

Возможная ошибка
```
{"ok":false,"message":"Data lifetime for the link has expired, create a new one"}
```
