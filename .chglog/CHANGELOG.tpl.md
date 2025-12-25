# {{ .Info.Title }}

{{ if .Versions -}}
{{ range .Versions }}
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}

{{ if .CommitGroups -}}
{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- [`{{ .Hash.Short }}`]({{ $.Info.RepositoryURL }}/commit/{{ .Hash.Long }}) {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}
{{ end -}}

{{ if .RevertCommits -}}
### ⏪ Reverts
{{ range .RevertCommits -}}
- [`{{ .Hash.Short }}`]({{ $.Info.RepositoryURL }}/commit/{{ .Hash.Long }}) {{ .Revert.Header }}
{{ end -}}
{{ end -}}

{{ if .NoteGroups -}}
{{ range .NoteGroups -}}
### 💥 {{ .Title }}
{{ range .Notes }}
- {{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}

{{ end -}}
{{ else }}
*(No releases yet)*  
{{ end -}}
