{{ $width := 70 -}}
{{/********** Global Vars **********/ -}}
{{ $hp := "/" -}}
{{ $route := .default.Route -}}
{{ $domain := .default.Domain -}}
{{ $tblwriter_data := list (list ".:: __________T___1___8___S__________ ::.") (list "") (list (print "gopher = " $domain ":70/1/" $route)) (list (print "gopher+tls = " $domain ":7043/1/" $route)) -}}
{{/********** Main **********/ -}}
{{ gmenu (gurl_for $hp) "Home" -}}
{{ ginfo (tablewriter (dict "data" $tblwriter_data "width" $width "text-alignment" "center" "box-separator" "#" "box-draw-separate-rows" false)) -}}
