package join_models

import (
	"github.com/harrisonzhao/supermeme/models"
	"github.com/Masterminds/squirrel"
)

type BestMemeResult struct {
	ID    int
	Score float64
}

func BestMemeResultsByKeywords(db models.XODB, keywords []string) ([]*BestMemeResult, error) {
	if len(keywords) == 0 {
		return nil, nil
	}
	sql, args, err := squirrel.Select(
		"m.id",
		"COUNT(mk.keyword) / COALESCE(NULLIF(m.num_keywords, 0), 1000) * LOG(GREATEST(m.net_ups, 1)) as score").
		From("meme m").
		Join("meme_keyword mk ON mk.meme_id = m.id").
		Where("mk.keyword IN (" + squirrel.Placeholders(len(keywords)) +")", keywords).
		GroupBy("m.id").
		OrderBy("score DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sql, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// load results
	res := []*BestMemeResult{}
	for rows.Next() {
		bmr := BestMemeResult{}

		// scan
		err = rows.Scan(&bmr.ID, &bmr.Score)
		if err != nil {
			return nil, err
		}

		res = append(res, &bmr)
	}

	return res, nil
}