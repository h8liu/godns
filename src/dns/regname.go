package dns

func RegParts(name *Name) (registered *Name, registrar *Name) {
    var last *Name
    cur := name
    curStr := cur.String()

    parent := cur.Parent()
    var parentStr string
    if parent != nil {
	    parentStr = parent.String()
    }

	for {
		if parent != nil && superNames[parentStr] && !notRegNames[curStr] {
			return last, cur
		}

        if cur.IsRoot() {
            return last, cur
        }

		if regNames[curStr] {
			return last, cur
		}

		// shift now, parent is not root
		last = cur
        cur = parent
        curStr = parentStr

        parent = parent.Parent()
        if parent != nil {
            parentStr = parent.String()
        }
	}

	// fail safe only
	return nil, nil
}

func IsRegistrar(name *Name) bool {
    _, reg := RegParts(name)
    return reg.Equal(name)
}
