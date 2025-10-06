package main

import (
	"slices"
)

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
		case "int", "number":
			vt = "sql.NullInt64"
		case "json":
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
		case "int", "number":
			vt = "int64"
		case "json":
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
	case "json":
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
	case "json":
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

func (f *FieldDef) GoLogType() string {
	return f.goLogType(f.Null)
}

func (f *FieldDef) GoLogNullType() string {
	return f.goLogType(true)
}

func (f *FieldDef) goLogType(isNull bool) string {
	vt := f.Type
	switch f.Type {
	case "autoincrement", "version":
		if isNull {
			vt = "logNullInt64"
		} else {
			vt = "slog.Int64"
		}
	case "int":
		if isNull {
			vt = "logNullInt64"
		} else {
			vt = "slog.Int"
		}
	case "int64":
		if isNull {
			vt = "logNullInt64"
		} else {
			vt = "slog.Int64"
		}
	case "text", "json", "password", "secret":
		if isNull {
			vt = "logNullString"
		} else {
			vt = "slog.String"
		}
	case "timestamp":
		if isNull {
			vt = "slog.Time"
		} else {
			vt = "slog.Time"
		}
	case "many-to-one":
		vt = "slog.Any"
	default:
		vt = "slogUnknown" + f.Type
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
