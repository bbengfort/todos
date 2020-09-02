// +build ignore

package main

import (
	"log"

	"github.com/bbengfort/todos/web"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(web.WebApp, vfsgen.Options{
		PackageName:  "web",
		BuildTags:    "!dev",
		VariableName: "WebApp",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
