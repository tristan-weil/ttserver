package gopher

import (
	"regexp"
	"strings"
)

type Query struct {
	Selector string
	ExtData  string
}

var lineRegexp = regexp.MustCompile(`` +
	`/*` +
	`(?:` +
	`(?P<ExtSelector>(?:URL))\:(?P<ExtData>.*)` +
	`|` +
	`(?:(?P<Selector>[\w/\.-_]+)/*)` +
	`)`,
)

func ParseQuery(line string) (*Query, error) {
	values := findNamedMatches(lineRegexp, line)
	if values == nil {
		return &Query{}, nil
	}

	var result Query

	if sel, ok := values["Selector"]; ok && sel != "" {
		result.Selector = sel
	} else {
		if gextSel, ok := values["ExtSelector"]; ok && gextSel != "" {
			result.Selector = gextSel
		}
		if gextSelData, ok := values["ExtData"]; ok && gextSelData != "" {
			result.ExtData = gextSelData
		}
	}

	if strings.HasSuffix(result.Selector, ".tpl") {
		result.Selector = strings.TrimSuffix(result.Selector, ".tpl")
	}

	result.Selector = strings.TrimRight(result.Selector, "/")

	return &result, nil
}

func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)
	if match == nil {
		return nil
	}

	namedGroups := regex.SubexpNames()

	results := map[string]string{}
	for i, name := range match {
		results[namedGroups[i]] = name
	}

	return results
}
