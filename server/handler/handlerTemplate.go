package handler

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hako/durafmt"
	"github.com/inhies/go-bytesize"
	gp_table "github.com/jedib0t/go-pretty/table"
	gp_text "github.com/jedib0t/go-pretty/text"
	ttconn "github.com/tristan-weil/ttserver/server/connection"
	ttversion "github.com/tristan-weil/ttserver/version"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func ServeConnHandlerCommonGetTextTemplatesFuncMap(c *ttconn.Connection) (map[string]interface{}, error) {
	return template.FuncMap{
		"normalize": func(text string) string {
			// from https://stackoverflow.com/questions/24588295/go-removing-accents-from-strings
			t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
			s, _, err := transform.String(t, text)
			if err != nil {
				return text
			}

			return s
		},

		"floating": func(width int, filler string, left string, right string) string {
			nbfiller := width - utf8.RuneCountInString(left) - utf8.RuneCountInString(right)
			nbfiller /= utf8.RuneCountInString(filler)

			return left + strings.Repeat(filler, nbfiller) + right
		},

		"rpad": func(width int, filler string, text string) string {
			nbfiller := width - utf8.RuneCountInString(text)
			nbfiller /= utf8.RuneCountInString(filler)

			return text + strings.Repeat(filler, nbfiller)
		},

		"lpad": func(width int, filler string, text string) string {
			nbfiller := width - utf8.RuneCountInString(text)
			nbfiller /= utf8.RuneCountInString(filler)

			return strings.Repeat(filler, nbfiller) + text
		},

		"center": func(width int, text string) string {
			return gp_text.AlignCenter.Apply(text, width)
		},

		"rjust": func(width int, text string) string {
			return gp_text.AlignRight.Apply(text, width)
		},

		"ljust": func(width int, text string) string {
			return gp_text.AlignLeft.Apply(text, width)
		},

		"underline": func(ch string, text string) string {
			width := utf8.RuneCountInString(text)
			line := strings.Repeat(ch, width)

			return fmt.Sprintf("%s\r\n%s", text, line)
		},

		"build_version": func() string {
			return fmt.Sprintf("%s", ttversion.Version)
		},

		"build_date": func() string {
			return fmt.Sprintf("%s", ttversion.BuildDate)
		},

		"bitcoincoreServices": func(flags interface{}) []string {
			services := map[string]uint64{
				"NONE":            0,
				"NETWORK":         1 << 0,
				"GETUTXO":         1 << 1,
				"BLOOM":           1 << 2,
				"WITNESS":         1 << 3,
				"XTHIN":           1 << 4,
				"NETWORK_LIMITED": 1 << 10,
			}

			var result []string
			var n uint64

			switch value := flags.(type) {
			default:
				n = 0
			case string:
				n, _ = strconv.ParseUint(value, 10, 32)
			case uint64:
				n = value
			}

			for service, bit := range services {
				if n&bit != 0 {
					result = append(result, service)
				}
			}

			return result
		},

		"bytesize": func(input interface{}) string {
			var n float64

			switch value := input.(type) {
			default:
				n = 0
			case string:
				n, _ = strconv.ParseFloat(value, 64)
			case float64:
				n = value
			}

			return bytesize.New(n).Format("%.02f ", "", false)
		},

		"duration": func(sec interface{}) string {
			var n int64

			switch value := sec.(type) {
			default:
				n = 0
			case string:
				n, _ = strconv.ParseInt(value, 10, 64)
			case float64:
				n = int64(value)
			case int64:
				n = value
			}

			duration := durafmt.Parse(time.Duration(n) * time.Second)

			return duration.String()
		},

		"tablewriter": func(config map[string]interface{}) string {
			buf := new(bytes.Buffer)

			t := gp_table.NewWriter()
			t.SetOutputMirror(buf)

			if config != nil {
				//
				// convert data
				//
				var nbColumns = 0
				var allRows []gp_table.Row
				for _, row := range config["data"].([]interface{}) {
					var curRow gp_table.Row
					curRow = row.([]interface{})
					nbColumns = len(row.([]interface{}))
					allRows = append(allRows, curRow)
				}

				//
				// global style
				//
				boxStyle := gp_table.BoxStyle{
					BottomLeft:       "+",
					BottomRight:      "+",
					BottomSeparator:  "+",
					Left:             "|",
					LeftSeparator:    "+",
					MiddleHorizontal: "-",
					MiddleSeparator:  "+",
					MiddleVertical:   "|",
					PaddingLeft:      " ",
					PaddingRight:     " ",
					Right:            "|",
					RightSeparator:   "+",
					TopLeft:          "+",
					TopRight:         "+",
					TopSeparator:     "+",
					UnfinishedRow:    "~",
				}

				if sep, ok := config["box-separator"]; ok {
					boxStyle.BottomSeparator = sep.(string)
					boxStyle.LeftSeparator = sep.(string)
					boxStyle.MiddleSeparator = sep.(string)
					boxStyle.RightSeparator = sep.(string)
					boxStyle.TopSeparator = sep.(string)
					boxStyle.Left = sep.(string)
					boxStyle.Right = sep.(string)
					boxStyle.TopLeft = sep.(string)
					boxStyle.TopRight = sep.(string)
					boxStyle.BottomLeft = sep.(string)
					boxStyle.BottomRight = sep.(string)
					boxStyle.MiddleVertical = sep.(string)
					boxStyle.MiddleHorizontal = sep.(string)
				}
				if sep, ok := config["box-separator-bottom"]; ok {
					boxStyle.BottomSeparator = sep.(string)
				}
				if sep, ok := config["box-separator-left"]; ok {
					boxStyle.LeftSeparator = sep.(string)
				}
				if sep, ok := config["box-separator-middle"]; ok {
					boxStyle.MiddleSeparator = sep.(string)
				}
				if sep, ok := config["box-separator-right"]; ok {
					boxStyle.RightSeparator = sep.(string)
				}
				if sep, ok := config["box-separator-top"]; ok {
					boxStyle.TopSeparator = sep.(string)
				}
				if sep, ok := config["box-bottom-left"]; ok {
					boxStyle.BottomLeft = sep.(string)
				}
				if sep, ok := config["box-bottom-right"]; ok {
					boxStyle.BottomRight = sep.(string)
				}
				if sep, ok := config["box-top-left"]; ok {
					boxStyle.TopLeft = sep.(string)
				}
				if sep, ok := config["box-top-right"]; ok {
					boxStyle.TopRight = sep.(string)
				}
				if sep, ok := config["box-left"]; ok {
					boxStyle.Left = sep.(string)
				}
				if sep, ok := config["box-middle-vertical"]; ok {
					boxStyle.MiddleVertical = sep.(string)
				}
				if sep, ok := config["box-right"]; ok {
					boxStyle.Right = sep.(string)
				}
				if sep, ok := config["box-middle-horizontal"]; ok {
					boxStyle.MiddleHorizontal = sep.(string)
				}
				if sep, ok := config["box-padding-left"]; ok {
					boxStyle.PaddingLeft = sep.(string)
				}
				if sep, ok := config["box-padding-right"]; ok {
					boxStyle.PaddingRight = sep.(string)
				}
				if sep, ok := config["box-unfinished-row"]; ok {
					boxStyle.UnfinishedRow = sep.(string)
				}

				boxOptions := gp_table.Options{
					DrawBorder:      true,
					SeparateColumns: true,
					SeparateFooter:  true,
					SeparateHeader:  true,
					SeparateRows:    true,
				}

				if q, ok := config["box-draw"]; ok {
					boxOptions.DrawBorder = q.(bool)
					boxOptions.SeparateColumns = q.(bool)
					boxOptions.SeparateFooter = q.(bool)
					boxOptions.SeparateHeader = q.(bool)
					boxOptions.SeparateRows = q.(bool)
				}
				if q, ok := config["box-draw-border"]; ok {
					boxOptions.DrawBorder = q.(bool)
				}
				if q, ok := config["box-draw-separate-columns"]; ok {
					boxOptions.SeparateColumns = q.(bool)
				}
				if q, ok := config["box-draw-separate-footer"]; ok {
					boxOptions.SeparateFooter = q.(bool)
				}
				if q, ok := config["box-draw-separate-header"]; ok {
					boxOptions.SeparateHeader = q.(bool)
				}
				if q, ok := config["box-draw-separate-rows"]; ok {
					boxOptions.SeparateRows = q.(bool)
				}

				t.SetStyle(gp_table.Style{
					Name:    "customstyle",
					Box:     boxStyle,
					Options: boxOptions,
				})

				//
				// columns style
				//
				var columnsWithFixedSize = make(map[int]int)
				var nbColumnsWithFixedSize = 0
				var totalWidthOfColumnsWithFixedSize = 0
				if w, ok := config["columns-width"]; ok {
					var wStrs []string
					for _, wStr := range w.([]interface{}) {
						wStrs = append(wStrs, wStr.(string))
					}
					for _, colAndWidthItem := range wStrs {
						var colwidth = strings.Split(colAndWidthItem, ",")
						colnum, errn := strconv.Atoi(colwidth[0])
						width, errw := strconv.Atoi(colwidth[1])
						if errn == nil && errw == nil {
							columnsWithFixedSize[colnum] = width
							totalWidthOfColumnsWithFixedSize += width
							nbColumnsWithFixedSize++
						}
					}
				}

				defaultColumnWidth := 0
				widthModulo := 0
				if w, ok := config["width"]; ok {
					nbSeparators := nbColumns + 1
					widthOfPaddings := nbColumns * (utf8.RuneCountInString(boxStyle.PaddingLeft) + utf8.RuneCountInString(boxStyle.PaddingRight))

					nbUnfixedColumns := nbColumns - nbColumnsWithFixedSize
					widthLeft := w.(int) - totalWidthOfColumnsWithFixedSize - (nbSeparators + widthOfPaddings)
					if nbUnfixedColumns > 0 {
						widthModulo = widthLeft % nbUnfixedColumns
						defaultColumnWidth = widthLeft / nbUnfixedColumns
					}
				}

				var columnsConfigs = make([]gp_table.ColumnConfig, nbColumns)
				var widthModuleAlreadyApplied = false
				for i := 0; i < nbColumns; i++ {
					curColumnConfig := gp_table.ColumnConfig{
						Number:      i + 1,
						Align:       gp_text.AlignLeft,
						AlignFooter: gp_text.AlignLeft,
						AlignHeader: gp_text.AlignLeft,
					}

					if w, ok := columnsWithFixedSize[i+1]; ok {
						curColumnConfig.WidthMin = w
						curColumnConfig.WidthMax = w
					} else {
						if !widthModuleAlreadyApplied {
							curColumnConfig.WidthMin = defaultColumnWidth + widthModulo
							curColumnConfig.WidthMax = defaultColumnWidth + widthModulo
							widthModuleAlreadyApplied = true
						} else {
							curColumnConfig.WidthMin = defaultColumnWidth
							curColumnConfig.WidthMax = defaultColumnWidth
						}
					}

					if align, ok := config["text-alignment"]; ok {
						switch align.(string) {
						case "right":
							curColumnConfig.Align = gp_text.AlignRight
							curColumnConfig.AlignFooter = gp_text.AlignRight
							curColumnConfig.AlignHeader = gp_text.AlignRight
						case "left":
							curColumnConfig.Align = gp_text.AlignLeft
							curColumnConfig.AlignFooter = gp_text.AlignLeft
							curColumnConfig.AlignHeader = gp_text.AlignLeft
						case "center":
							curColumnConfig.Align = gp_text.AlignCenter
							curColumnConfig.AlignFooter = gp_text.AlignCenter
							curColumnConfig.AlignHeader = gp_text.AlignCenter
						case "justify":
							curColumnConfig.Align = gp_text.AlignJustify
							curColumnConfig.AlignFooter = gp_text.AlignJustify
							curColumnConfig.AlignHeader = gp_text.AlignJustify
						default:
							curColumnConfig.Align = gp_text.AlignDefault
							curColumnConfig.AlignFooter = gp_text.AlignDefault
							curColumnConfig.AlignHeader = gp_text.AlignDefault
						}
					}

					columnsConfigs[i] = curColumnConfig
				}

				t.SetColumnConfigs(columnsConfigs)

				//
				// header/footer
				//
				if data, ok := config["header-data"]; ok {
					t.AppendHeader(data.([]interface{}))
				}

				if data, ok := config["footer-data"]; ok {
					t.AppendFooter(data.([]interface{}))
				}

				//
				// data
				//
				if paging, ok := config["data-paging-size"]; ok {
					t.SetPageSize(paging.(int))
				}

				t.AppendRows(allRows)
			}
			t.Render()

			return buf.String()
		},
	}, nil
}
