package server

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// link-format ::= ids | summary | full

// link-query ::= link-type ( , link-type )+

// link-type ::= link-type-name [ : link-format ]

// link-type-name ::= <name from item_link_types table>

// LinkFormat represents the possible formats for embedding link
// information.
type LinkFormat uint

// Link format options.
const (
	IDsFormat LinkFormat = iota
	SummaryFormat
	FullFormat
)

// LinkInfo maps from the names of link types to be included to the
// formats to use for them.
type LinkInfo map[string]LinkFormat

// ParseLinkParams parses the link_format and links query parameters.
func ParseLinkParams(qs url.Values) (LinkInfo, error) {
	defaultLinkFormat, err := parseLinkFormat(qs.Get("link_format"), true)
	if err != nil {
		return nil, err
	}

	linkParam := qs.Get("links")
	if linkParam == "" {
		return nil, nil
	}
	links := strings.Split(linkParam, ",")
	info := LinkInfo{}
	for _, link := range links {
		format := defaultLinkFormat

		parts := strings.Split(link, ":")
		if len(parts) == 2 {
			format, err = parseLinkFormat(parts[1], false)
			if err != nil {
				return nil, err
			}
			link = parts[0]
		}

		info[link] = format
	}

	return info, nil
}

func parseLinkFormat(format string, allowBlank bool) (LinkFormat, error) {
	switch format {
	case "":
		if allowBlank {
			return IDsFormat, nil
		}
	case "ids":
		return IDsFormat, nil
	case "summary":
		return SummaryFormat, nil
	case "full":
		return FullFormat, nil
	}

	return IDsFormat, errors.New("invalid link_format parameter")
}

func (format LinkFormat) String() string {
	switch format {
	case IDsFormat:
		return "ids"
	case SummaryFormat:
		return "summary"
	case FullFormat:
		return "full"
	}
	return "unknown"
}

func (links LinkInfo) String() string {
	ss := []string{}
	for name, format := range links {
		ss = append(ss, name+":"+format.String())
	}
	return strings.Join(ss, ",")
}
