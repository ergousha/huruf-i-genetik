package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"huruf-genetik/models"
	"huruf-genetik/repository"
)

type EngineService struct {
	repo     *repository.Neo4jRepository
	analyzer *AnalyzerService
}

func NewEngineService(repo *repository.Neo4jRepository, analyzer *AnalyzerService) *EngineService {
	return &EngineService{
		repo:     repo,
		analyzer: analyzer,
	}
}

// alemWeights maps Alems to their cosmological frequency weights.
var alemWeights = map[string]int{
	"Ceberut":       3,
	"Melekut":       2,
	"Orta":          1,
	"Asagi":         0,
	"GirisCikisSiz": 0, // Fallback/default for Elif if needed
}

// ProcessText takes an input string, parses it, updates the graph, and returns the analysis.
func (e *EngineService) ProcessText(ctx context.Context, text string) (*models.AnalysisResult, error) {
	chars := e.analyzer.CleanText(text)
	if len(chars) == 0 {
		return nil, fmt.Errorf("no valid arabic characters found in input")
	}

	result := &models.AnalysisResult{
		OriginalText: text,
		Resonance: models.ResonanceProfile{
			AlemCounts: make(map[string]int),
		},
	}

	// 1. Fetch letter properties from DB & Calculate Resonance
	for _, ch := range chars {
		letter, err := e.repo.GetLetter(ctx, ch)
		if err != nil {
			log.Printf("Error fetching letter %s: %v", ch, err)
			continue
		}
		if letter != nil {
			result.Letters = append(result.Letters, *letter)
			
			// Sequence map
			if len(result.Sequence.Blocks) == 0 || result.Sequence.Blocks[len(result.Sequence.Blocks)-1] != letter.Alem {
				result.Sequence.Blocks = append(result.Sequence.Blocks, letter.Alem)
			}

			// Resonance profiling
			result.Resonance.AlemCounts[letter.Alem]++
			result.Resonance.TotalFrequency += alemWeights[letter.Alem]
		}
	}

	// 2. Parse Relations and Update Database Graph
	relations := e.analyzer.ParseRelations(chars)
	for _, rel := range relations {
		err := e.repo.SaveTextRelation(ctx, rel.From, rel.To, rel.Type)
		if err != nil {
			log.Printf("Error saving relation %s->%s (%s): %v", rel.From, rel.To, rel.Type, err)
		}
	}

	// 3. Build Graph representation for Frontend
	nodeMap := make(map[string]int)
	nodeID := 1
	for _, l := range result.Letters {
		if _, exists := nodeMap[l.Char]; !exists {
			nodeMap[l.Char] = nodeID
			result.GraphNodes = append(result.GraphNodes, models.GraphNode{
				ID:    nodeID,
				Label: l.Char,
				Group: l.Alem,
				Title: fmt.Sprintf("%s (%s) - %s", l.Name, l.Alem, l.Mertebe),
			})
			nodeID++
		}
	}

	for _, rel := range relations {
		if fromID, ok1 := nodeMap[rel.From]; ok1 {
			if toID, ok2 := nodeMap[rel.To]; ok2 {
				result.GraphEdges = append(result.GraphEdges, models.GraphEdge{
					From:   fromID,
					To:     toID,
					Label:  rel.Type,
					Weight: 1, // Visual weight
				})
			}
		}
	}

	// 4. Calculate Berzah Centrality
	berzah, err := e.repo.CalculateBerzah(ctx)
	if err == nil {
		result.BerzahCentrality = berzah
	} else {
		log.Printf("Error calculating berzah: %v", err)
	}

	return result, nil
}

// GetDashboardMetrics fetches macroscopic correlations from the database.
func (e *EngineService) GetDashboardMetrics(ctx context.Context) (*repository.MacroCorrelations, error) {
	return e.repo.GetMacroCorrelations(ctx)
}

