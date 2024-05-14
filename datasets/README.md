# Датасеты

Оценка производительности Dgraph производится при выполнении запросов над
следующими датасетами:

- [MOOC User Action Dataset](https://snap.stanford.edu/data/act-mooc.html);
- [Elliptic++ Transactions Dataset](https://github.com/git-disl/EllipticPlusPlus/tree/main/Transactions%20Dataset);
- [California road network](https://snap.stanford.edu/data/roadNet-CA.html)
- [Stablecoin ERC20 Transactions Dataset](https://snap.stanford.edu/data/ERC20-stablecoins.html).

Под каждый датасет выделен отдельный каталог со следующей структурой:

- `schema.dql` - DQL-схема датасета;
- `src` - исходные файлы датасета;
- `queries` - DQL-запросы;
- `configs` - конфигурационные файлы (в частности, `convert.yml`, в котором
  описываются правила преобразования исходных `csv`/`tsv`-файлов в
  Dgraph-расширение формата RDF).

Исходные файлы добавлены в `.gitignore`; скачать их можно по ссылкам выше.
