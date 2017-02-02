package joinmodels

import (
	"github.com/harrisonzhao/supermeme/models"
)

type RandomMemeIdResult struct {
	ID int
}

func RandomMemeId(db models.XODB) (int, error) {
	var err error

	// sql query
	const sqlstr = `SELECT FLOOR(RAND() * MAX(id)) id FROM alpha.meme`

	// run query
	models.XOLog(sqlstr)
	res := RandomMemeIdResult{}

	err = db.QueryRow(sqlstr).Scan(&res.ID)
	if err != nil {
		return 0, err
	}

	return res.ID, nil
}