// ProcessQuranImport fetches the entire Quran and saves relations in a batch.
func (e *EngineService) ProcessQuranImport(ctx context.Context) error {
	log.Println("Starting Quran Import Process...")
	importer := NewQuranImporter()

	surahs, err := importer.FetchAllSurahs()
	if err != nil {
		return err
	}

	var allRelations []map[string]interface{}
	var signatures []models.SurahSignature
	totalAyahs := 0

	for _, surah := range surahs {
		var cCount, mCount, oCount, aCount float64
		totalLetters := 0.0

		for _, ayah := range surah.Ayahs {
			totalAyahs++
			chars := e.analyzer.CleanText(ayah)
			
			// Calculate frequencies for the Surah signature
			for _, ch := range chars {
				l, _ := e.repo.GetLetter(ctx, ch)
				if l != nil {
					totalLetters++
					switch l.Alem {
					case "Ceberut":
						cCount++
					case "Melekut":
						mCount++
					case "Orta":
						oCount++
					case "Asagi":
						aCount++
					}
				}
			}

			// Parse relations for the global graph
			relations := e.analyzer.ParseRelations(chars)
			for _, rel := range relations {
				allRelations = append(allRelations, map[string]interface{}{
					"from":   rel.From,
					"to":     rel.To,
					"type":   rel.Type,
					"weight": 1,
				})
			}
		}

		if totalLetters > 0 {
			signatures = append(signatures, models.SurahSignature{
				Number:  surah.Number,
				Name:    surah.Name,
				Ceberut: cCount / totalLetters,
				Melekut: mCount / totalLetters,
				Orta:    oCount / totalLetters,
				Asagi:   aCount / totalLetters,
			})
		}
	}

	log.Printf("Fetched %d ayahs across %d surahs. Processing relations...\n", totalAyahs, len(surahs))
	log.Printf("Parsed %d global relations. Saving to Neo4j in batch...\n", len(allRelations))

	err = e.repo.SaveTextRelationsBatch(ctx, allRelations)
	if err != nil {
		return err
	}

	log.Println("Relations saved. Now saving Surah signatures...")
	err = e.repo.SaveSurahSignatures(ctx, signatures)
	if err != nil {
		return err
	}
	
	log.Println("Quran Import Successfully Completed!")
	return nil
}

func (e *EngineService) GetAlemTransitionMatrix(ctx context.Context) ([]repository.Transition, error) {
	return e.repo.GetAlemTransitionMatrix(ctx)
}

func (e *EngineService) GetRootWordJourney(ctx context.Context, text string) (*repository.RootJourney, error) {
	chars := e.analyzer.CleanText(text)
	if len(chars) == 0 {
		return nil, fmt.Errorf("no valid arabic characters found in input")
	}
	return e.repo.GetRootWordJourney(ctx, chars)
}

