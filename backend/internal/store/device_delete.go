package store

import (
	"context"
	"database/sql"
	"errors"
)

// DeleteDevice removes a device's structured content in one transaction.
// Shared (source, name) links have no interface: keep other scopes when
// another Function still owns that pair, rather than deleting shared data.
func DeleteDevice(ctx context.Context, db *sql.DB, name, now string) error {
	if name == "" {
		return ErrDeviceNotFound
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := New(tx)
	if _, err := q.GetRegisteredDevice(ctx, name); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		uses, err := q.CountDeviceNameUsages(ctx, name)
		if err != nil {
			return err
		}
		if uses == 0 {
			return ErrDeviceNotFound
		}
	}

	for _, table := range []string{"parameter_links", "criteria_links", "examples"} {
		_, err := tx.ExecContext(ctx, `DELETE FROM `+table+` AS content
			WHERE device_name = $1 OR (
				EXISTS (SELECT 1 FROM functions f WHERE f.interface = $1
					AND f.source = content.function_source AND f.name = content.function_name)
				AND NOT EXISTS (SELECT 1 FROM functions f WHERE f.interface <> $1
					AND f.source = content.function_source AND f.name = content.function_name)
			)`, name)
		if err != nil {
			return err
		}
	}
	for _, statement := range []string{
		`DELETE FROM parameter_definitions WHERE device_name = $1`,
		`DELETE FROM criteria_definitions WHERE device_name = $1`,
		`DELETE FROM function_grammars WHERE function_interface = $1`,
		`DELETE FROM script_key_audits WHERE function_interface = $1`,
	} {
		if _, err := tx.ExecContext(ctx, statement, name); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sequences SET updated_at = $2 WHERE id IN
		(SELECT sequence_id FROM sequence_steps WHERE function_interface = $1 OR device_name = $1)`, name, now); err != nil {
		return err
	}
	// Re-number only affected sequences, preserving their surviving step IDs/order.
	if _, err := tx.ExecContext(ctx, `WITH affected AS (
		SELECT DISTINCT sequence_id FROM sequence_steps
		WHERE function_interface = $1 OR device_name = $1
	), surviving AS (
		SELECT id, ROW_NUMBER() OVER (PARTITION BY sequence_id ORDER BY position, id) AS position
		FROM sequence_steps WHERE sequence_id IN (SELECT sequence_id FROM affected)
		AND function_interface <> $1 AND device_name <> $1
	)
	UPDATE sequence_steps SET position = surviving.position
	FROM surviving WHERE sequence_steps.id = surviving.id`, name); err != nil {
		return err
	}
	for _, statement := range []string{
		`DELETE FROM sequence_steps WHERE function_interface = $1 OR device_name = $1`,
		`DELETE FROM functions WHERE interface = $1`,
		`DELETE FROM devices WHERE name = $1`,
	} {
		if _, err := tx.ExecContext(ctx, statement, name); err != nil {
			return err
		}
	}
	return tx.Commit()
}
