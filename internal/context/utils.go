package context

import (
	"fmt"
	"unicode/utf8"
)

func ListPluginStatus(ctx *ProjectCtx) {
	fmt.Println("AUREOLE PLUGINS STATUS")

	for appName, app := range ctx.Apps {
		fmt.Printf("\nAPP: %s\n", appName)

		for name, authn := range app.Authenticators {
			printStatus(name, authn)
		}

		for name, authz := range app.Authorizers {
			printStatus(name, authz)
		}

		printStatus("identity", app.Identity)
	}

	if len(ctx.Collections) != 0 {
		fmt.Println("\nCOLLECTION PLUGINS")
		for name, plugin := range ctx.Collections {
			printStatus(name, plugin)
		}
	}

	if len(ctx.Storages) != 0 {
		fmt.Println("\nSTORAGE PLUGINS")
		for name, plugin := range ctx.Storages {
			printStatus(name, plugin)
		}
	}

	if len(ctx.Hashers) != 0 {
		fmt.Println("\nHASHER PLUGINS")
		for name, plugin := range ctx.Hashers {
			printStatus(name, plugin)
		}
	}

	if len(ctx.Senders) != 0 {
		fmt.Println("\nSENDER PLUGINS")
		for name, plugin := range ctx.Senders {
			printStatus(name, plugin)
		}
	}

	if len(ctx.CryptoKeys) != 0 {
		fmt.Println("\nCRYPTOKEY PLUGINS")
		for name, plugin := range ctx.CryptoKeys {
			printStatus(name, plugin)
		}
	}

	if len(ctx.Admins) != 0 {
		fmt.Println("\nADMIN PLUGINS")
		for name, plugin := range ctx.Admins {
			printStatus(name, plugin)
		}
	}
}

func printStatus(name string, plugin interface{}) {
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	resetColor := "\033[0m"

	checkMark, _ := utf8.DecodeRuneInString("\u2714")
	crossMark, _ := utf8.DecodeRuneInString("\u274c")

	if plugin != nil {
		fmt.Printf("%s%s - %v%s\n", colorGreen, name, string(checkMark), resetColor)
	} else {
		fmt.Printf("%s%s - %v%s\n", colorRed, name, string(crossMark), resetColor)
	}
}
