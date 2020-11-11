{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{ $route := .default.Route -}}
{{/********** Main **********/ -}}
{{ cat ":: AGGREGATOR" (upper (replace "aggregator/" "" $route)) "::" }}
{{ $main := . -}}
{{ range $fetchKey := keys . | sortAlpha -}}
{{      if ne $fetchKey "default" -}}
{{          $fetch := get $main $fetchKey -}}
{{          with $fetch.Data -}}
{{              print " " }}
{{              $title := list (normalize .Title) -}}
{{              $link := list .Link -}}
{{              $length := len .Items -}}
{{              if eq $length 0 -}}
{{                  tablewriter (dict "data" (list $title $link) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") -}}
{{              else -}}
{{                  $last := list (date "(2006-01-02 15:04:05)" ((index .Items 0).PublishedParsed)) -}}
{{                  tablewriter (dict "data" (list $title $link $last) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") -}}
{{                  $counter := 0 -}}
{{                  range .Items -}}
{{                      if lt $counter 15 -}}
{{                          $ititle := .Title -}}
{{                          $ipublishedParsed := date "2006-01-02 15:04:05" .PublishedParsed -}}
{{                          print ":: " (abbrev (int (sub $width 3)) (normalize $ititle)) }}
{{                          $counter = add $counter 1 -}}
{{                      end -}}
{{                  end -}}
{{              end -}}
{{          end -}}
{{      end -}}
{{ end }}
{{ template "footer.tpl" . }}
