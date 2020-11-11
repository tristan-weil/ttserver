package gopher

import (
	"regexp"

	tthandler "github.com/tristan-weil/ttserver/server/handler"
)

type (
	gopherItemType byte

	gopherItem struct {
		Type        gopherItemType
		ExtraType   string
		Description string
		Selector    string
		Host        string
		Port        string
	}
)

const (
	TEXT      = gopherItemType('0') // Plain text file
	MENU      = gopherItemType('1') // Gopher submenu
	PHONEBOOK = gopherItemType('2') // CCSO flat database; other databases
	ERROR     = gopherItemType('3') // Error message
	BINHEX    = gopherItemType('4') // Macintosh BinHex file
	ARCHIVE   = gopherItemType('5') // Archive file (zip, tar, gzip, etc)
	UUENCODED = gopherItemType('6') // UUEncoded file
	INDEX     = gopherItemType('7') // Search query
	TELNET    = gopherItemType('8') // Telnet session
	BINARY    = gopherItemType('9') // Binary file
	GIF       = gopherItemType('g') // GIF format graphics file
	IMAGE     = gopherItemType('I') // Image file
	DOC       = gopherItemType('d') // Word processing document (ps, pdf, doc, etc)
	AUDIO     = gopherItemType('s') // Sound file
	VIDEO     = gopherItemType(';') // Video file
	HTML      = gopherItemType('h') // HTML document
	INFO      = gopherItemType('i') // Info line
	REDUNDANT = gopherItemType('+') // Redundant server
	TN3270    = gopherItemType('T') // Telnet to: tn3270 series server
	MIME      = gopherItemType('M') // MIME file (mbox, emails, etc)
	CALENDAR  = gopherItemType('c') // Calendar file
)

var TYPES_REGEXP = regexp.MustCompile(`^([0123456789gIds;hi+TMc])(.*?\s+)(((URL|TITLE):.*?)\s+)?([\w\d_\-.]+\s+\d+)$`)

func (g *gopherItem) Bytes() []byte {
	b := []byte{}

	b = append(b, byte(g.Type))
	b = append(b, []byte(g.Description)...)
	b = append(b, tthandler.TAB)

	if g.Type == HTML && g.ExtraType == "URL" {
		b = append(b, []byte("URL:")...)
	}

	if g.Type == INFO && g.ExtraType == "TITLE" {
		b = append(b, []byte("TITLE")...)
	}

	if g.Selector != "" {
		b = append(b, []byte(g.Selector)...)
	}

	b = append(b, tthandler.TAB)

	b = append(b, []byte(g.Host)...)
	b = append(b, tthandler.TAB)
	b = append(b, []byte(g.Port)...)
	b = append(b, []byte(tthandler.CRLF)...)

	return b
}

func (g *gopherItem) String() string {
	return string(g.Bytes())
}
