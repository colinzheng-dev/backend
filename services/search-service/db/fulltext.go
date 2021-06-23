package db

import (
	"errors"
	"github.com/rs/zerolog/log"
	itemModel "github.com/veganbase/backend/services/item-service/model"
	itemTypes "github.com/veganbase/backend/services/item-service/model/types"
	"strings"
)


// FullText performs a full text search for a given query string,
// returning a list of item IDs in order of relevance.
func (pg *PGClient) FullText(query string,
	itemType *itemModel.ItemType,
	approval *[]itemTypes.ApprovalState) ([]string, error) {
	ids := []string{}
	err := pg.DB.Select(&ids,
		fullText+typeApprovalWhere(itemType, approval) + fullTextOrder, query)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

const fullText = `
SELECT item_id
  FROM item_full_text, websearch_to_tsquery($1) query
 WHERE query @@ full_text`
const fullTextOrder = ` ORDER BY ts_rank_cd(full_text, query) DESC`

func typeApprovalWhere(itemType *itemModel.ItemType,
	approval *[]itemTypes.ApprovalState) string {
	wheres := []string{}
	if itemType != nil {
		wheres = append(wheres, "item_type = '"+itemType.String()+"'")
	}
	if approval != nil {
		if len(*approval) == 1 {
			wheres = append(wheres, "approval = '"+(*approval)[0].String()+"'")
		} else {
			as := []string{}
			for _, a := range *approval {
				as = append(as, "'"+a.String()+"'")
			}
			wheres = append(wheres, "approval IN ("+strings.Join(as, ",")+")")
		}
	}
	if len(wheres) == 0 {
		return ""
	}
	return " AND " + strings.Join(wheres, " AND ")
}


// AddFullText adds full-text information to the search index for an
// item.
func (pg *PGClient) AddFullText(id string,
	itemType itemModel.ItemType,
	approval itemTypes.ApprovalState,
	name, description, content string, tags []string) error {
	result, err := pg.DB.Exec(addFullText, id, itemType, approval,
		name, description, strings.Join(tags, " "), content)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("couldn't update full text index entry for '" + id + "'")
	}
	return nil
}

const addFullText = `
INSERT INTO item_full_text(item_id, item_type, approval, full_text)
 VALUES ($1, $2, $3,
         setweight(to_tsvector($4), 'A') ||
         setweight(to_tsvector($5), 'B') ||
         setweight(to_tsvector($6), 'C') ||
         setweight(to_tsvector($7), 'D'))
 ON CONFLICT (item_id)
 DO UPDATE SET item_type = $2, approval = $3,
   full_text = setweight(to_tsvector($4), 'A') ||
               setweight(to_tsvector($5), 'B') ||
               setweight(to_tsvector($6), 'C') ||
               setweight(to_tsvector($7), 'D')`

// ItemRemoved deletes search index information for an item.
func (pg *PGClient) ItemRemoved(id string) {
	_, err := pg.DB.Exec(deleteFullText, id)
	if err != nil {
		log.Error().Err(err).
			Msgf("removing full text information for item ID '%s", id)
	}
	_, err = pg.DB.Exec(deleteGeo, id)
	if err != nil {
		log.Error().Err(err).
			Msgf("removing geo information for item ID '%s", id)
	}
}

const deleteFullText = `DELETE FROM item_full_text WHERE item_id = $1`
const deleteGeo = `DELETE FROM item_locations WHERE item_id = $1`
