package paths

import "os"

func configDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return ""
	}
	return appData + "\\Footprint"
}
