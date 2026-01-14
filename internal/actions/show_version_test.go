package actions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShowVersion_PrintsVersion(t *testing.T) {
	var printed string

	deps := actionDependencies{
		Printf: func(format string, a ...any) (int, error) {
			printed = fmt.Sprintf(format, a...)
			return len(printed), nil
		},
		Version: func() string {
			return "1.2.3"
		},
	}

	err := showVersion(nil, nil, deps)

	require.NoError(t, err)
	require.Equal(t, "fp version 1.2.3\n", printed)
}
