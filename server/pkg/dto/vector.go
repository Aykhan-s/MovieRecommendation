package dto

import (
	"fmt"
	"math"
)

type CountVectorizer struct {
	WordIndex map[string]int
}

func NewCountVectorizer() *CountVectorizer {
	return &CountVectorizer{}
}

func (cv *CountVectorizer) SetWordIndexes(docs [][]string) {
	cv.WordIndex = make(map[string]int)
	index := 0
	for _, doc := range docs {
		for _, word := range doc {
			if word == "" {
				continue
			}
			if _, exists := cv.WordIndex[word]; !exists {
				cv.WordIndex[word] = index
				index++
			}
		}
	}
}

func (cv *CountVectorizer) Vectorize(doc []string) []uint8 {
	vector := make([]uint8, len(cv.WordIndex))
	for _, word := range doc {
		vector[cv.WordIndex[word]]++
	}
	return vector
}

func CosineSimilarity(a, b []uint8) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("slices must have the same length")
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		x := float64(a[i])
		y := float64(b[i])
		dotProduct += x * y
		normA += x * x
		normB += y * y
	}

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))), nil
}
