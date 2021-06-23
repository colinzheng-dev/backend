package db

import (
	"strings"
)

func paramsWhere(ps *DatabaseParams) string {
	es := []string{}

	if ps.PostTypes != nil && len(ps.PostTypes) > 0 {
		es = append(es, `post_type IN ('`+strings.Join(ps.PostTypes, "', '")+`')`)
	}

	if ps.Subject != nil {
		es = append(es, `subject IN ('`+*ps.Subject+`')`)
	}

	if len(es) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(es, ` AND `)
}

func paramsOrderBy(ps *DatabaseParams) string {
	if ps.SortBy == nil {
		return ` ORDER BY created_at ASC `
	}

	switch ps.SortBy.Field {
	case "rank", "user_bought":
		return " ORDER BY attrs -> '" + ps.SortBy.Field + "' " + ps.SortBy.Order + " "
	default:
		return ` ORDER BY created_at ASC `
	}

}
