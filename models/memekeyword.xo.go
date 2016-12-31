// Package models contains the types for schema 'alpha'.
package models

// GENERATED BY XO. DO NOT EDIT.

import "errors"

// MemeKeyword represents a row from 'alpha.meme_keyword'.
type MemeKeyword struct {
	MemeID   int      `json:"meme_id"`   // meme_id
	Keyword  string   `json:"keyword"`   // keyword
	WordType WordType `json:"word_type"` // word_type
	Weight   int      `json:"weight"`    // weight

	// xo fields
	_exists, _deleted bool
}

// Exists determines if the MemeKeyword exists in the database.
func (mk *MemeKeyword) Exists() bool {
	return mk._exists
}

// Deleted provides information if the MemeKeyword has been deleted from the database.
func (mk *MemeKeyword) Deleted() bool {
	return mk._deleted
}

// Insert inserts the MemeKeyword to the database.
func (mk *MemeKeyword) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if mk._exists {
		return errors.New("insert failed: already exists")
	}

	// sql query
	const sqlstr = `INSERT INTO alpha.meme_keyword (` +
		`meme_id, word_type, weight` +
		`) VALUES (` +
		`?, ?, ?` +
		`)`

	// run query
	XOLog(sqlstr, mk.MemeID, mk.WordType, mk.Weight)
	res, err := db.Exec(sqlstr, mk.MemeID, mk.WordType, mk.Weight)
	if err != nil {
		return err
	}

	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set primary key and existence
	mk.Keyword = string(id)
	mk._exists = true

	return nil
}

// Update updates the MemeKeyword in the database.
func (mk *MemeKeyword) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !mk._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if mk._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE alpha.meme_keyword SET ` +
		`meme_id = ?, word_type = ?, weight = ?` +
		` WHERE keyword = ?`

	// run query
	XOLog(sqlstr, mk.MemeID, mk.WordType, mk.Weight, mk.Keyword)
	_, err = db.Exec(sqlstr, mk.MemeID, mk.WordType, mk.Weight, mk.Keyword)
	return err
}

// Save saves the MemeKeyword to the database.
func (mk *MemeKeyword) Save(db XODB) error {
	if mk.Exists() {
		return mk.Update(db)
	}

	return mk.Insert(db)
}

// Delete deletes the MemeKeyword from the database.
func (mk *MemeKeyword) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !mk._exists {
		return nil
	}

	// if deleted, bail
	if mk._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM alpha.meme_keyword WHERE keyword = ?`

	// run query
	XOLog(sqlstr, mk.Keyword)
	_, err = db.Exec(sqlstr, mk.Keyword)
	if err != nil {
		return err
	}

	// set deleted
	mk._deleted = true

	return nil
}

// Meme returns the Meme associated with the MemeKeyword's MemeID (meme_id).
//
// Generated from foreign key 'meme_keyword_ibfk_1'.
func (mk *MemeKeyword) Meme(db XODB) (*Meme, error) {
	return MemeByID(db, mk.MemeID)
}

// MemeKeywordsByKeyword retrieves a row from 'alpha.meme_keyword' as a MemeKeyword.
//
// Generated from index 'keyword_idx'.
func MemeKeywordsByKeyword(db XODB, keyword string) ([]*MemeKeyword, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`meme_id, keyword, word_type, weight ` +
		`FROM alpha.meme_keyword ` +
		`WHERE keyword = ?`

	// run query
	XOLog(sqlstr, keyword)
	q, err := db.Query(sqlstr, keyword)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*MemeKeyword{}
	for q.Next() {
		mk := MemeKeyword{
			_exists: true,
		}

		// scan
		err = q.Scan(&mk.MemeID, &mk.Keyword, &mk.WordType, &mk.Weight)
		if err != nil {
			return nil, err
		}

		res = append(res, &mk)
	}

	return res, nil
}

// MemeKeywordByKeyword retrieves a row from 'alpha.meme_keyword' as a MemeKeyword.
//
// Generated from index 'meme_keyword_keyword_pkey'.
func MemeKeywordByKeyword(db XODB, keyword string) (*MemeKeyword, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`meme_id, keyword, word_type, weight ` +
		`FROM alpha.meme_keyword ` +
		`WHERE keyword = ?`

	// run query
	XOLog(sqlstr, keyword)
	mk := MemeKeyword{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, keyword).Scan(&mk.MemeID, &mk.Keyword, &mk.WordType, &mk.Weight)
	if err != nil {
		return nil, err
	}

	return &mk, nil
}
