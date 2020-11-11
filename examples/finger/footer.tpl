{{ $width := 70 -}}
{{/********** Main **********/ -}}
{{ "" }}
{{ lpad $width "_" "_" }}
{{ rjust $width (print "powered by ttserver-" build_version) -}}
