package store

import (
	"context"
	"database/sql"
)

func RenameFunctionAssociations(ctx context.Context, tx *sql.Tx, old, renamed Function) error {
	if old.Source == renamed.Source && old.Name == renamed.Name {
		return nil
	}
	shared, err := New(tx).FunctionNaturalKeyExists(ctx, FunctionNaturalKeyExistsParams{
		Source: old.Source,
		Name:   old.Name,
	})
	if err != nil {
		return err
	}

	for _, association := range []struct{ table, key, payload string }{
		{"parameter_links", "name", "name, position, required, source_file"},
		{"criteria_links", "name", "name, position, required, source_file"},
		{"examples", "script", "script, meaning"},
	} {
		args := []any{old.Source, old.Name, renamed.Source, renamed.Name, renamed.UpdatedAt}
		if shared {
			if _, err := tx.ExecContext(ctx, `INSERT INTO `+association.table+`
				(function_source, function_name, device_name, created_at, updated_at, `+association.payload+`)
				SELECT $3, $4, device_name, created_at, $5, `+association.payload+`
				FROM `+association.table+` WHERE function_source = $1 AND function_name = $2
				ON CONFLICT (function_source, function_name, device_name, `+association.key+`) DO NOTHING`, args...); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+association.table+` AS old
			USING `+association.table+` AS target
			WHERE old.function_source = $1 AND old.function_name = $2
				AND target.function_source = $3 AND target.function_name = $4
				AND target.device_name = old.device_name AND target.`+association.key+` = old.`+association.key, args[:4]...); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE `+association.table+`
			SET function_source = $3, function_name = $4, updated_at = $5
			WHERE function_source = $1 AND function_name = $2`, args...); err != nil {
			return err
		}
	}
	return nil
}
