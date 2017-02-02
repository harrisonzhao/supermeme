package joinmodels

import (
	"math/rand"

	"github.com/Masterminds/squirrel"
	"github.com/harrisonzhao/supermeme/models"
)

var randomGenerator *rand.Rand = rand.New(rand.NewSource(9724597496))

type BestMemeResult struct {
	ID    int
	Score float64
}

func BestMemeResultsByKeywords(db models.XODB, keywords []string, limit int) (*BestMemeResult, error) {
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
		Limit(uint64(limit)).
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
	results := make([]*BestMemeResult, 0)
	for rows.Next() {
		bmr := BestMemeResult{}

		// scan
		err = rows.Scan(&bmr.ID, &bmr.Score)
		if err != nil {
			return nil, err
		}

		results = append(results, &bmr)
	}
	if len(results) == 0 {
		return nil, err
	}
	randIndex := randomGenerator.Intn(len(results))

	return results[randIndex], nil
}
