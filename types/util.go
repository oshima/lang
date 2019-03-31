package types

// Same checks if two input types are same or not.
func Same(ty1 Type, ty2 Type) bool {
	switch v1 := ty1.(type) {
	case *Int:
		_, ok := ty2.(*Int)
		return ok
	case *Bool:
		_, ok := ty2.(*Bool)
		return ok
	case *String:
		_, ok := ty2.(*String)
		return ok
	case *Range:
		_, ok := ty2.(*Range)
		return ok
	case *Array:
		v2, ok := ty2.(*Array)
		if !ok {
			return false
		}
		if v1.Len != v2.Len {
			return false
		}
		return Same(v1.ElemType, v2.ElemType)
	case *Func:
		v2, ok := ty2.(*Func)
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
		// ty1 is nil
		return ty2 == nil
	}
}
