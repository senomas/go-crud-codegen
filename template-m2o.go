package main

import "fmt"

func (f *FieldDef) RefKeys() ([]FieldRef, error) {
	res := []FieldRef{}
	md := f.Model()
	mdr, err := md.model(f.Ref)
	if err != nil {
		return nil, err
	}
	if mdr == nil {
		return nil, fmt.Errorf("referenced model %s for field %s is nil", f.Ref, f.ID)
	}
	for _, ms := range f.CRefKeys {
		v, err := mdr.Field(ms.ID)
		ref := *v
		if err != nil {
			return nil, err
		}
		res = append(res, FieldRef{
			ID:    ms.ID,
			Field: ms.Field,
			Ref:   &ref,
		})
	}
	return res, nil
}

func (f *FieldDef) RefFields() ([]FieldRef, error) {
	res := []FieldRef{}
	md := f.Model()
	mdr, err := md.model(f.Ref)
	if err != nil {
		return nil, err
	}
	for _, ms := range f.CRefFields {
		v, err := mdr.Field(ms)
		ref := *v
		if err != nil {
			return nil, err
		}
		res = append(res, FieldRef{
			ID:    ms,
			Field: ms,
			Ref:   &ref,
		})
	}
	return res, nil
}
