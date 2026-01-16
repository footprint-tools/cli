package dispatchers

type CommandFunc func(args []string, flags []string) error

type Resolution struct {
	Node     *DispatchNode
	Args     []string
	Flags    []string
	Execute  CommandFunc
	ExitCode int // Non-zero to exit with specific code after Execute
}

type FlagScope int

const (
	FlagScopeGlobal FlagScope = iota
	FlagScopeLocal
)

type FlagDescriptor struct {
	Names       []string
	ValueHint   string
	Description string
	Scope       FlagScope
}

type ArgSpec struct {
	Name        string
	Description string
	Required    bool
}

type DispatchNode struct {
	Name        string
	Path        []string
	Summary     string            // One-line summary for listings
	Description string            // Longer explanation for individual help
	Usage       string
	Flags       []FlagDescriptor
	Args        []ArgSpec
	Children    map[string]*DispatchNode
	Action      CommandFunc
	Category    CommandCategory
}
