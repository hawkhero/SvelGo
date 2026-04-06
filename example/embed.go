package main

import (
	"embed"
	"io/fs"
	"log"

	svelgo "github.com/svelgo/svelgo"
)

//go:embed all:static
var embeddedStatic embed.FS

func init() {
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		log.Fatal("embed: could not sub static/:", err)
	}
	svelgo.SetStaticFS(sub)
}
