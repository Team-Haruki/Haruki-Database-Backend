//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	if err := entc.Generate("./schema", &gen.Config{
		Package: "haruki-database/database/schema/chunithm/music",
		Target:  "../../../../database/schema/chunithm/music",
	}); err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
