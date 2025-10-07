package model

import (
	"log/slog"
)

func logNullString(key string, v jsql.NullString) slog.Attr {
	if v.Valid {
		return slog.String(key, v.String)
	}
	return slog.Any(key, nil)
}

func logNullInt64(key string, v jsql.NullInt64) slog.Attr {
	if v.Valid {
		return slog.Int64(key, v.Int64)
	}
	return slog.Any(key, nil)
}
