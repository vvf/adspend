АПИ (JSON):
POST /record - добавление записи, тело - JSON
Результат:
Успешный (HTTP 200): {"Success":true,"Message":""}
Ошибка (HTTP 406): {"Success":false,"Message":"подробности ошибки"}  - неверный JSON
Ошибка (HTTP 500): {"Success":false,"Message":"подробности ошибки"}  - ошибка при записи в БД.

Получение:

GET
/record/<field>/<filterValue>/Count - получение количества строк.
<field> = («Timestamp»|«Action»|«SSP»)  — Название поля по которому фильтрация
<filterValue> - значение поля по которому идет фильтрация, для поля TimeStamp - это может быть диапазон значений ([0-9]*-[0-9]*), но может быть и конкретное значение.

Возвращаемый JSON в формате:
"aggregate":"Count" - аггрегирующая функция
"data": <aggregateResult> - количество строк
"field":<field>, - поле по которому была фильтрация
"filterValue":<filterValue>, - значения поля при фильтрации
"ts":1478616062 - время когда начал выполняться запрос

/record/<field>/<filterValue> - получение самих строк.

Возвращаемый JSON в формате:
"сount»:<integer> - количество обхектов в массиве data
"data»:<Array> - массив строк
"isPartial":true - если выданы не все данные (ограничение максимального количества строк можно будет вынести в настройки как и многое другое)
        или отсутствует если выданы все существующие данные по фильтру.
"field":<field>, - поле по которому была фильтрация
"filterValue":<filterValue>, - значения поля при фильтрации
"ts":1478616062 - время когда начал выполняться запрос

Итого сейчас набор выглядит так:

/record/Timestamp/<filterValue>
/record/Timestamp/<filterValue>/Count
/record/Action/<filterValue>
/record/Action/<filterValue>/Count
/record/SSP/<filterValue>
/record/SSP/<filterValue>/Count
