// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/timshannon/threenamesinahat/files"
)

// requirements for google fonts and vuejs
const defaultCSP = "default-src 'self';font-src fonts.gstatic.com;style-src 'self' fonts.googleapis.com; script-src 'self' 'unsafe-eval'"

type TemplateHandlerFunc func(*templateWriter, *http.Request)

func templateHandler(handler TemplateHandlerFunc, templates ...string) http.HandlerFunc {
	tmpl := loadTemplates(templates...)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", defaultCSP)

		handler(&templateWriter{
			ResponseWriter: w,
			template:       tmpl,
		}, r)
	}
}

// template writers are passed into the http handler call
// carrying the template with them:
type templateWriter struct {
	http.ResponseWriter
	template *template.Template
}

func (t *templateWriter) execute(tdata interface{}) {
	// have to execute into a separate buffer, otherwise the partially executed template will show up
	// with the error page template
	var b bytes.Buffer
	err := t.template.Execute(&b, tdata)

	if err != nil {
		// TODO: Handle error
		log.Printf("Error executing template: %s", err)
	} else {
		_, err = io.Copy(t, &b)
		if err != nil {
			log.Printf("Error Copying template data to template writer: %s", err)
		}
	}
}

func loadTemplates(templateFiles ...string) *template.Template {
	tmpl := ""

	partialsDir := "partials"

	partials, err := files.AssetDir(partialsDir)
	if err != nil {
		panic(fmt.Errorf("Error loading partials directory: %s", err))
	}

	for i := range partials {
		str, err := files.Asset(filepath.Join(partialsDir, partials[i]))
		if err != nil {
			panic(fmt.Errorf("Loading partial %s: %s", filepath.Join(partialsDir, partials[i]), err))
		}
		tmpl += string(str)
	}

	for i := range templateFiles {
		str, err := files.Asset(templateFiles[i])
		if err != nil {
			panic(fmt.Errorf("Loading template file %s: %s", templateFiles[i], err))
		}
		tmpl += string(str)
	}

	// change delims to work with Vuejs
	return template.Must(template.New("").Delims("[[", "]]").Parse(tmpl))
}

func emptyTemplate(w *templateWriter, r *http.Request) {
	w.execute(nil)
}
