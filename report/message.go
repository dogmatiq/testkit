package report

import (
	"github.com/dogmatiq/dapper"
	"github.com/dogmatiq/dogma"
)

// RenderMessage returns a human-readable representation of v.
func RenderMessage(v dogma.Message) string {
	return printer.Format(v)
}

// printer is the Dapper printer used to render values.
var printer dapper.Printer

func init() {
	printer = dapper.Printer{
		// Copy the default config.
		Config: dapper.DefaultPrinter.Config,
	}

	// Then modify the settings we want to change.
	printer.Config.OmitPackagePaths = true
}
