package main

func (f *FieldDef) MapKeys() ([]FieldRef, error) {
	res := []FieldRef{}
	for _, ms := range f.CMapKeys {
		ref, err := f.Model().Field(ms.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, FieldRef{
			ID:    ms.ID,
			Field: ms.Field,
			Ref:   ref,
		})
	}
	return res, nil
}
