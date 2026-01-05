package dispatchers

func NewNode(
	name string,
	parent *DispatchNode,
	summary string,
	usage string,
	flags []FlagDescriptor,
	args []ArgSpec,
	action CommandFunc,
) *DispatchNode {

	node := &DispatchNode{
		Name:     name,
		Summary:  summary,
		Usage:    usage,
		Flags:    flags,
		Args:     args,
		Action:   action,
		Children: make(map[string]*DispatchNode),
	}

	if parent == nil {
		node.Path = []string{name}
	} else {
		node.Path = append(parent.Path, name)
		parent.Children[name] = node
	}

	return node
}
