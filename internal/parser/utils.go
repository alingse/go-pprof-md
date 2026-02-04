package parser

// Helper functions for parsing protobuf profiles

func getString(prof *protoProfile, idx int64) string {
	if idx < 0 || int(idx) >= len(prof.StringTable) {
		return ""
	}
	return prof.StringTable[idx]
}

func findLocation(prof *protoProfile, id uint64) *protoLocation {
	for _, loc := range prof.Location {
		if loc.Id == id {
			return loc
		}
	}
	return nil
}

func findFunction(prof *protoProfile, id uint64) *protoFunction {
	if id == 0 {
		return nil
	}
	for _, fn := range prof.Function {
		if fn.Id == id {
			return fn
		}
	}
	return nil
}

func buildCallStack(prof *protoProfile, locationIdx []uint64) []string {
	stack := []string{}
	for _, locIdx := range locationIdx {
		loc := findLocation(prof, locIdx)
		if loc == nil {
			continue
		}

		for _, line := range loc.Line {
			fn := findFunction(prof, line.FunctionId)
			if fn != nil {
				funcName := getString(prof, fn.Name)
				if funcName != "" {
					stack = append(stack, funcName)
				}
			}
		}
	}
	return stack
}
