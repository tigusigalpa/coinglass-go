# Funding-rate exchange list schema fixture

- Source: https://docs.coinglass.com/reference/fr-exchange-list
- Retrieved: 2026-10-05
- Fixture kind: `doc_schema_example`
- Transformation: the documentation example contains `//` comments, which were removed to form valid JSON. This file is not captured provider wire traffic.

The fixture preserves the documented parent `symbol`, the order of both margin
lists, and the documented row fields. It does not establish provider requiredness
or nullability beyond the illustrated shape.
