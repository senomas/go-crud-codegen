package main

import (
	"fmt"
	"os"
	"strings"
)

func GenRepoGet(f *os.File, name string, md *ModelDef) error {
	fk := []string{}
	where := []string{}
	wargs := []string{}
	idx := 1
	for _, pk := range md.PKey {
		ok := false
		for _, fd := range md.Fields {
			if fd.ID == pk {
				fk = append(fk, fmt.Sprintf("%s %s", fd.Field, fd.GoType))
				where = append(where, fmt.Sprintf("%s = $%d", fd.Field, idx))
				wargs = append(wargs, fmt.Sprintf("%s", fd.Field))
				idx++
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("model %s pkey %s not found", name, pk)
		}
	}

	args := []string{}
	fields := []string{}
	for _, fd := range md.Fields {
		args = append(args, fmt.Sprintf("    &obj.%s,\n", fd.ID))
		fields = append(fields, fmt.Sprintf("      %s", fd.Field))
	}
	_, err := fmt.Fprintf(f, "\nfunc (r *%sRepositoryImpl) Get(%s) (*%s, error) {\n", name, strings.Join(fk, ", "), name)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  sql := `\n    SELECT\n%s\n    FROM %s\n    WHERE\n      %s`\n",
		strings.Join(fields, ",\n"), md.Table, strings.Join(where, " AND\n    "))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  var obj %s;\n  err := r.db.QueryRowContext(r.ctx, sql, %s).Scan(\n      %s)\n",
		name, strings.Join(wargs, ","), strings.Join(args, ""))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  if err != nil {\n  return nil, err\n  }\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  return &obj, nil\n}\n")
	return err
}
