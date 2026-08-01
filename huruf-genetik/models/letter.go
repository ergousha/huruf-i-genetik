package models

// Letter represents a Quranic letter node in the Neo4j Graph.
type Letter struct {
	Char    string `json:"char"`
	Name    string `json:"name"`
	Alem    string `json:"alem"`
	Mertebe string `json:"mertebe"`
	Cins    string `json:"cins"`
}

// GeneticSequence represents the sequential flow of Alems in a given text.
type GeneticSequence struct {
	Blocks []string `json:"blocks"` // e.g., ["Ceberut", "Melekut", "Orta"]
}

// ResonanceProfile holds the calculated frequency of the text based on its letters.
type ResonanceProfile struct {
	TotalFrequency int            `json:"total_frequency"`
	AlemCounts     map[string]int `json:"alem_counts"`
}

// AnalysisResult is the aggregate output of a text analysis.
type AnalysisResult struct {
	OriginalText   string           `json:"original_text"`
	Letters        []Letter         `json:"letters"`
	Sequence       GeneticSequence  `json:"sequence"`
	Resonance      ResonanceProfile `json:"resonance"`
	GraphNodes     []GraphNode      `json:"nodes"`
	GraphEdges     []GraphEdge      `json:"edges"`
	BerzahCentrality map[string]int `json:"berzah_centrality,omitempty"`
}

// GraphNode is used for frontend visualization.
type GraphNode struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Group string `json:"group"` // Maps to Alem for color coding
	Title string `json:"title"` // Tooltip showing Mertebe, Name etc.
}

// GraphEdge is used for frontend visualization.
type GraphEdge struct {
	From   int    `json:"from"`
	To     int    `json:"to"`
	Label  string `json:"label"` // VASIL, FASIL, MED
	Weight int    `json:"weight"`
}

// SurahSignature represents the frequency of Alems in a specific Surah
type SurahSignature struct {
	Number  int     `json:"number"`
	Name    string  `json:"name"`
	Ceberut float64 `json:"ceberut"`
	Melekut float64 `json:"melekut"`
	Orta    float64 `json:"orta"`
	Asagi   float64 `json:"asagi"`
}

// EsmaCorrelation represents the similarity match between an Esma and a Surah
type EsmaCorrelation struct {
	SurahName  string  `json:"surah_name"`
	Similarity float64 `json:"similarity"` // 0.0 to 1.0 (or percentage)
}
