package main

import "slices"

func (f *FieldDef) GoType() string {
	vt := f.Type
	if f.Null {
		switch f.Type {
		case "autoincrement":
			vt = "sql.NullInt64"
		case "version":
			vt = "sql.NullInt64"
		case "text":
			vt = "sql.NullString"
		case "password", "secret":
			vt = "sql.NullString"
		case "timestamp":
			vt = "time.Time"
		case "many-to-one":
			vt = "*" + f.Ref
		case "many-to-many":
			vt = "[]" + f.Ref
		}
	} else {
		switch f.Type {
		case "autoincrement":
			vt = "int64"
		case "version":
			vt = "int64"
		case "text":
			vt = "string"
		case "password", "secret":
			vt = "string"
		case "timestamp":
			vt = "time.Time"
		case "many-to-one":
			vt = "*" + f.Ref
		case "many-to-many":
			vt = "[]" + f.Ref
		}
	}
	return vt
}

func (f *FieldDef) GoSqlNullType() string {
	vt := f.Type
	switch f.Type {
	case "autoincrement":
		vt = "sql.NullInt64"
	case "ver":
		vt = "sql.NullInt64"
	case "text":
		vt = "sql.NullString"
	case "password", "secret":
		vt = "sql.NullString"
	case "many-to-one":
		vt = "*" + f.Ref
	}
	if f.Null {
		vt = "*" + vt
	}
	return vt
}

func (f *FieldDef) GoSqlNullValue() string {
	vt := f.Type
	switch f.Type {
	case "autoincrement":
		vt = "Int64"
	case "version":
		vt = "Int64"
	case "text":
		vt = "String"
	case "password", "secret":
		vt = "String"
	case "many-to-one":
		vt = ""
	}
	if f.Null {
		vt = "*" + vt
	}
	return vt
}

func (f *FieldDef) IsPk() bool {
	return slices.Contains(f.Model().CPKeys, f.ID)
}

func (f *FieldDef) IsUpdatable() bool {
	if f.Type == "password" || f.Type == "version" {
		return false
	}
	return !slices.Contains(f.Model().CPKeys, f.ID)
}
