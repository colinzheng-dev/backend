# Search service

## Postgres full-text search

 - Column type of `TSVECTOR` for full text search items.
 - Normalise text and convert to `TSVECTOR` with
   `to_tsvector('english', 'text data')`.
 - Can assign different weights to different lexemes in `TSVECTOR`:
   treat name, tags, etc. differently to description and body text?
   There are four levels of weighting, `A` to `D`.
 - Queries represented by `TSQUERY` type: boolean operators, "followed
   by" operator, weight matching, prefix matching.
 - Normalise search terms and convert to `TSQUERY` with `to_tsquery`
   functions.
 - Matching uses stemming, so results may be confusing for prefix
   searches.
 - `websearch_to_tsquery` function is probably the thing to use for
   processing search terms: handles logical AND of multiple words,
   "followed by" for quoted phrases, stop word removal, logical OR
   ("cat OR dog") and negation ("dog -poodle").
 - Relevance sorting of results.


## Search service functionality

 - Endpoints:
    * GET /
    * GET /healthz
    * GET /search/geo?location=lat,lon&dist=d
    * GET /search/full_text?q=query
    * GET /search/region?region=query
 - Index maintenance:
    * Listen for item service pub/sub messages: create item, update
      item, delete item and update index tables.
    * Sync with item service at startup (take a simple approach for
      now, assuming we won't have more than ~10,000 items):
       - Get all item IDs from item service.
       - For each item ID, get search information from item service:
         need new item service route to return searchable text and
         location.
       - Upsert entries in location and full text index tables.



## Examples

```
UPDATE pgweb SET textsearchable_index_col =
  to_tsvector('english', coalesce(title,'') || ' ' || coalesce(body,''));
```

```
UPDATE tt SET ti =
  setweight(to_tsvector(coalesce(title,'')), 'A')    ||
  setweight(to_tsvector(coalesce(keyword,'')), 'B')  ||
  setweight(to_tsvector(coalesce(abstract,'')), 'C') ||
  setweight(to_tsvector(coalesce(body,'')), 'D');
```

```
SELECT title
  FROM pgweb
 WHERE textsearchable_index_col @@ to_tsquery('create & table')
 ORDER BY last_mod_date DESC
 LIMIT 10;
```

```
SELECT title, ts_rank_cd(textsearch, query) AS rank
  FROM apod, to_tsquery('neutrino|(dark & matter)') query
 WHERE query @@ textsearch
 ORDER BY rank DESC
 LIMIT 10;
```

```
SELECT item_id
  FROM item_full_text, websearch_to_tsquery($1) query
 WHERE query @@ full_text
 ORDER BY ts_rank_cd(full_text, query) DESC
 LIMIT 10;
```
