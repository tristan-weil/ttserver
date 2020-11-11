{{ $width := 70 -}}
{{/********** Global Vars **********/ -}}
{{ $route := .default.Route -}}
{{ $domain := .default.Domain -}}
{{ $tblwriter_data := list (list ".:: __________T___1___8___S__________ ::.") (list "") (list (print "finger = " $route "@" $domain ":79")) (list (print "finger+tls = " $route "@" $domain ":7943")) -}}
{{/********** Main **********/ -}}
{{ tablewriter (dict "data" $tblwriter_data "width" $width "text-alignment" "center" "box-separator" "#" "box-draw-separate-rows" false) -}}
