package utils

import (
	"bytes"
	"text/template"
)

type DomainParams struct {
	Name     string
	Memory   int
	VCPU     int
	DiskPath string
}

func LoadDomainXML(path string, params DomainParams) (string, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", err
	}
	return buf.String(), nil
}
