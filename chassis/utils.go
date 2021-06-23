package chassis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//BuildErrorFromErrMsg reads an http.Response and return it an error´. This is useful when calling methods
// internally via REST Client
func BuildErrorFromErrMsg(response *http.Response) error {
	errorMessage := ErrResp{}
	rspBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err = json.Unmarshal(rspBody, &errorMessage); err != nil {
		return err
	}

	return errors.New(errorMessage.Message)
}

func BuildInArgument(args []string) string {
	return  "('" + strings.Join(args, `','`) + `')`
}

func GetTotalResultsFromQuery(originalQuery string, total *uint, db *sqlx.DB) error {
	//this will guarantee a match when calling strings.Split
	originalQuery = strings.ReplaceAll(originalQuery, "FROM", "from")
	originalQuery = strings.ReplaceAll(originalQuery, "ORDER BY", "order by")
	q := "SELECT COUNT(*) AS total FROM " + strings.Split(strings.Split(originalQuery, "from")[1], "order by")[0]
	if err := db.Get(total, q); err != nil {
		return err
	}
	return nil
}

func BuildPaginationResponse(w http.ResponseWriter, r *http.Request, page, perPage, total uint) {

	brokenURI := strings.Split(r.RequestURI, "?")
	w.Header().Set("Results", strconv.Itoa(int(total)))

	links := make(map[string]string)
	totalPages := (total / perPage) + 1

	if len(brokenURI) == 1 {
		prefix := r.Host + r.RequestURI
		links["next"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", page+1, perPage)
		links["first"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", 1, perPage)
		links["last"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", totalPages, perPage)
		AddLinksToHeader(w, &links)
		return
	}

	paramsString := brokenURI[1]
	var newParams []string
	if strings.Contains(paramsString, "page") || strings.Contains(paramsString, "per_page") {
		params := strings.Split(paramsString, "&")
		for _, param := range params {
			p := strings.Split(param, "=")
			if p[0] == "page" || p[0] == "per_page" {
				continue
			}
			newParams = append(newParams, param)
		}
	}

	if len(newParams) == 0 {
		prefix := r.Host + brokenURI[0]
		if page > 1 {
			links["before"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", page-1, perPage)
		}
		if page < totalPages {
			links["next"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", page+1, perPage)
		}
		links["first"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", 1, perPage)
		links["last"] = prefix + fmt.Sprintf("?page=%d&per_page=%d", totalPages, perPage)

		AddLinksToHeader(w, &links)
		return
	}

	prefix := r.Host + brokenURI[0] + "?" + strings.Join(newParams, "&")
	if page > 1 {
		links["before"] = prefix + fmt.Sprintf("&page=%d&per_page=%d", page-1, perPage)
	}
	if page < totalPages {
		links["next"] = prefix + fmt.Sprintf("&page=%d&per_page=%d", page+1, perPage)
	}
	links["first"] = prefix + fmt.Sprintf("&page=%d&per_page=%d", 1, perPage)
	links["last"] = prefix + fmt.Sprintf("&page=%d&per_page=%d", totalPages, perPage)

	AddLinksToHeader(w, &links)
	return

}

func AddLinksToHeader(w http.ResponseWriter, links *map[string]string) {
	//<http://example.org>; rel="foo"; title="bar", <http://test.com>; rel="baz"
	var values []string
	for k, v := range *links {
		values = append(values, fmt.Sprintf(`<https://%s>"; rel="%s"`, v , k))
	}

	w.Header().Set("Link", strings.Join(values, ","))
}

func Wait(value int, unit time.Duration) {
	period := time.Tick(time.Duration(value) * unit)
	for range period {
		break
	}
}

func FormatCurrencyValue(currency string, value int) string {
	parsedTwoDecimalDigits := fmt.Sprintf("%.2f", float64(value)/100)
	switch strings.ToUpper(currency) {
	case "EUR":
		return "€" + " " + parsedTwoDecimalDigits
	case "GBP":
		return "£" + " " + parsedTwoDecimalDigits
	case "USD":
		return "$" + " " + parsedTwoDecimalDigits
	default:
		return currency + " " + parsedTwoDecimalDigits
	}
}