{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
:: BLOG 1804/12 ::

{{ tablewriter (dict "data" (list (list "1804/12/02: having fun during this cruise!")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") -}}
What shall we do with a drunken sailor,
What shall we do with a drunken sailor,
What shall we do with a drunken sailor,
Early in the morning?

Shave his belly with a rusty razor,
Shave his belly with a rusty razor,
Shave his belly with a rusty razor,
Early in the morning.
{{ template "footer.tpl" . }}
