// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/timshannon/threenamesinahat/files"
)

// serveStatic serves a static file or directory.
// assumes one param for directories
func serveStatic(fileOrDir string, compress bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modTime := time.Time{}
		if r.Method != "GET" {
			notFound(w, r)
			return
		}

		file := r.URL.Path

		var reader *bytes.Reader
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && compress &&
			w.Header().Get("Content-Encoding") != "gzip" {
			data, err := files.AssetCompressed(file)
			if err != nil {
				notFound(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "gzip")
			reader = bytes.NewReader(data)
		} else {
			data, err := files.Asset(file)
			if err != nil {
				notFound(w, r)
				return
			}
			reader = bytes.NewReader(data)
		}

		http.ServeContent(w, r, file, modTime, reader)
	}
}
