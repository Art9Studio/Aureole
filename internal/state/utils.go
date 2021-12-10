package state

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

func ListPluginStatus(p *Project) {
	fmt.Println("AUREOLE PLUGINS STATUS")

	for appName, app := range p.Apps {
		fmt.Printf("\nAPP: %s\n", appName)

		printStatus("identity manager", p.Apps[appName].IdentityManager)

		for name, authn := range app.Authenticators {
			printStatus(name, authn)
		}

		printStatus("authorizer", p.Apps[appName].Authorizer)
	}

	if len(p.Storages) != 0 {
		fmt.Println("\nSTORAGE PLUGINS")
		for name, plugin := range p.Storages {
			printStatus(name, plugin)
		}
	}

	if len(p.KeyStorages) != 0 {
		fmt.Println("\nKEY STORAGE PLUGINS")
		for name, plugin := range p.KeyStorages {
			printStatus(name, plugin)
		}
	}

	if len(p.Hashers) != 0 {
		fmt.Println("\nHASHER PLUGINS")
		for name, plugin := range p.Hashers {
			printStatus(name, plugin)
		}
	}

	if len(p.Senders) != 0 {
		fmt.Println("\nSENDER PLUGINS")
		for name, plugin := range p.Senders {
			printStatus(name, plugin)
		}
	}

	if len(p.CryptoKeys) != 0 {
		fmt.Println("\nCRYPTOKEY PLUGINS")
		for name, plugin := range p.CryptoKeys {
			printStatus(name, plugin)
		}
	}

	if len(p.Admins) != 0 {
		fmt.Println("\nADMIN PLUGINS")
		for name, plugin := range p.Admins {
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

	if plugin != nil && !reflect.ValueOf(plugin).IsNil() {
		fmt.Printf("%s%s - %v%s\n", colorGreen, name, string(checkMark), resetColor)
	} else {
		fmt.Printf("%s%s - %v%s\n", colorRed, name, string(crossMark), resetColor)
	}
}
