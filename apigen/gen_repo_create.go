package main

import (
	"fmt"
	"os"
	"strings"
)

func GenRepoCreate(f *os.File, name string, md *ModelDef) error {
	_, err := fmt.Fprintf(f, "\nfunc (r *%sRepositoryImpl) Create(obj %s) (*%s, error) {\n", name, name, name)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(f, "  tx, err := r.db.BeginTx(r.ctx, nil)\n  if err != nil {\n    return nil, err\n  }\n  defer tx.Rollback()\n")
	if err != nil {
		return err
	}

	pargs := []string{}
	fields := []string{}
	args := []string{}
	pidx := 1
	for _, fd := range md.Fields {
		if fd.Type != "autoincrement" {
			fields = append(fields, fmt.Sprintf("      %s", fd.Field))
			if pidx == 1 {
				pargs = append(pargs, fmt.Sprintf("$%d", pidx))
			} else if (pidx-1)%5 == 0 {
				pargs = append(pargs, fmt.Sprintf(",\n      $%d", pidx))
			} else {
				pargs = append(pargs, fmt.Sprintf(", $%d", pidx))
			}
			pidx++
			args = append(args, fmt.Sprintf("    obj.%s,\n", fd.ID))
		}
	}
	_, err = fmt.Fprintf(f, "  sql := `\n    INSERT INTO %s (\n%s)\n    VALUES (\n      %s)`\n",
		md.Table, strings.Join(fields, ",\n"), strings.Join(pargs, ""))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  res, err := r.db.ExecContext(r.ctx, sql,\n    %s\n  )\n", strings.Join(args, ""))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  if err != nil {\n  return nil, err\n  }\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  obj.ID, err = res.LastInsertId()\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  if err != nil {\n  return nil, err\n  }\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "  err = tx.Commit()\n")
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
