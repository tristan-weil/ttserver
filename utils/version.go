package utils

import (
	"bytes"
	"fmt"
	"runtime"
	"text/template"

	ttversion "github.com/tristan-weil/ttserver/version"
)

var versionTemplate = `Version:      {{.Version}}
Go version:   {{.GoVersion}}
Built:        {{.BuildTime}}
OS/Arch:      {{.Os}}/{{.Arch}}`

func PrintVersion() {
	buf := new(bytes.Buffer)

	tmpl, err := template.New("").Parse(versionTemplate)
	if err != nil {
		fmt.Printf("unable to get version: %s\n", err)
	}

	v := struct {
		Version   string
		Codename  string
		GoVersion string
		BuildTime string
		Os        string
		Arch      string
	}{
		Version:   ttversion.Version,
		GoVersion: runtime.Version(),
		BuildTime: ttversion.BuildDate,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	err = tmpl.Execute(buf, v)
	if err != nil {
		fmt.Printf("unable to get version: %s\n", err)
	}

	fmt.Println(buf.String())
}
