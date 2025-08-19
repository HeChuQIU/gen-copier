package test

import (
	"hechuqiu.github.io/gen-copier/test/packageA"
	"hechuqiu.github.io/gen-copier/test/packageB"
)

//go:generate gen-copier packageA.TestSource packageB.TestTarget $GOFILE

func test() {
	ts := &packageA.TestSource{}
	tg := &packageB.TestTarget{}
	ts.CopyTo(tg)
}
