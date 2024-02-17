контакты для связи -  телеграмм @ilyaVinokurow

Оглавление 
### Команды для тестирования в терминале
### Структура Работы программы 
### Структура проекта (файлы и папки)
### Как запустить
### Что сделано и не сделано



# Команды для тестирования в терминале

Посылаем значение выражения (exp)
curl -v -d "exp=13" -X POST http://localhost:8000/

!!!!! Есть проблемма при отправке через консоль символа + он заменяется на пробел
в браузере такой проблеммы нет 
curl -v -d "exp=2/2*(2+2)" -X POST http://localhost:8000/


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


# Как запустить 

Запуск проводился в терминале в VScode в Windows 10
Данная программа запускается как монолит и не разделена на подпрограммы 
запускается 
go run .\cmd\main.go

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


# Что сделано и не сделано
0 readme с описанием программы и как запустить 
1 Гуи для взаимодействия с пользователем в виде html страницы
2 