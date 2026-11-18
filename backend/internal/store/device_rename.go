package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrDeviceNotFound = errors.New("device not found")
var ErrDeviceExists = errors.New("a device with that name already exists")

func RenameDevice(ctx context.Context, db *sql.DB, oldName, newName, now string) (Device, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Device{}, err
	}
	defer tx.Rollback()
	queries := New(tx)
	current, err := queries.GetRegisteredDevice(ctx, oldName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Device{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		uses, countErr := queries.CountDeviceNameUsages(ctx, oldName)
		if countErr != nil {
			return Device{}, countErr
		}
		if uses == 0 {
			return Device{}, ErrDeviceNotFound
		}
		current = Device{Name: oldName, CreatedAt: now, UpdatedAt: now}
	}
	if oldName == newName {
		return current, nil
	}
	if _, err := queries.GetRegisteredDevice(ctx, newName); err == nil {
		return Device{}, ErrDeviceExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Device{}, err
	}
	uses, err := queries.CountDeviceNameUsages(ctx, newName)
	if err != nil {
		return Device{}, err
	}
	if uses > 0 {
		return Device{}, ErrDeviceExists
	}
	statements := []string{
		`UPDATE functions SET interface = $2, updated_at = $3 WHERE interface = $1`,
		`UPDATE parameter_links SET device_name = $2, updated_at = $3 WHERE device_name = $1`,
		`UPDATE criteria_links SET device_name = $2, updated_at = $3 WHERE device_name = $1`,
		`UPDATE parameter_definitions SET device_name = $2, updated_at = $3 WHERE device_name = $1`,
		`UPDATE criteria_definitions SET device_name = $2, updated_at = $3 WHERE device_name = $1`,
		`UPDATE examples SET device_name = $2, updated_at = $3 WHERE device_name = $1`,
		`UPDATE function_grammars SET function_interface = $2, updated_at = $3 WHERE function_interface = $1`,
		`UPDATE script_key_audits SET function_interface = $2, updated_at = $3 WHERE function_interface = $1`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, oldName, newName, now); err != nil {
			return Device{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sequences SET updated_at = $2 WHERE id IN
		(SELECT sequence_id FROM sequence_steps WHERE function_interface = $1 OR device_name = $1)`, oldName, now); err != nil {
		return Device{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sequence_steps SET
		function_interface = CASE WHEN function_interface = $1 THEN $2 ELSE function_interface END,
		device_name = CASE WHEN device_name = $1 THEN $2 ELSE device_name END
		WHERE function_interface = $1 OR device_name = $1`, oldName, newName); err != nil {
		return Device{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM devices WHERE name = $1`, oldName); err != nil {
		return Device{}, err
	}
	renamed, err := queries.InsertDevice(ctx, InsertDeviceParams{Name: newName, CreatedAt: current.CreatedAt, UpdatedAt: now, Source: current.Source})
	if err != nil {
		return Device{}, err
	}
	if err := tx.Commit(); err != nil {
		return Device{}, err
	}
	return renamed, nil
}