func (e *EngineService) CalculateEsmaCorrelation(ctx context.Context, text string) ([]models.EsmaCorrelation, error) {
	chars := e.analyzer.CleanText(text)
	if len(chars) == 0 {
		return nil, fmt.Errorf("no valid arabic characters found in input")
	}

	var cCount, mCount, oCount, aCount float64
	totalLetters := 0.0

	// Calculate Esma vector
	for _, ch := range chars {
		l, _ := e.repo.GetLetter(ctx, ch)
		if l != nil {
			totalLetters++
			switch l.Alem {
			case "Ceberut":
				cCount++
			case "Melekut":
				mCount++
			case "Orta":
				oCount++
			case "Asagi":
				aCount++
			}
		}
	}

	if totalLetters == 0 {
		return nil, fmt.Errorf("could not map input to any known Alems")
	}

	eV := []float64{
		cCount / totalLetters,
		mCount / totalLetters,
		oCount / totalLetters,
		aCount / totalLetters,
	}

	surahs, err := e.repo.GetAllSurahSignatures(ctx)
	if err != nil {
		return nil, err
	}

	// 1. Calculate Mean and StdDev for each dimension across all Surahs
	var means, stdDevs [4]float64
	n := float64(len(surahs))

	for _, s := range surahs {
		means[0] += s.Ceberut
		means[1] += s.Melekut
		means[2] += s.Orta
		means[3] += s.Asagi
	}
	for i := 0; i < 4; i++ {
		means[i] /= n
	}

	for _, s := range surahs {
		stdDevs[0] += (s.Ceberut - means[0]) * (s.Ceberut - means[0])
		stdDevs[1] += (s.Melekut - means[1]) * (s.Melekut - means[1])
		stdDevs[2] += (s.Orta - means[2]) * (s.Orta - means[2])
		stdDevs[3] += (s.Asagi - means[3]) * (s.Asagi - means[3])
	}
	for i := 0; i < 4; i++ {
		stdDevs[i] = math.Sqrt(stdDevs[i] / n)
		// Prevent division by zero
		if stdDevs[i] == 0 {
			stdDevs[i] = 1.0
		}
	}

	// 2. Normalize the Esma Vector (Z-Score)
	zEV := [4]float64{}
	for i := 0; i < 4; i++ {
		zEV[i] = (eV[i] - means[i]) / stdDevs[i]
	}

	var correlations []models.EsmaCorrelation
	var maxDist, minDist float64
	minDist = math.MaxFloat64
	var distances []float64

	for _, s := range surahs {
		// Normalize Surah Vector (Z-Score)
		zSV := [4]float64{
			(s.Ceberut - means[0]) / stdDevs[0],
			(s.Melekut - means[1]) / stdDevs[1],
			(s.Orta - means[2]) / stdDevs[2],
			(s.Asagi - means[3]) / stdDevs[3],
		}

		// Calculate Euclidean Distance on Z-Scores
		var sumSq float64
		for i := 0; i < 4; i++ {
			diff := zEV[i] - zSV[i]
			sumSq += diff * diff
		}
		
		dist := math.Sqrt(sumSq)
		distances = append(distances, dist)

		if dist > maxDist {
			maxDist = dist
		}
		if dist < minDist {
			minDist = dist
		}
	}

	// 3. Convert Distances to highly spread Percentages (Min-Max scaling over the results)
	// The smallest distance becomes 100%, the largest becomes 1%
	for idx, s := range surahs {
		dist := distances[idx]
		
		// To create exponential separation, we can square the normalized score
		// First, invert the distance so closer is higher (0 to 1)
		normalizedInverted := 1.0
		if maxDist != minDist {
			normalizedInverted = 1.0 - ((dist - minDist) / (maxDist - minDist))
		}
		
		// Exponential boost for top matches (makes differences huge)
		finalScore := math.Pow(normalizedInverted, 3) * 100.0

		// Ensure it never shows 0 unless it's exactly the worst
		if finalScore < 0.1 {
			finalScore = 0.1
		}

		correlations = append(correlations, models.EsmaCorrelation{
			SurahName:  fmt.Sprintf("%d. %s", s.Number, s.Name),
			Similarity: finalScore,
		})
	}

	// Sort by highest similarity
	sort.Slice(correlations, func(i, j int) bool {
		return correlations[i].Similarity > correlations[j].Similarity
	})

	// Return Top 10
	if len(correlations) > 10 {
		correlations = correlations[:10]
	}

	return correlations, nil
}

