package main

import "fmt"

func mapSelf(m any, f any) ([]map[string]any, error) {
	if md, ok := m.(ModelDef); ok {
		if fd, ok := f.(FieldDef); ok {
			if rk, ok := fd.Extras["mapSelf"].([]any); ok {
				res := []map[string]any{}
				for _, k := range rk {
					if kmap, ok := k.(map[string]any); ok {
						if kk, ok := kmap["key"].(string); ok {
							nf := true
							for _, rf := range md.Fields {
								if kk == rf.ID {
									res = append(res, map[string]any{"key": k, "ref": rf})
									nf = false
								}
							}
							if nf {
								return nil, fmt.Errorf("model has no fields")
							}
						} else {
							return nil, fmt.Errorf("fd.Extras['mapSelf'] contains map without string id: %v", kmap)
						}
					} else {
						return nil, fmt.Errorf("fd.Extras['mapSelf'] contains non map[string]any: %T", k)
					}
				}
				return res, nil
			}
			return nil, fmt.Errorf("fd.Extras['mapSelf'] is not []any: %T", fd.Extras["mapSelf"])
		}
		return nil, fmt.Errorf("f is not FieldDef: %T", f)
	}
	return nil, fmt.Errorf("m is not ModelDef: %T", m)
}

func mapRef(m any, f any) ([]map[string]any, error) {
	if md, ok := m.(ModelDef); ok {
		if fd, ok := f.(FieldDef); ok {
			if fmod, ok := md.Extras["model"].(func(string) (*ModelDef, error)); ok {
				if ref, ok := fd.Extras["ref"].(string); ok {
					rm, err := fmod(ref)
					if err != nil {
						return nil, err
					}
					if rk, ok := fd.Extras["mapRef"].([]any); ok {
						res := []map[string]any{}
						for _, k := range rk {
							if kmap, ok := k.(map[string]any); ok {
								if kk, ok := kmap["key"].(string); ok {
									nf := true
									for _, rf := range rm.Fields {
										if kk == rf.ID {
											res = append(res, map[string]any{"key": k, "ref": rf, "refModel": rm})
											nf = false
										}
									}
									if nf {
										return nil, fmt.Errorf("ref model %s has no fields", ref)
									}
								} else {
									return nil, fmt.Errorf("fd.Extras['mapRef'] contains map without string id: %v", kmap)
								}
							} else {
								return nil, fmt.Errorf("fd.Extras['mapRef'] contains non map[string]any: %T", k)
							}
						}
						return res, nil
					}
					return nil, fmt.Errorf("fd.Extras['mapRef'] is not []any: %T", fd.Extras["mapRef"])
				}
				return nil, fmt.Errorf("fd.Extras['ref'] is not string: %T", fd.Extras["ref"])
			}
			return nil, fmt.Errorf("m.Extras['model'] is not func(string) (*ModelDef, error): %T", md.Extras["model"])
		}
		return nil, fmt.Errorf("f is not FieldDef: %T", f)
	}
	return nil, fmt.Errorf("m is not ModelDef: %T", m)
}

func mapFields(m any, f any) ([]map[string]any, error) {
	if md, ok := m.(ModelDef); ok {
		if fd, ok := f.(FieldDef); ok {
			if fmod, ok := md.Extras["model"].(func(string) (*ModelDef, error)); ok {
				if ref, ok := fd.Extras["ref"].(string); ok {
					rm, err := fmod(ref)
					if err != nil {
						return nil, err
					}
					if rk, ok := fd.Extras["mapFields"].([]any); ok {
						res := []map[string]any{}
						for _, k := range rk {
							if kk, ok := k.(string); ok {
								nf := true
								for _, rf := range rm.Fields {
									if kk == rf.ID {
										res = append(res, map[string]any{"key": kk, "ref": rf, "refModel": rm})
										nf = false
									}
								}
								if nf {
									return nil, fmt.Errorf("ref model %s has no fields", ref)
								}
							} else {
								return nil, fmt.Errorf("fd.Extras['mapFields'] contains non string id: %T", k)
							}
						}
						return res, nil
					}
					return nil, fmt.Errorf("fd.Extras['mapFields'] is not []any: %T", fd.Extras["mapRef"])
				}
				return nil, fmt.Errorf("fd.Extras['ref'] is not string: %T", fd.Extras["ref"])
			}
			return nil, fmt.Errorf("m.Extras['model'] is not func(string) (*ModelDef, error): %T", md.Extras["model"])
		}
		return nil, fmt.Errorf("f is not FieldDef: %T", f)
	}
	return nil, fmt.Errorf("m is not ModelDef: %T", m)
}
