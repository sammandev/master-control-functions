package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// DeviceInUseError refuses a device delete while Sequence steps still use
// the device, either as a Function interface or as a step's device scope.
type DeviceInUseError struct {
	Device    string
	Sequences []string
}

func (e *DeviceInUseError) Error() string {
	return e.Device + " is used by Sequences: " + strings.Join(e.Sequences, ", ") + ". Remove those steps first."
}

// DeleteDevice removes a device's structured content in one transaction.
// Shared (source, name) links have no interface: keep other scopes when
// another Function still owns that pair, rather than deleting shared data.
func DeleteDevice(ctx context.Context, db *sql.DB, name string) error {
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

	sequences, err := sequencesUsingDevice(ctx, tx, name)
	if err != nil {
		return err
	}
	if len(sequences) > 0 {
		return &DeviceInUseError{Device: name, Sequences: sequences}
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
	for _, statement := range []string{
		`DELETE FROM functions WHERE interface = $1`,
		`DELETE FROM devices WHERE name = $1`,
	} {
		if _, err := tx.ExecContext(ctx, statement, name); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func sequencesUsingDevice(ctx context.Context, tx *sql.Tx, name string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT s.name FROM sequences s
		JOIN sequence_steps step ON step.sequence_id = s.id
		WHERE step.function_interface = $1 OR step.device_name = $1
		ORDER BY s.name`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var sequence string
		if err := rows.Scan(&sequence); err != nil {
			return nil, err
		}
		names = append(names, sequence)
	}
	return names, rows.Err()
}
