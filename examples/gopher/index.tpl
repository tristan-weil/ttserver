{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
Welcome to this `Gopher' server!
{{ ginfo (tablewriter (dict "data" (list (list "ABOUT")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
{{ gmenu (gurl_for "about/server") (floating $width " " ":: about/server" "about this server") -}}
{{ gmenu (gurl_for "about/stats") (floating $width " " ":: about/stats" "stats of the server") -}}
{{ gurl "https://t18s.fr" (floating $width " " ":: https://t18s.fr " "more information available") }}
{{ ginfo (tablewriter (dict "data" (list (list "BLOG" )) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
{{ gmenu (gurl_for "blog/1969/07") (floating $width " " ":: blog/1969/07" "entries for 1969/07") -}}
{{ gmenu (gurl_for "blog/1804/12") (floating $width " " ":: blog/1804/12" "entries for 1804/12") -}}
{{ gmenu (gurl_for "blog/1492/10") (floating $width " " ":: blog/1492/10" "entries for 1492/10") }}
{{ ginfo (tablewriter (dict "data" (list (list "AGGREGATORS" )) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
{{ gmenu (gurl_for "aggregator/unix") (floating $width " " ":: aggregator/unix" "unix news from the world") -}}
{{ gmenu (gurl_for "aggregator/music") (floating $width " " ":: aggregator/music" "good music feeds") }}
{{ template "footer.tpl" . }}
