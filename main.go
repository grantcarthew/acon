package main

import "github.com/grantcarthew/acon/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
