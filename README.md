контакты для связи -  телеграмм @ilyaVinokurow

Оглавление 
### Команды для тестирования в терминале
### Структура Работы программы 
### Структура проекта (файлы и папки)
### Как запустить
### Что сделано и не сделано



# Команды для тестирования в терминале

Посылаем значение выражения (exp)
curl -v -d "exp=13" -X POST http://localhost:8000/newtask

!!!!! Есть проблемма при отправке через консоль символа + он заменяется на пробел
в браузере такой проблеммы нет 
curl -v -d "exp=2/2*(2+2)" -X POST http://localhost:8000/newtask
В результате выражение не принимается и считается не валидным и мы получает ответ 400
Решение вводить выражение с плюсов в браузере

get запрос
curl -v -X GET http://localhost:8000/getAgentList

### Структура Работы программы 
Всё общение идёт через http протокол  и Post запросы
(основные части front/front.go agent/agent.go  server/server.go)

Фронт работает с пользоватаелем (через html страницу)обрабатывает запросы и выдаёт результат.
фронт находится на localhost:8000
Фронт передаёт серверу(оркестратору) Таски (выражения для вычисления) и получает ответы от сервера.

Сервер передаёт выражение на вычесление агенту.
Сервер находится на localhost:8010

Агент создаёт новый поток и выполняет в нём выражение.
(начальный агент находится на localhost:8020 )
Количество потоков ограничено количеством потоков процесоора на котором запущен агент

<img src="./srv.png" height="700"/></h1>

# Структура проекта (файлы и папки)
Проект содержит файлы модуля, модуль называется -  last
go.mod  go.sum 
В файле go.mod можно увидеть версию go и необходимые зависимости

README.md srv.png - файлы редми

Папки
/cmd
Из этой папки запускается программа, есть 2 варианта: как монолит 
go run .\cmd\main.go
 и по отдлеьности частями:

/agentmain
go run .\cmd\agentmain\main.go
/frontmain
go run .\cmd\frontmain\main.go

/servermain
go run .\cmd\servermain\main.go  


/agent
Содержит сам агент, который запускается из  .\cmd\main.go

/front
Содержит фронт, который запускается из  .\cmd\main.go
/front/template
Сщдуржит шаблоны(html) которые нужны для взаимоделйствияс пользователем 

/server
Содержит сервер(для оркестрирования) , который запускается из  .\cmd\main.go

# Взаимодействие компонентов (фронта сервера и агента)
    Пользователь передаёт данные в фронт(обрабатывает запросы и посылает нужные запросы сервверу)
    Общаются через пост запросы передачей json
    Сервер(оркестратор) получает таски от фронта и отдаёт агенту, как только получил от него запрос

# Как запустить 

Для запуска нужно go 1.16 ; git ; скачать go get github.com/overseven/go-math-expression-parser

Запуск проводился в терминале в VScode в Windows 10
Данная программа запускается как монолит и не разделена на подпрограммы 
запускается 
go run .\cmd\main.go
Можно также запустить и раздельно фронт сервер и агента
по отдельности фронт (go run .\cmd\frontmain\main.go) сервер (go run .\cmd\servermain\main.go) агента (go run .\cmd\agentmain\main.go )
!!Важно закрывать нужно через Ctl + C так обрабатываются сочетание клавиш и идёт сохраниение в файл, есть задержка перез закрытием 10 миллисекунд

в проекте использую модуль с именем last(нужно проверить включены ли модули 
go env

должно быть 
set GO111MODULE=
или
set GO111MODULE=on)
команда для изменения переменной go env -w GO111MODULE=
 
программа запускалась с версией  go 1.16
со старой может поломаться
что бы запустить со старой надо изменить версию в модуле (файл go.mod)
скачать недостающие зависимости  go mod tidy

если не скачались зависимости то необходимо скачать пакет для вычисления выражения 
github.com/overseven/go-math-expression-parser
go get github.com/overseven/go-math-expression-parser



# Запросы


# Что сделано и не сделано
0 readme с описанием программы и как запустить 

1 Программа запускается и все примеры с вычислением арифметических выражений корректно работают - 10 баллов
    рограмма запускается, можно запустить двумя способами: сразу всю программу(go run .\cmd\main.go), 
    по отдельности фронт (go run .\cmd\frontmain\main.go) сервер (go run .\cmd\servermain\main.go) агента (go run .\cmd\agentmain\main.go )
    !!Важно закрывать нужно через Ctl + C так обрабатываются сочетание клавиш и идёт сохраниение в файл, есть задержка перез закрытием 10 миллисекунд
    при запуске агента можно указать порт на котором он будет выполняться 
    Зарезервировнные порты 8000 - для фронта  8010 - для сервера, 
    Если запускать на одном порту то будут ошибки и ничего не запустится

2 Программа запускается и выполняются произвольные примеры с вычислением арифметических выражений - 10 баллов
    Примеры желательно задавать через форму, елс и через консоль то могут быть проблемы со знаком + (заменяется на пробел) 
    Как вычисляются и передаются таски можно проследить по логам

3 Можно перезапустить любой компонент системы и система корректно обработает перезапуск (результаты сохранены, система продолжает работать) - 10 баллов
    фронт не перезапускается, другие компоненты перезапускаются с сохранением состоянием, 
    состояние сохраняется  в текстовый файл в формате json

4 Система предосталяет графический интерфейс для вычисления арифметических выражений - 10 баллов
    Граффический интерфейс расположен по адресу http://localhost:8000/

5 Реализован мониторинг воркеров - 20 баллов
    есть мониторинг воркеров(у меня агенты) http://localhost:8000/servers
    Но нет отображения задач которые выполняются на них

6 Реализован интерфейс для мориторинга воркеров - 10 баллов
    (Не совсем понял что нужно сделать )
    Реализовал выдачу списка работающих агентов curl -v -X GET http://localhost:8000/getAgentList

7 Вам понятна кодовая база и структура проекта - 10 баллов (это субъективный критерий, но чем проще ваше решение - тем лучше).
Проверяющий в этом пункте честно отвечает на вопрос: "Смогу я сделать пулл-реквест в проект без нервного срыва"
    Постарался сделать больше коментов, но могут быть не точности обращаться в телеграм см вверху
8 У системы есть документация со схемами, которая наглядно отвечает на вопрос: "Как это все работает" - 10 баллов

9. Выражение должно иметь возможность выполняться разными агентами - 10 баллов
  Не сделано, выражение пересылается целиком на агент и там вычисляется  