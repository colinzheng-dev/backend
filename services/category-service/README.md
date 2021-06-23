# Category service


## Routes

```
GET /categories
```

Get list of categories and their JSON schemas.

```
GET /category/{category}
```

Get the entries of a single category as a JSON object keyed by
category item label. Each entry in the object corresponds to the JSON
schema for the category.

```
PUT /category/{category}/{name}
```

Add an entry to a category under a given label. The request body
should contain a JSON value matching the JSON schema of the category.
