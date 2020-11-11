{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
:: STATS OF THIS SERVER SINCE LAST RESTART ::
(don't forget to reload your browser)

{{ if .prometheus.Data -}}
{{      $cacheData := list -}}
{{      $cacheHeaderData := list -}}
{{      $connActive := -1 -}}
{{      $procData := list -}}
{{      $procHeaderData := list -}}
{{      $connData := list -}}
{{      $connHeaderData := list -}}
{{      $connActData := list -}}
{{      $connActHeaderData := list -}}
{{      $goDataMap := dict -}}
{{      range $el := .prometheus.Data -}}
{{/********** Go Stats **********/ -}}
{{          if eq $el.Name "go_goroutines" -}}
{{              $_ := set $goDataMap "go_goroutines" (index $el.Metrics 0).Value -}}
{{          end -}}
{{          if eq $el.Name "go_memstats_alloc_bytes" -}}
{{              $_ := set $goDataMap "go_memstats_alloc_bytes" (bytesize (index $el.Metrics 0).Value) -}}
{{          end -}}
{{          if eq $el.Name "process_resident_memory_bytes" -}}
{{              $_ := set $goDataMap "process_resident_memory_bytes" (bytesize (index $el.Metrics 0).Value) -}}
{{          end -}}
{{          if eq $el.Name "process_open_fds" -}}
{{              $_ := set $goDataMap "process_open_fds" (index $el.Metrics 0).Value -}}
{{          end -}}
{{          if eq $el.Name "process_start_time_seconds" -}}
{{              $_ := set $goDataMap "process_start_time_seconds" (duration (sub (now | unixEpoch) (float64 (index $el.Metrics 0).Value))) -}}
{{          end -}}
{{/********** Cache status **********/ -}}
{{          if eq $el.Name "ttserver_route_cache_status_total" -}}
{{              $header := list -}}
{{              $header = append $header "route" -}}
{{                  $curSpaceData := list -}}
{{                  $headerStatus := list -}}
{{                  $routesList := list -}}
{{                  range $m := $el.Metrics -}}
{{                          $headerStatus = append $headerStatus $m.Labels.status -}}
{{                          $routesList = append $routesList $m.Labels.route -}}
{{                  end -}}
{{                  $headerStatus = uniq $headerStatus -}}
{{                  $routesList = sortAlpha $routesList | uniq -}}
{{                  range $hs := $headerStatus -}}
{{                      $header = append $header $hs -}}
{{                  end -}}
{{                  range $route := $routesList -}}
{{                      $curData := list -}}
{{                      $curData = append $curData $route -}}
{{                      range $hs := $headerStatus -}}
{{                          $curValue := "0" -}}
{{                          range $recm := $el.Metrics -}}
{{                                  if eq $recm.Labels.route $route -}}
{{                                      if eq $recm.Labels.status $hs -}}
{{                                          $curValue = $recm.Value  -}}
{{                                      end -}}
{{                                  end -}}
{{                          end -}}
{{                          $curData = append $curData $curValue -}}
{{                      end -}}
{{                      $curSpaceData = append $curSpaceData $curData -}}
{{                  end -}}
{{              $cacheData = $curSpaceData -}}
{{              $cacheHeaderData = uniq $header -}}
{{          end -}}
{{/********** Connections **********/ -}}
{{          if eq $el.Name "ttserver_active_conn" -}}
{{              $header := list -}}
{{              $header = append $header "count" -}}
{{              $allData := list -}}
{{              range $m := $el.Metrics -}}
{{                      $curData := list -}}
{{                      $curData = append $curData $m.Value -}}
{{                      $allData = append $allData $curData -}}
{{              end -}}
{{              $connActData = $allData -}}
{{              $connActHeaderData = uniq $header -}}
{{          end -}}
{{/********** Process Duration **********/ -}}
{{          if eq $el.Name "ttserver_process_duration_microseconds" -}}
{{              $header := list -}}
{{              $header = append $header "route" -}}
{{              $header = append $header "code" -}}
{{              $header = append $header "count" -}}
{{              $header = append $header "sum" -}}
{{                  $curSpaceData := list -}}
{{                  $routesList := list -}}
{{                  $headerQuantiles := list -}}
{{                  range $m := $el.Metrics -}}
{{                          $routesList = append $routesList $m.Labels.route -}}
{{                  end -}}
{{                  $routesList = sortAlpha $routesList | uniq -}}
{{                  range $route := $routesList -}}
{{                      $codesList := list -}}
{{                      range $m := $el.Metrics -}}
{{                              if eq $m.Labels.route $route -}}
{{                                  $codesList = append $codesList $m.Labels.code -}}
{{                              end -}}
{{                      end -}}
{{                      $codesList = sortAlpha $codesList | uniq -}}
{{                      range $code := $codesList -}}
{{                          $curData := list -}}
{{                          $curData = append $curData $route -}}
{{                          $curData = append $curData $code -}}
{{                          $curCount := "NaN" -}}
{{                          $curSum := "NaN" -}}
{{                          $quantList := list -}}
{{                          range $recm := $el.Metrics -}}
{{                                  if eq $recm.Labels.route $route -}}
{{                                      if eq $recm.Labels.code $code -}}
{{                                          $curCount = $recm.Count -}}
{{                                          $curSum = div (int64 (float64 $recm.Sum)) 1000 -}}
{{                                          range $k, $v := $recm.Quantiles -}}
{{                                              $headerQuantiles = append $headerQuantiles $k -}}
{{                                              $quantList = append $quantList (div (int64 (float64 $v)) 1000) -}}
{{                                          end -}}
{{                                      end -}}
{{                                  end -}}
{{                          end -}}
{{                          $curData = append $curData $curCount -}}
{{                          $curData = append $curData $curSum -}}
{{                          range $q := $quantList -}}
{{                              $curData = append $curData $q -}}
{{                          end -}}
{{                          $curSpaceData = append $curSpaceData $curData -}}
{{                      end -}}
{{                  end -}}
{{                  $headerQuantiles = uniq $headerQuantiles -}}
{{                  range $q := $headerQuantiles -}}
{{                      $header = append $header $q -}}
{{                  end -}}
{{              $procData = $curSpaceData -}}
{{              $procHeaderData = uniq $header -}}
{{          end -}}
{{/********** Connection Duration **********/ -}}
{{          if eq $el.Name "ttserver_request_duration_microseconds" -}}
{{              $header := list -}}
{{              $header = append $header "code" -}}
{{              $header = append $header "count" -}}
{{              $header = append $header "sum" -}}
{{                  $curSpaceData := list -}}
{{                  $routesList := list -}}
{{                  $headerQuantiles := list -}}
{{                  range $m := $el.Metrics -}}
{{                          $curData := list -}}
{{                          $curData = append $curData $m.Labels.code -}}
{{                          $curData = append $curData $m.Count -}}
{{                          $curData = append $curData (div (int64 (float64 $m.Sum)) 1000) -}}
{{                          range $k, $v := $m.Quantiles -}}
{{                              $headerQuantiles = append $headerQuantiles $k -}}
{{                              $curData = append $curData (div (int64 (float64 $v)) 1000) -}}
{{                          end -}}
{{                          $curSpaceData = append $curSpaceData $curData -}}
{{                  end -}}
{{                  $headerQuantiles = uniq $headerQuantiles -}}
{{                  range $q := $headerQuantiles -}}
{{                      $header = append $header $q -}}
{{                  end -}}
{{              $connData = $curSpaceData -}}
{{              $connHeaderData = uniq $header -}}
{{          end -}}
{{      end -}}
{{/********** Display **********/ -}}
{{ tablewriter (dict "data" (list (list "Go Stats")) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") }}
Go Routines: {{ (get $goDataMap "go_goroutines") }}
Alloc: {{ (get $goDataMap "go_memstats_alloc_bytes") }}
RSS: {{ (get $goDataMap "process_resident_memory_bytes") }}
Open fds: {{ (get $goDataMap "process_open_fds") }}
Started: {{ (get $goDataMap "process_start_time_seconds") }}

{{ tablewriter (dict "data" (list (list "Cache status")) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") }}
{{ tablewriter (dict "data" $cacheData "header-data" $cacheHeaderData "width" $width "text-alignment" "left") }}

{{ tablewriter (dict "data" (list (list "Active Connections")) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") }}
{{ tablewriter (dict "data" $connActData "header-data" $connActHeaderData "width" $width "text-alignment" "left") }}

{{ tablewriter (dict "data" (list (list "Connections Duration (in ms)")) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") }}
{{ tablewriter (dict "data" $connData "header-data" $connHeaderData "width" $width "text-alignment" "left") }}

{{ tablewriter (dict "data" (list (list "Template Processing (in ms)")) "width" $width "text-alignment" "center" "box-draw-separate-rows" false "box-separator" "~" "box-left" ")" "box-right" ")") }}
{{ tablewriter (dict "data" $procData "header-data" $procHeaderData "width" $width "text-alignment" "left") }}
{{ end }}
{{ template "footer.tpl" . }}
