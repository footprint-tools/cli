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

func Root(spec RootSpec) *DispatchNode {
	return NewNode(
		spec.Name,
		nil,
		spec.Summary,
		spec.Usage,
		spec.Flags,
		nil,
		nil,
	)
}

func Group(spec GroupSpec) *DispatchNode {
	return NewNode(
		spec.Name,
		spec.Parent,
		spec.Summary,
		spec.Usage,
		nil,
		nil,
		nil,
	)
}

func Command(spec CommandSpec) *DispatchNode {
	node := NewNode(
		spec.Name,
		spec.Parent,
		spec.Summary,
		spec.Usage,
		spec.Flags,
		spec.Args,
		spec.Action,
	)

	node.Category = spec.Category
	return node
}
