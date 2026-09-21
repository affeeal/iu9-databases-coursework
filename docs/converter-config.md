# Converter configuration reference

Each dataset directory contains `convert.yml`, a `sources/` directory, and
optionally `schema.dql` and DQL queries. Conversion writes `output.rdf` in the
dataset directory only after every input succeeds.

## Minimal complete example

```yaml
files:
  - name: people.csv
    delimiter: ","
    comment: "#"
    declarations:
      - name: person
        type: id
        prefix: p
      - name: name
        type: string
      - name: score
        type: int
    artificial_declaration:
      name: row
      type: id
      prefix: r
    rdfs:
      - subject: person
        predicate: name
        object: name
      - subject: person
        predicate: score
        object: score
```

## File fields

- `files`: ordered list of source-file conversions;
- `name`: filename below the dataset's `sources/` directory;
- `headers`: optional column names for a headerless file; when omitted, the
  first non-comment record supplies the headers;
- `delimiter`: optional single Unicode rune; comma is the default;
- `comment`: optional single Unicode rune marking comment lines;
- `declarations`: source columns and their converter types;
- `artificial_declaration`: optional generated zero-based row ID appended only
  when this mapping is present;
- `entity_facets`: facets recorded for an entity and reusable by later RDF
  rules;
- `rdfs`: triples emitted for each nonempty object value.

## Declarations and types

Each declaration has a unique `name`, a `type`, and an optional `prefix` used to
form stable blank-node labels for IDs.

- `id`: a blank-node identifier; required for RDF subjects;
- `int`: integer literal;
- `float`: floating-point literal;
- `string`: string literal with RDF escaping.

Input headers must be unique and must contain every ordinary declaration.
The same checks apply to explicit `headers`; every record must have exactly that
many fields. For example, the original roadNet file has a commented-out header:

```yaml
name: roadNet-CA.txt
headers: [FromNodeId, ToNodeId]
delimiter: "\t"
comment: "#"
```

Artificial declaration names must not collide with an input header. Unknown
types, columns, casts, and rule references are rejected with the source filename
and field name where applicable.

## RDF rules

- `subject`: declared `id` column used as the RDF subject;
- `predicate`: Dgraph predicate placed in angle brackets;
- `object`: declared source or artificial column;
- `cast_object_to`: optional converter type used only for this object;
- `facets`: ordered inline facet rules;
- `entity_facets_id`: optional entity ID whose previously recorded facets are
  appended in sorted-key order.

An ordinary facet contains `key` and `value`. The value must refer to a non-ID
declaration. An entity facet additionally contains `id`, which must refer to an
ID declaration:

```yaml
entity_facets:
  - id: transaction
    key: timestamp
    value: timestamp

rdfs:
  - subject: sender
    predicate: transfers
    object: receiver
    entity_facets_id: transaction
    facets:
      - key: amount
        value: amount
```

String values escape quotes, backslashes, newlines, tabs, and control
characters. Facet maps are rendered deterministically. Empty RDF objects are
skipped. Integer and float lexical values retain the historical quoted form and
are loaded together with the corresponding Dgraph schema.

Types select RDF formatting; they do not parse or numerically convert input
values. IDs, predicates, and facet keys are emitted verbatim and must already
be valid Dgraph identifiers. Blank-node labels concatenate `prefix` and the
input value: choose prefixes that cannot collide across entity types. Entity
facets accumulate in file/record order; a later value replaces an earlier value
for the same entity and key. Reference them only after their defining records
have been processed, and avoid duplicate keys between inline and entity facets.

## Validation and failure behavior

The configuration must be one YAML document with at least one source file.
The YAML decoder rejects unknown fields and additional documents. Source paths
must be relative and remain lexically within `sources/`; configuration and
source symlinks are trusted local input, not a filesystem security boundary.
Declarations are validated before
input is read; headers are then checked for duplicates and missing columns.
Subjects and entity IDs must have type `id` and nonempty values, while facet
values must not have type `id`. A
conversion is written to a temporary file in the dataset directory, flushed,
closed, and atomically renamed to `output.rdf`. An error leaves any prior
complete output unchanged and removes the temporary file.
