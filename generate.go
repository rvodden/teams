//go:build ignore

package main

import (
	"github.com/rvodden/teams/internal/codegen"
	"github.com/rvodden/teams/model"
)

func main() {
	codegen.GenerateCodeFile("person", "people", model.Person{})
	codegen.GenerateCodeFile("team", "teams", model.Team{})
}
