# Dataset layout

The coursework defines conversions and DQL workloads for four external graph
datasets:

- [MOOC User Action Dataset](https://snap.stanford.edu/data/act-mooc.html):
  [configuration](act-mooc/convert.yml), [schema](act-mooc/schema.dql),
  [queries](act-mooc/queries);
- [Elliptic++ Transactions Dataset](https://github.com/git-disl/EllipticPlusPlus/tree/main/Transactions%20Dataset):
  [configuration](elliptic++/convert.yml), [schema](elliptic++/schema.dql),
  [queries](elliptic++/queries);
- [California road network](https://snap.stanford.edu/data/roadNet-CA.html):
  [configuration](roadNet-CA/convert.yml), [schema](roadNet-CA/schema.dql),
  [queries](roadNet-CA/queries);
- [Stablecoin ERC20 Transactions Dataset](https://snap.stanford.edu/data/ERC20-stablecoins.html):
  [configuration](ERC20-stablecoins/convert.yml),
  [schema](ERC20-stablecoins/schema.dql),
  [queries](ERC20-stablecoins/queries).

Each dataset directory uses this local layout:

```text
dataset/
  convert.yml   tracked conversion rules
  schema.dql    tracked Dgraph schema
  queries/      tracked DQL workloads
  sources/      ignored upstream CSV/TSV files
  output.rdf    ignored generated converter output
```

Download source files directly from the upstream provider and place them under
the corresponding `sources/` directory with the filenames declared by
`convert.yml`. The repository does not redistribute those files. Dataset
licensing and permitted reuse are controlled by the upstream providers; review
their current terms before use.

For an offline example that requires no external download, use the committed
[minimal fixture](../internal/converter/testdata/minimal).
