package assistant

import "math"

// This starting cutoff keeps the current paraphrase example (0.652) and rejects
// the unrelated closing-time question (0.537). Revisit it as we collect examples.
const minimumFAQSimilarity = 0.60

type faqEntry struct {
	text   string
	vector []float32
}

// retrieveFAQ returns the FAQ entry whose vector is most similar to the question.
func retrieveFAQ(questionVector []float32, entries []faqEntry) (faqEntry, float64, bool) {
	bestEntry := faqEntry{}
	bestScore := -1.0

	for _, entry := range entries {
		score := cosineSimilarity(questionVector, entry.vector)
		if score > bestScore {
			bestScore = score
			bestEntry = entry
		}
	}

	return bestEntry, bestScore, bestScore >= minimumFAQSimilarity
}

// cosineSimilarity measures how closely two vectors point in the same direction.
func cosineSimilarity(left, right []float32) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return -1
	}

	var dotProduct, leftLength, rightLength float64
	for i := range left {
		dotProduct += float64(left[i]) * float64(right[i])
		leftLength += float64(left[i]) * float64(left[i])
		rightLength += float64(right[i]) * float64(right[i])
	}
	if leftLength == 0 || rightLength == 0 {
		return -1
	}

	return dotProduct / (math.Sqrt(leftLength) * math.Sqrt(rightLength))
}
