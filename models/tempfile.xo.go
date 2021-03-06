// Package models contains the types for schema 'alpha'.
package models

// GENERATED BY XO. DO NOT EDIT.

import (
	"errors"
	"time"
)

// TempFile represents a row from 'alpha.temp_files'.
type TempFile struct {
	ID          int        `json:"id"`           // id
	MessageID   string     `json:"message_id"`   // message_id
	FileName    string     `json:"file_name"`    // file_name
	TimeCreated *time.Time `json:"time_created"` // time_created

	// xo fields
	_exists, _deleted bool
}

// Exists determines if the TempFile exists in the database.
func (tf *TempFile) Exists() bool {
	return tf._exists
}

// Deleted provides information if the TempFile has been deleted from the database.
func (tf *TempFile) Deleted() bool {
	return tf._deleted
}

// Insert inserts the TempFile to the database.
func (tf *TempFile) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if tf._exists {
		return errors.New("insert failed: already exists")
	}

	// sql query
	const sqlstr = `INSERT INTO alpha.temp_files (` +
		`message_id, file_name, time_created` +
		`) VALUES (` +
		`?, ?, ?` +
		`)`

	// run query
	XOLog(sqlstr, tf.MessageID, tf.FileName, tf.TimeCreated)
	res, err := db.Exec(sqlstr, tf.MessageID, tf.FileName, tf.TimeCreated)
	if err != nil {
		return err
	}

	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set primary key and existence
	tf.ID = int(id)
	tf._exists = true

	return nil
}

// Update updates the TempFile in the database.
func (tf *TempFile) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !tf._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if tf._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE alpha.temp_files SET ` +
		`message_id = ?, file_name = ?, time_created = ?` +
		` WHERE id = ?`

	// run query
	XOLog(sqlstr, tf.MessageID, tf.FileName, tf.TimeCreated, tf.ID)
	_, err = db.Exec(sqlstr, tf.MessageID, tf.FileName, tf.TimeCreated, tf.ID)
	return err
}

// Save saves the TempFile to the database.
func (tf *TempFile) Save(db XODB) error {
	if tf.Exists() {
		return tf.Update(db)
	}

	return tf.Insert(db)
}

// Delete deletes the TempFile from the database.
func (tf *TempFile) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !tf._exists {
		return nil
	}

	// if deleted, bail
	if tf._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM alpha.temp_files WHERE id = ?`

	// run query
	XOLog(sqlstr, tf.ID)
	_, err = db.Exec(sqlstr, tf.ID)
	if err != nil {
		return err
	}

	// set deleted
	tf._deleted = true

	return nil
}

// TempFilesByMessageID retrieves a row from 'alpha.temp_files' as a TempFile.
//
// Generated from index 'message_id_idx'.
func TempFilesByMessageID(db XODB, messageID string) ([]*TempFile, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, message_id, file_name, time_created ` +
		`FROM alpha.temp_files ` +
		`WHERE message_id = ?`

	// run query
	XOLog(sqlstr, messageID)
	q, err := db.Query(sqlstr, messageID)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*TempFile{}
	for q.Next() {
		tf := TempFile{
			_exists: true,
		}

		// scan
		err = q.Scan(&tf.ID, &tf.MessageID, &tf.FileName, &tf.TimeCreated)
		if err != nil {
			return nil, err
		}

		res = append(res, &tf)
	}

	return res, nil
}

// TempFileByID retrieves a row from 'alpha.temp_files' as a TempFile.
//
// Generated from index 'temp_files_id_pkey'.
func TempFileByID(db XODB, id int) (*TempFile, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, message_id, file_name, time_created ` +
		`FROM alpha.temp_files ` +
		`WHERE id = ?`

	// run query
	XOLog(sqlstr, id)
	tf := TempFile{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, id).Scan(&tf.ID, &tf.MessageID, &tf.FileName, &tf.TimeCreated)
	if err != nil {
		return nil, err
	}

	return &tf, nil
}
