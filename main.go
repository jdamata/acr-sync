package main

import (
	cmd "github.com/jdamata/acr-sync/cmd"
)

var version = "dev"

func main() {
	cmd.Execute(version)
}