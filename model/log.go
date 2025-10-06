package model

import (
	"database/sql"
	"log/slog"
)

func logNullString(key string, v sql.NullString) slog.Attr {
	if v.Valid {
		return slog.String(key, v.String)
	}
	return slog.Any(key, nil)
}

func logNullInt64(key string, v sql.NullInt64) slog.Attr {
	if v.Valid {
		return slog.Int64(key, v.Int64)
	}
	return slog.Any(key, nil)
}
