# Оценка ресурсов, потраченных на выполнение запроса

Вспомогательная программа `benchmark.go` выполняет read-only DQL-запрос и (наивно)
отслеживает память и время, затраченные на выполнение. Программа замеряет свободную
оперативную память до выполнения запроса и отслеживает её минимальное значение во
время выполнения; разница двух значений выводится в байтах. Затраченное на выполнение
запроса время в наносекундах берётся из debug-информации.

Допустимые опции программы:

- `query-path` - путь к файлу с DQL-запросом;
- `host` - хост сервера Dgraph (по умолчанию, `localhost`);
- `port` - порт сервера Dgraph (по умолчанию, `9080`);
- `print-respond` - печатать ли JSON-результат выполнения запроса (по умолчанию, `false`);
- `duration` - частота замера свободной оперативной памяти во время выполнения (по умолчанию, `100ms`).