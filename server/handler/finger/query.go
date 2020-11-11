package finger

import (
	"regexp"
	"strings"
)

type Query struct {
	Username string   // Username, can be blank
	Hostname []string // Hostname (zero or more), not used
}

/*
   From RFC 1288 the BNF is as follows. This is matchable via a regexp.

        {Q1}    ::= [{W}|{W}{S}{U}]{C}
        {Q2}    ::= [{W}{S}][{U}]{H}{C}
        {U}     ::= username
        {H}     ::= @hostname | @hostname{H}
        {W}     ::= /W
        {S}     ::= <SP> | <SP>{S}
        {C}     ::= <CRLF>
*/
var lineRegexp = regexp.MustCompile(`` +
	`\s*` + // [{S}]
	`(?P<U>[\w-\./]+)?` + // [{U}]
	`(?P<H>(@[\w-\.]+)+)*`, // {H}
)

func ParseQuery(line string) (*Query, error) {
	values := findNamedMatches(lineRegexp, line)
	if values == nil {
		return &Query{}, nil
	}

	var result Query

	if username, ok := values["U"]; ok {
		result.Username = username
	}

	if hostnames, ok := values["H"]; ok {
		result.Hostname = strings.Split(hostnames, "@")
	}

	if strings.HasSuffix(result.Username, ".tpl") {
		result.Username = strings.TrimSuffix(result.Username, ".tpl")
	}

	result.Username = strings.TrimRight(result.Username, "/")

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
