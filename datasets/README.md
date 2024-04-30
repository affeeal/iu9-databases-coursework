# Датасеты

Оценка производительности Dgraph производится при выполнении запросов над
следующими датасетами:

- [MOOC User Action Dataset](https://snap.stanford.edu/data/act-mooc.html);
- [Elliptic++ Transactions Dataset](https://github.com/git-disl/EllipticPlusPlus/tree/main/Transactions%20Dataset);
- TODO;
- TODO.

Под каждый датасет выделен отдельный каталог со следующей структурой:

- `source` - исходные файлы датасета;
- `transformed` - преобразованные исходные файлы в формате, пригодном для
  bulk import;
- `schema` - DQL-схема датасета;
- `queries` - DQL-запросы.

Исходные файлы добавлены в `.gitignore`; скачать их можно по ссылкам выше.