// 99 Names of Allah (Esma-ül Hüsna) without "El-" prefix, with Turkish transliteration
var EsmaulHusna = []string{
	"الله (Allah)", "رحمن (Rahman)", "رحيم (Rahim)", "ملك (Melik)", "قدوس (Kuddüs)", "سلام (Selam)", "مؤمن (Mü'min)", "مهيمن (Müheymin)", "عزيز (Aziz)", "جبار (Cebbar)",
	"متكبر (Mütekebbir)", "خالق (Halik)", "بارئ (Bari)", "مصور (Musavvir)", "غفار (Gaffar)", "قهار (Kahhar)", "وهاب (Vehhab)", "رزاق (Rezzak)", "فتاح (Fettah)", "عليم (Alim)",
	"قابض (Kabıd)", "باسط (Basıt)", "خافض (Hafıd)", "رافع (Rafi)", "معز (Mu'ız)", "مذل (Müzil)", "سميع (Semi)", "بصير (Basir)", "حكم (Hakem)", "عدل (Adl)",
	"لطيف (Latif)", "خبير (Habir)", "حليم (Halim)", "عظيم (Azim)", "غفور (Gafur)", "شكور (Şekur)", "علي (Aliyy)", "كبير (Kebir)", "حفيظ (Hafiz)", "مقيت (Mukit)",
	"حسيب (Hasib)", "جليل (Celil)", "كريم (Kerim)", "رقيب (Rakib)", "مجيب (Mucib)", "واسع (Vasi)", "حكيم (Hakim)", "ودود (Vedud)", "مجيد (Mecid)", "باعث (Bais)",
	"شهيد (Şehid)", "حق (Hakk)", "وكيل (Vekil)", "قوي (Kaviyy)", "متين (Metin)", "ولي (Veliyy)", "حميد (Hamid)", "محصي (Muhsi)", "مبدئ (Mübdi)", "معيد (Muid)",
	"محيي (Muhyi)", "مميت (Mümit)", "حي (Hayy)", "قيوم (Kayyum)", "واجد (Vacid)", "ماجد (Macid)", "واحد (Vahid)", "صمد (Samed)", "قادر (Kadir)", "مقتدر (Muktedir)",
	"مقدم (Mukaddim)", "مؤخر (Muahhir)", "أول (Evvel)", "آخر (Ahir)", "ظاهر (Zahir)", "باطن (Batın)", "والي (Vali)", "متعالي (Müteali)", "بر (Berr)", "تواب (Tevvab)",
	"منتقم (Müntakim)", "عفو (Afüvv)", "رؤوف (Rauf)", "مالك الملك (Malik-ül Mülk)", "ذو الجلال والإكرام (Zül-Celali vel-İkram)", "مقسط (Muksit)", "جامع (Cami')", "غني (Ganiyy)", "مغني (Muğni)", "مانع (Mani')",
	"ضار (Darr)", "نافع (Nafi')", "نور (Nur)", "هادي (Hadi)", "بديع (Bedi')", "باقي (Baki)", "وارث (Varis)", "رشيد (Reşid)", "صبور (Sabur)",
}

type MatrixRow struct {
	Esma   string    `json:"esma"`
	Scores []float64 `json:"scores"` // Scores ordered by Surah number 1-114
}

type EsmaMatrixResponse struct {
	SurahNames []string    `json:"surah_names"`
	Matrix     []MatrixRow `json:"matrix"`
}

