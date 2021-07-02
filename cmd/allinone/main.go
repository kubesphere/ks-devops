package main

import (
	"kubesphere.io/devops/cmd/allinone/app"
	"log"
)

func main() {
	cmd := app.NewCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
