package types

// Same checks if the two input types are same or not.
func Same(typ1 Type, typ2 Type) bool {
	switch v1 := typ1.(type) {
	case *Int:
		_, ok := typ2.(*Int)
		return ok
	case *Bool:
		_, ok := typ2.(*Bool)
		return ok
	case *String:
		_, ok := typ2.(*String)
		return ok
	case *Range:
		_, ok := typ2.(*Range)
		return ok
	case *Array:
		v2, ok := typ2.(*Array)
		if !ok {
			return false
		}
		if v1.Len != v2.Len {
			return false
		}
		return Same(v1.ElemType, v2.ElemType)
	case *Func:
		v2, ok := typ2.(*Func)
		if !ok {
			return false
		}
		if len(v1.ParamTypes) != len(v2.ParamTypes) {
			return false
		}
		for i := range v1.ParamTypes {
			if !Same(v1.ParamTypes[i], v2.ParamTypes[i]) {
				return false
			}
		}
		return Same(v1.ReturnType, v2.ReturnType)
	default:
		// typ1 is nil
		return typ2 == nil
	}
}
