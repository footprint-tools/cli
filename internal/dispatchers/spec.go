package dispatchers

type RootSpec struct {
	Name    string
	Summary string
	Usage   string
	Flags   []FlagDescriptor
}

type GroupSpec struct {
	Name    string
	Parent  *DispatchNode
	Summary string
	Usage   string
}

type CommandSpec struct {
	Name     string
	Parent   *DispatchNode
	Summary  string
	Usage    string
	Flags    []FlagDescriptor
	Args     []ArgSpec
	Action   CommandFunc
	Category CommandCategory
}
