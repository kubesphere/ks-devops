package main

import (
	"kubesphere.io/devops/cmd/tools/jwt/app"
	"log"
)

func main() {
	if err := app.NewCmd(&app.DefaultK8sClientFactory{}).Execute(); err != nil {
		log.Fatalln(err)
	}
}
