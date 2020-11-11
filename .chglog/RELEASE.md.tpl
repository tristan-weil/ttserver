{{ range .Versions }}
{{      range .CommitGroups -}}
### {{ .Title }}
{{          range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{          end }}
{{      end -}}

{{-     if .RevertCommits -}}
### Reverts
{{          range .RevertCommits -}}
- {{ .Revert.Header }}
{{          end }}
{{      end -}}

{{-     if .NoteGroups -}}
{{          range .NoteGroups -}}
### {{ .Title }}
{{              range .Notes }}
{{ .Body }}
{{              end }}
{{          end -}}
{{      end -}}
{{ end -}}

{{- if .Versions -}}
{{      range .Versions -}}
{{          if .Tag.Previous -}}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
{{          end -}}
{{      end -}}
{{ end -}}
