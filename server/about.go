// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/url"
)

const issueURLBase = "https://github.com/timshannon/threenamesinahat/issues/new"

type attribute struct {
	Name        string
	URL         string
	Author      string
	LicenseType string
}

func aboutTemplate(w *templateWriter, r *http.Request) {
	w.execute(struct {
		IssueLink   string
		Attribution []attribute
	}{
		IssueLink:   issueLink(r),
		Attribution: attribution,
	})
}

func issueLink(r *http.Request) string {
	v := url.Values{}
	v.Add("subject", "")
	v.Add("body", "<details><summary>User Agent</summary>"+r.UserAgent()+"</details>")
	return issueURLBase + "?" + v.Encode()
}

var attribution = []attribute{
	{
		Name:        "Go",
		URL:         "http://golang.org",
		Author:      "The Go Authors",
		LicenseType: "BSD",
	}, {
		Name:        "Vue.js",
		URL:         "https://vuejs.org/",
		Author:      "Yuxi (Evan) You",
		LicenseType: "MIT",
	},
	{
		Name:        "vfsgen",
		URL:         "https://github.com/shurcooL/vfsgen",
		Author:      "Dmitri Shuralyov",
		LicenseType: "MIT",
	},
	{
		Name:        "PaperCSS",
		URL:         "https://www.getpapercss.com",
		Author:      "Rhyne Vlaservich",
		LicenseType: "ISC",
	},
	{
		Name:        "Gorrilla Websocket",
		URL:         "https://github.com/gorilla/websocket",
		Author:      "Gary Burd and Joachim Bauch",
		LicenseType: "BSD 2-Clause",
	},
	{
		Name:        "Ticking Clock, A.wav",
		URL:         "https://www.jshaw.co.uk/inspectorj-freesound-library",
		Author:      "InspectorJ (www.jshaw.co.uk) of Freesound.org",
		LicenseType: "CC BY 3.0",
	},
	{
		Name:        "GEM Projector 4.m4a",
		URL:         "https://freesound.org/people/guyburns/sounds/432520/",
		Author:      "guyburns",
		LicenseType: "CC0 1.0",
	},
	{
		Name:        "horn_fail_wahwah_2.wav ",
		URL:         "https://freesound.org/people/TaranP/sounds/362205/",
		Author:      "TaranP",
		LicenseType: "CC BY 3.0",
	},
	{
		Name:        "Applause",
		URL:         "https://freesound.org/people/lebaston100/sounds/253712/",
		Author:      "lebaston100",
		LicenseType: "CC BY 3.0",
	},
	{
		Name:        "Bell, Counter, A.wav",
		URL:         "https://www.jshaw.co.uk/inspectorj-freesound-library",
		Author:      "InspectorJ (www.jshaw.co.uk) of Freesound.org",
		LicenseType: "CC BY 3.0",
	},
	{
		Name:        "Alarm1.mp3",
		URL:         "https://freesound.org/people/kwahmah_02/sounds/250629/",
		Author:      "kwahmah_02",
		LicenseType: "CC BY 3.0",
	},
	{
		Name:        "Alert4",
		URL:         "https://freesound.org/people/RICHERlandTV/sounds/351539/",
		Author:      "RICHERlandTV",
		LicenseType: "CC BY 3.0",
	},
}
