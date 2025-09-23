package model

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// RunSQLFile splits a SQL file into statements and executes them in one tx.
func RunSQLFile(ctx context.Context, db *sql.DB, filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	stmts, err := splitSQL(string(b))
	if err != nil {
		return err
	}
	if len(stmts) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, s := range stmts {
		if strings.TrimSpace(s) == "" {
			continue
		}
		if _, err = tx.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("exec failed: %w\nSQL:\n%s", err, s)
		}
	}
	return tx.Commit()
}

// splitSQL splits on top-level semicolons; understands quotes, dollar-quoted strings, and comments.
func splitSQL(src string) ([]string, error) {
	var out []string
	var sb strings.Builder

	inS, inD := false, false        // ' ' , " "
	inLine, inBlock := false, false // --, /* */
	var dollar string               // "", or "$tag$"
	r := []rune(src)

	for i := 0; i < len(r); i++ {
		c := r[i]

		// handle line/block comments
		if !inS && !inD && dollar == "" && !inLine && !inBlock {
			if c == '-' && i+1 < len(r) && r[i+1] == '-' {
				inLine = true
			} else if c == '/' && i+1 < len(r) && r[i+1] == '*' {
				inBlock = true
			}
		}

		if inLine {
			if c == '\n' {
				inLine = false
				sb.WriteRune(c)
			}
			continue
		}
		if inBlock {
			if c == '*' && i+1 < len(r) && r[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}

		// dollar-quoted strings: $tag$ ... $tag$
		if !inS && !inD {
			if dollar == "" && c == '$' {
				// try to read $tag$
				j := i + 1
				for j < len(r) && (unicode.IsLetter(r[j]) || unicode.IsDigit(r[j]) || r[j] == '_') {
					j++
				}
				if j < len(r) && r[j] == '$' {
					dollar = string(r[i : j+1]) // include both $
					sb.WriteString(dollar)
					i = j
					continue
				}
			} else if dollar != "" && c == '$' {
				// possible end $tag$
				if i+len(dollar)-1 < len(r) && string(r[i:i+len(dollar)]) == dollar {
					sb.WriteString(dollar)
					i += len(dollar) - 1
					dollar = ""
					continue
				}
			}
		}

		// normal quotes
		if dollar == "" {
			if c == '\'' && !inD {
				inS = !inS
			} else if c == '"' && !inS {
				inD = !inD
			}
		}

		// top-level semicolon splits statements
		if c == ';' && !inS && !inD && dollar == "" {
			out = append(out, sb.String())
			sb.Reset()
			continue
		}

		sb.WriteRune(c)
	}
	if strings.TrimSpace(sb.String()) != "" {
		out = append(out, sb.String())
	}
	return out, nil
}
