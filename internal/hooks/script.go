package hooks

func Script(fpPath string, source string) string {
	return "#!/bin/sh\n" +
		"FP_SOURCE=" + source + " " +
		fpPath + " record >/dev/null 2>&1\n"
}
