package dns

func RegPart(name *Name) (registered *Name, registrar *Name) {
	if name.IsRoot() {
		return name, name
	}

	cur := name
	parent := cur.Parent() // parent is not nil
	grand := parent.Parent()

	parentStr := parent.String()
	var grandStr string
	if grand != nil {
		grandStr = grand.String()
	}

	for {
		if grand != nil && superNames[grandStr] && !notRegNames[parentStr] {
			return cur, parent
		}

		if regNames[parentStr] {
			return cur, parent
		}

		if parent.IsRoot() {
			return cur, parent
		}

		// shift now, parent is not root
		cur = parent

		parent = grand
		parentStr = grandStr

		grand = grand.Parent()
		if grand != nil {
			grandStr = grand.String()
		}
	}

	// fail safe only
	return nil, nil
}
