package main

import "fmt"

func refKeys(m any, f any) ([]map[string]any, error) {
	if md, ok := m.(ModelDef); ok {
		if fd, ok := f.(FieldDef); ok {
			if ref, ok := fd.Extras["ref"].(string); ok {
				if fmod, ok := md.Extras["model"].(func(string) (*ModelDef, error)); ok {
					rm, err := fmod(ref)
					if err != nil {
						return nil, err
					}
					if rm == nil {
						return nil, fmt.Errorf("ref model %s not found", ref)
					}
					if rk, ok := fd.Extras["refKeys"].([]any); ok {
						res := []map[string]any{}
						for _, k := range rk {
							if km, ok := k.(map[string]any); ok {
								if kr, ok := km["ref"].(string); ok {
									found := false
									for _, rfd := range rm.Fields {
										if rfd.ID == kr {
											res = append(res, map[string]any{
												"field": km["field"],
												"ref":   rfd,
											})
											found = true
										}
									}
									if !found {
										return nil, fmt.Errorf("ref field %s not found in model %s", kr, rm.ID)
									}
								}
							} else {
								return nil, fmt.Errorf("fd.Extras['refKeys'] contains non map[string]any: %T", k)
							}
						}
						return res, nil
					}
					return nil, fmt.Errorf("fd.Extras['refKeys'] is not []string: %T", fd.Extras["refKeys"])
				}
				return nil, fmt.Errorf("m.Extras['model'] is not func(string) (*ModelDef, error): %T", md.Extras["model"])
			}
			return nil, fmt.Errorf("fd.Extras['ref'] is not string: %T", fd.Extras["ref"])
		}
		return nil, fmt.Errorf("f is not FieldDef: %T", f)
	}
	return nil, fmt.Errorf("m is not ModelDef: %T", m)
}

func refFields(m any, f any) ([]any, error) {
	if md, ok := m.(ModelDef); ok {
		if fd, ok := f.(FieldDef); ok {
			if ref, ok := fd.Extras["ref"].(string); ok {
				if fmod, ok := md.Extras["model"].(func(string) (*ModelDef, error)); ok {
					rm, err := fmod(ref)
					if err != nil {
						return nil, err
					}
					if rm == nil {
						return nil, fmt.Errorf("ref model %s not found", ref)
					}
					if rk, ok := fd.Extras["refFields"].([]any); ok {
						res := []any{}
						for _, k := range rk {
							if kr, ok := k.(string); ok {
								found := false
								for _, rfd := range rm.Fields {
									if rfd.ID == kr {
										res = append(res, rfd)
										found = true
									}
								}
								if !found {
									return nil, fmt.Errorf("ref field %s not found in model %s", kr, rm.ID)
								}
							} else {
								return nil, fmt.Errorf("fd.Extras['refFields'] contains non map[string]any: %T", k)
							}
						}
						return res, nil
					}
					return nil, fmt.Errorf("fd.Extras['refFields'] is not []string: %T", fd.Extras["refKeys"])
				}
				return nil, fmt.Errorf("m.Extras['model'] is not func(string) (*ModelDef, error): %T", md.Extras["model"])
			}
			return nil, fmt.Errorf("fd.Extras['ref'] is not string: %T", fd.Extras["ref"])
		}
		return nil, fmt.Errorf("f is not FieldDef: %T", f)
	}
	return nil, fmt.Errorf("m is not ModelDef: %T", m)
}
