package main

import (
	"os"
)

const version = "0.1.0-golang"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printGlobalHelp()
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		printGlobalHelp()
		os.Exit(2)
	}

	panic("not implemented")
}

func printGlobalHelp() {
	panic("not implemented")
}

func cmdClone(args []string, triesPath string) []map[string]interface{} {
	panic("not implemented")
}

func cmdInit(args []string, triesPath string) {
	panic("not implemented")
}

func cmdCd(args []string, triesPath string, andType, andConfirm string, andExit bool, andKeys []string) []map[string]interface{} {
	panic("not implemented")
}

func parseTestKeys(spec string) []string {
	panic("not implemented")
}

func extractOptionWithValue(args *[]string, optName string) string {
	panic("not implemented")
}

func isFish() bool {
	panic("not implemented")
}
