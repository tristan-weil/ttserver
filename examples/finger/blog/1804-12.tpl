{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
:: BLOG 1804/12 ::

{{ tablewriter (dict "data" (list (list "1804/12/02: new job!")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") -}}
I am really excited!
I am starting a new job today and I hope I'll do fine.
:-)
{{ template "footer.tpl" . }}