func (e *EngineService) CalculateEsmaMatrix(ctx context.Context) (*EsmaMatrixResponse, error) {
	surahs, err := e.repo.GetAllSurahSignatures(ctx)
	if err != nil {
		return nil, err
	}

	// 1. Calculate Mean and StdDev for each dimension across all Surahs
	var means, stdDevs [4]float64
	n := float64(len(surahs))

	for _, s := range surahs {
		means[0] += s.Ceberut
		means[1] += s.Melekut
		means[2] += s.Orta
		means[3] += s.Asagi
	}
	for i := 0; i < 4; i++ {
		means[i] /= n
	}

	for _, s := range surahs {
		stdDevs[0] += (s.Ceberut - means[0]) * (s.Ceberut - means[0])
		stdDevs[1] += (s.Melekut - means[1]) * (s.Melekut - means[1])
		stdDevs[2] += (s.Orta - means[2]) * (s.Orta - means[2])
		stdDevs[3] += (s.Asagi - means[3]) * (s.Asagi - means[3])
	}
	for i := 0; i < 4; i++ {
		stdDevs[i] = math.Sqrt(stdDevs[i] / n)
		if stdDevs[i] == 0 {
			stdDevs[i] = 1.0
		}
	}

	// Prepare Surah names and pre-calculate Z-Scores for all 114 Surahs
	var surahNames []string
	var surahZScores [][4]float64
	// Sort surahs by number to ensure 1-114 ordering
	sort.Slice(surahs, func(i, j int) bool {
		return surahs[i].Number < surahs[j].Number
	})

	for _, s := range surahs {
		surahNames = append(surahNames, fmt.Sprintf("%d. %s", s.Number, s.Name))
		surahZScores = append(surahZScores, [4]float64{
			(s.Ceberut - means[0]) / stdDevs[0],
			(s.Melekut - means[1]) / stdDevs[1],
			(s.Orta - means[2]) / stdDevs[2],
			(s.Asagi - means[3]) / stdDevs[3],
		})
	}

	response := &EsmaMatrixResponse{
		SurahNames: surahNames,
	}

	// Find global Min and Max distances to normalize across the ENTIRE matrix
	// This ensures colors are relative to all 99 Esmas and 114 Surahs.
	var rawDistances [][]float64
	var globalMaxDist, globalMinDist float64
	globalMinDist = math.MaxFloat64

	for _, esmaName := range EsmaulHusna {
		chars := e.analyzer.CleanText(esmaName)
		var cCount, mCount, oCount, aCount float64
		totalLetters := 0.0

		for _, ch := range chars {
			l, _ := e.repo.GetLetter(ctx, ch)
			if l != nil {
				totalLetters++
				switch l.Alem {
				case "Ceberut":
					cCount++
				case "Melekut":
					mCount++
				case "Orta":
					oCount++
				case "Asagi":
					aCount++
				}
			}
		}

		if totalLetters == 0 {
			rawDistances = append(rawDistances, make([]float64, 114))
			continue
		}

		eV := [4]float64{
			cCount / totalLetters,
			mCount / totalLetters,
			oCount / totalLetters,
			aCount / totalLetters,
		}

		zEV := [4]float64{}
		for i := 0; i < 4; i++ {
			zEV[i] = (eV[i] - means[i]) / stdDevs[i]
		}

		var esmaDistances []float64
		for _, zSV := range surahZScores {
			var sumSq float64
			for i := 0; i < 4; i++ {
				diff := zEV[i] - zSV[i]
				sumSq += diff * diff
			}
			dist := math.Sqrt(sumSq)
			esmaDistances = append(esmaDistances, dist)

			if dist > globalMaxDist {
				globalMaxDist = dist
			}
			if dist < globalMinDist {
				globalMinDist = dist
			}
		}
		rawDistances = append(rawDistances, esmaDistances)
	}

	// Convert all raw distances to normalized scores
	for i, esmaName := range EsmaulHusna {
		row := MatrixRow{Esma: esmaName}
		
		for _, dist := range rawDistances[i] {
			if dist == 0 && totalLettersCheck(esmaName) == 0 { // Placeholder check
				row.Scores = append(row.Scores, 0)
				continue
			}

			normalizedInverted := 1.0
			if globalMaxDist != globalMinDist {
				normalizedInverted = 1.0 - ((dist - globalMinDist) / (globalMaxDist - globalMinDist))
			}
			
			finalScore := math.Pow(normalizedInverted, 3) * 100.0
			if finalScore < 0.1 {
				finalScore = 0.1
			}
			
			// Round to 2 decimal places for cleaner JSON
			finalScore = math.Round(finalScore*100) / 100
			row.Scores = append(row.Scores, finalScore)
		}
		response.Matrix = append(response.Matrix, row)
	}

	return response, nil
}

// Simple helper to check if Esma is empty
func totalLettersCheck(text string) int {
	return len(strings.TrimSpace(text))
}
