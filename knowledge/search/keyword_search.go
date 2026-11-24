package search

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
)

// KeywordSearch performs full-text search using PostgreSQL tsvector / BM25.
type KeywordSearch struct {
	db *sql.DB
}

// KeywordResult is a single match from the keyword index.
type KeywordResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Content    string  `json:"content"`
	Rank       float64 `json:"rank"` // BM25 or tsrank score
}

func NewKeywordSearch(db *sql.DB) *KeywordSearch {
	return &KeywordSearch{db: db}
}

// Search executes a full-text query against the chunks table.
func (ks *KeywordSearch) Search(query string, topK int, filters map[string]interface{}) []KeywordResult {
	tsquery := ks.NormaliseQuery(query)

	rows, err := ks.db.Query(
		`SELECT c.id, c.document_id, c.content,
		        ts_rank(c.search_vector, to_tsquery('english', $1)) AS rank
		 FROM chunks c
		 WHERE c.search_vector @@ to_tsquery('english', $1)
		 ORDER BY rank DESC
		 LIMIT $2`,
		tsquery, topK,
	)
	if err != nil {
		log.Printf("KeywordSearch: Full-text search failed: %v", err)
		return nil
	}
	defer rows.Close()

	var results []KeywordResult
	for rows.Next() {
		var r KeywordResult
		if err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.Content, &r.Rank); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results
}

// NormaliseQuery converts user input into a PostgreSQL tsquery-compatible string.
func (ks *KeywordSearch) NormaliseQuery(query string) string {
	re := regexp.MustCompile(`\w+`)
	words := re.FindAllString(query, -1)
	return strings.Join(words, " & ")
}