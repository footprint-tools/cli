package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Skryensya/footprint/internal/config"
)

var pagerDisabled bool

// DisablePager disables the pager globally.
func DisablePager() {
	pagerDisabled = true
}

func Pager(content string) {
	if pagerDisabled {
		fmt.Print(content)
		return
	}

	pager := "less"

	lines, err := config.ReadLines()
	if err == nil {
		if configMap, err := config.Parse(lines); err == nil {
			if p, ok := configMap["pager"]; ok && p != "" {
				pager = p
			}
		}
	}

	cmd := exec.Command(pager, "-FRSX")
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Print(content)
	}
}
