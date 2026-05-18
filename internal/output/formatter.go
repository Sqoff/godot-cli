package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
	}
}

func PrintText(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
