{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
:: BLOG 1969/07 ::

{{ ginfo (tablewriter (dict "data" (list (list "1969/07/20: sorry the delay guys!")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
Well, it seems my Internet connection is pretty bad those days.

I am not sure if I'll be able to send the pics of my last trip.
{{ template "footer.tpl" . }}
