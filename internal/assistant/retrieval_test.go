package assistant

import "testing"

func TestRetrieveFAQ(t *testing.T) {
	tests := []struct {
		name      string
		question  []float32
		entries   []faqEntry
		wantText  string
		wantFound bool
		wantScore float64
	}{
		{
			name:     "selects closest FAQ above cutoff",
			question: []float32{0.8, 0.6},
			entries: []faqEntry{
				{text: "Overdraft fee", vector: []float32{1, 0}},
				{text: "Domestic transfer", vector: []float32{0, 1}},
			},
			wantText:  "Overdraft fee",
			wantFound: true,
			wantScore: 0.8,
		},
		{
			name:      "rejects closest FAQ below cutoff",
			question:  []float32{0.5, 0.8660254},
			entries:   []faqEntry{{text: "Domestic transfer", vector: []float32{1, 0}}},
			wantText:  "Domestic transfer",
			wantFound: false,
			wantScore: 0.5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry, score, found := retrieveFAQ(test.question, test.entries)
			if entry.text != test.wantText {
				t.Errorf("selected FAQ = %q, want %q", entry.text, test.wantText)
			}
			if found != test.wantFound {
				t.Errorf("found = %t, want %t", found, test.wantFound)
			}
			if difference := score - test.wantScore; difference < -0.001 || difference > 0.001 {
				t.Errorf("score = %.3f, want %.3f", score, test.wantScore)
			}
		})
	}
}
