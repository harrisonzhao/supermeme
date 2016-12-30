package joinmodels

import (
	"github.com/Masterminds/squirrel"
	"github.com/harrisonzhao/supermeme/models"
)

type BestMemeResult struct {
	ID    int
	Score float64
}

func BestMemeResultsByKeywords(db models.XODB, keywords []string) (*BestMemeResult, error) {
	if len(keywords) == 0 {
		return nil, nil
	}
	keywordArgs := make([]interface{}, len(keywords))
	for i := range keywords {
		keywordArgs[i] = keywords[i]
	}
	sql, args, err := squirrel.Select(
		"m.id",
		"COUNT(mk.keyword) / COALESCE(NULLIF(m.num_keywords, 0), 1000) * LOG(GREATEST(m.net_ups, 1)) as score").
		From("meme m").
		Join("meme_keyword mk ON mk.meme_id = m.id").
		Where("mk.keyword IN ("+squirrel.Placeholders(len(keywords))+")", keywordArgs...).
		GroupBy("m.id").
		OrderBy("score DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// load results
	var res *BestMemeResult = nil
	for rows.Next() {
		bmr := BestMemeResult{}

		// scan
		err = rows.Scan(&bmr.ID, &bmr.Score)
		if err != nil {
			return nil, err
		}

		res = &bmr
	}

	return res, nil
}
