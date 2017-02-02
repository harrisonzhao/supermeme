package joinmodels

import (
	"github.com/harrisonzhao/supermeme/models"
)

func RandomMeme(db models.XODB) (*models.Meme, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, source, url, top_text, bottom_text, net_ups, views, num_keywords, meme_name, imgur_bg_image ` +
		`FROM alpha.meme ` +
		`WHERE id >= FLOOR(RAND() * MAX(id)) LIMIT 1`

	// run query
	models.XOLog(sqlstr)
	m := models.Meme{}

	err = db.QueryRow(sqlstr).Scan(&m.ID, &m.Source, &m.URL, &m.TopText, &m.BottomText, &m.NetUps, &m.Views, &m.NumKeywords, &m.MemeName, &m.ImgurBgImage)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
