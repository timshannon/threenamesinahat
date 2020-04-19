// +build ignore

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"github.com/timshannon/threenamesinahat/files"
)

func main() {
	err := vfsgen.Generate(files.Assets, vfsgen.Options{
		PackageName:  "files",
		BuildTags:    "!dev",
		VariableName: "Assets",
		Filename:     "./files/assets_vfsdata.go",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
