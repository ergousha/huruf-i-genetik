package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"huruf-genetik/config"
	"huruf-genetik/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jRepository struct {
	db *config.Database
}

func NewNeo4jRepository(db *config.Database) *Neo4jRepository {
	return &Neo4jRepository{db: db}
}

const seedData = `[
  {"char": "ا", "name": "Elif", "alem": "GirisCikisSiz", "mertebe": "Havas,HavasulHavas,OzununOzu,YaratiklarinIlistigi", "cins": "Tekil"},
  {"char": "ء", "name": "Hemze", "alem": "Ceberut", "mertebe": "Belirtilmemis", "cins": "Belirtilmemis"},
  {"char": "ب", "name": "Ba", "alem": "Asagi", "mertebe": "EnSeckin,HavasulHavas,OzununOzu,FirikEhli", "cins": "Uclu"},
  {"char": "ت", "name": "Ta", "alem": "Orta", "mertebe": "HavasulHavas,NurlarEhli", "cins": "Uclu"},
  {"char": "ث", "name": "Se", "alem": "Orta", "mertebe": "OzununOzu,NurlarEhli", "cins": "Dortlu"},
  {"char": "ج", "name": "Cim", "alem": "Orta", "mertebe": "Avam,Gonderilen,FirikEhli,DuasiKarisik", "cins": "Uclu"},
  {"char": "ح", "name": "Ha", "alem": "Melekut", "mertebe": "Avam,Havas,HavasulHavas,Gonderilen", "cins": "Uclu"},
  {"char": "خ", "name": "Hı", "alem": "Melekut", "mertebe": "Gonderilen", "cins": "Uclu"},
  {"char": "د", "name": "Dal", "alem": "Orta", "mertebe": "Avam,OzununOzu,YaratiklarinIlistigi", "cins": "Ikili"},
  {"char": "ذ", "name": "Zel", "alem": "Orta", "mertebe": "YaratiklarinIlistigi,NurlarEhli", "cins": "Ikili"},
  {"char": "ر", "name": "Re", "alem": "Orta", "mertebe": "Havas,OzununOzu,YaratiklarinIlistigi", "cins": "Belirtilmemis"},
  {"char": "ز", "name": "Ze", "alem": "Orta", "mertebe": "OzununOzu,YaratiklarinIlistigi,NurlarEhli,DuasiKarisik", "cins": "Belirtilmemis"},
  {"char": "س", "name": "Sin", "alem": "Orta", "mertebe": "Havas,HavasulHavas,OzununOzu,NurlarEhli", "cins": "Belirtilmemis"},
  {"char": "ش", "name": "Şın", "alem": "Orta", "mertebe": "NurlarEhli", "cins": "Belirtilmemis"},
  {"char": "ص", "name": "Sad", "alem": "Orta", "mertebe": "Havas,HavasulHavas", "cins": "Ikili"},
  {"char": "ض", "name": "Dat", "alem": "Orta", "mertebe": "Avam,NurlarEhli", "cins": "Belirtilmemis"},
  {"char": "ط", "name": "Tı", "alem": "Orta", "mertebe": "Havas,HavasulHavas,OzununOzu,NurlarEhli", "cins": "Belirtilmemis"},
  {"char": "ظ", "name": "Zı", "alem": "Orta", "mertebe": "OzununOzu,NurlarEhli", "cins": "Belirtilmemis"},
  {"char": "ع", "name": "Ayın", "alem": "Melekut", "mertebe": "Havas,Yuksek", "cins": "Belirtilmemis"},
  {"char": "غ", "name": "Gayın", "alem": "Melekut", "mertebe": "Avam,Havas,HavasulHavas,NurlarEhli", "cins": "Belirtilmemis"},
  {"char": "ف", "name": "Fe", "alem": "Orta", "mertebe": "OzununOzu,NurlarEhli,DuasiKarisik", "cins": "Belirtilmemis"},
  {"char": "ق", "name": "Kaf", "alem": "Orta", "mertebe": "Havas,HavasulHavas,DuasiKarisik", "cins": "Belirtilmemis"},
  {"char": "ك", "name": "Kef", "alem": "Orta", "mertebe": "Havas,HavasulHavas,Gonderilen,Yuksek", "cins": "Tekil"},
  {"char": "ل", "name": "Lam", "alem": "Orta", "mertebe": "Havas,HavasulHavas,OzununOzu,DuasiKarisik", "cins": "Tekil"},
  {"char": "م", "name": "Mim", "alem": "Asagi", "mertebe": "Havas,OzununOzu,Yuksek", "cins": "Tekil"},
  {"char": "ن", "name": "Nun", "alem": "Orta", "mertebe": "Havas,HavasulHavas,OzununOzu,NurlarEhli,Dortlu_Ortada", "cins": "Tekil"},
  {"char": "ه", "name": "He", "alem": "Ceberut", "mertebe": "Havas,OzununOzu", "cins": "Tekil"},
  {"char": "و", "name": "Vav", "alem": "Asagi", "mertebe": "Havas,HavasulHavas,OzununOzu,YaratiklarinIlistigi", "cins": "Tekil"},
  {"char": "ي", "name": "Ya", "alem": "Orta", "mertebe": "Havas,HavasulHavas,OzununOzu,DuasiKarisik,Dortlu_Ortada", "cins": "Belirtilmemis"}
]`

// Seed ensures that all 29 letters are present in the database.
func (repo *Neo4jRepository) Seed(ctx context.Context) error {
	var letters []models.Letter
	err := json.Unmarshal([]byte(seedData), &letters)
	if err != nil {
		return fmt.Errorf("failed to parse seed data: %w", err)
	}

	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// Transaction 1: Schema creation (cannot be mixed with data writes in the same transaction in Neo4j)
	_, _ = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "CREATE CONSTRAINT letter_char IF NOT EXISTS FOR (l:Letter) REQUIRE l.char IS UNIQUE", nil)
		if err != nil {
			log.Printf("Warning: Failed to create constraint: %v\n", err)
		}
		return nil, nil
	})

	// Transaction 2: Data seeding
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			UNWIND $letters AS l
			MERGE (node:Letter {char: l.char})
			SET node.name = l.name,
			    node.alem = l.alem,
			    node.mertebe = l.mertebe,
			    node.cins = l.cins
		`

		var letterMaps []map[string]interface{}
		for _, l := range letters {
			letterMaps = append(letterMaps, map[string]interface{}{
				"char":    l.Char,
				"name":    l.Name,
				"alem":    l.Alem,
				"mertebe": l.Mertebe,
				"cins":    l.Cins,
			})
		}

		result, err := tx.Run(ctx, query, map[string]interface{}{"letters": letterMaps})
		if err != nil {
			return nil, err
		}
		summary, _ := result.Consume(ctx)
		log.Printf("Seed completed. Nodes created: %d, Properties set: %d\n", summary.Counters().NodesCreated(), summary.Counters().PropertiesSet())
		return nil, nil
	})

	return err
}

// GetLetter retrieves a letter node from the database.
func (repo *Neo4jRepository) GetLetter(ctx context.Context, char string) (*models.Letter, error) {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `MATCH (l:Letter {char: $char}) RETURN l.char, l.name, l.alem, l.mertebe, l.cins`
		result, err := tx.Run(ctx, query, map[string]interface{}{"char": char})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			char, _ := record.Get("l.char")
			name, _ := record.Get("l.name")
			alem, _ := record.Get("l.alem")
			mertebe, _ := record.Get("l.mertebe")
			cins, _ := record.Get("l.cins")

			return &models.Letter{
				Char:    char.(string),
				Name:    name.(string),
				Alem:    alem.(string),
				Mertebe: mertebe.(string),
				Cins:    cins.(string),
			}, nil
		}
		return nil, nil
	})

	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.(*models.Letter), nil
}

// SaveTextRelation creates or updates relations (VASIL, FASIL, MED) between two characters.
func (repo *Neo4jRepository) SaveTextRelation(ctx context.Context, char1, char2, relType string) error {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Using dynamic relationship type requires slightly different cypher or APOC.
		// Since relType is one of VASIL, FASIL, MED we can construct the query string safely.
		if relType != "VASIL" && relType != "FASIL" && relType != "MED" {
			return nil, fmt.Errorf("invalid relation type: %s", relType)
		}

		query := fmt.Sprintf(`
			MATCH (a:Letter {char: $char1}), (b:Letter {char: $char2})
			MERGE (a)-[r:%s]->(b)
			ON CREATE SET r.weight = 1
			ON MATCH SET r.weight = coalesce(r.weight, 0) + 1
		`, relType)

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"char1": char1,
			"char2": char2,
		})
		return nil, err
	})

	return err
}

// CalculateBerzah identifies the most prominent letters acting as bridges (Betweenness-like).
// It returns a map of Character -> Number of connected paths.
func (repo *Neo4jRepository) CalculateBerzah(ctx context.Context) (map[string]int, error) {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Simplified degree centrality for "Berzah" concept as a starting point.
		// Finds letters that connect different Alems most frequently.
		query := `
			MATCH (a:Letter)-[r]->(b:Letter)
			WHERE a.alem <> b.alem
			RETURN a.char AS char, sum(r.weight) AS score
			ORDER BY score DESC LIMIT 10
		`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		berzahMap := make(map[string]int)
		for result.Next(ctx) {
			record := result.Record()
			char, _ := record.Get("char")
			score, _ := record.Get("score")
			berzahMap[char.(string)] = int(score.(int64))
		}
		return berzahMap, nil
	})

	if err != nil {
		return nil, err
	}
	return res.(map[string]int), nil
}

// SaveTextRelationsBatch saves multiple relations efficiently using UNWIND.
// Expected map keys for relation: from, to, type, weight
func (repo *Neo4jRepository) SaveTextRelationsBatch(ctx context.Context, relations []map[string]interface{}) error {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// We separate by relation type since dynamic relationship type in MERGE is tricky without APOC.
	vasilRels := make([]map[string]interface{}, 0)
	fasilRels := make([]map[string]interface{}, 0)
	medRels := make([]map[string]interface{}, 0)

	for _, r := range relations {
		t := r["type"].(string)
		if t == "VASIL" {
			vasilRels = append(vasilRels, r)
		} else if t == "FASIL" {
			fasilRels = append(fasilRels, r)
		} else if t == "MED" {
			medRels = append(medRels, r)
		}
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		if len(vasilRels) > 0 {
			query := `
				UNWIND $rels AS r
				MATCH (a:Letter {char: r.from}), (b:Letter {char: r.to})
				MERGE (a)-[rel:VASIL]->(b)
				ON CREATE SET rel.weight = r.weight
				ON MATCH SET rel.weight = rel.weight + r.weight
			`
			if _, err := tx.Run(ctx, query, map[string]interface{}{"rels": vasilRels}); err != nil {
				return nil, err
			}
		}

		if len(fasilRels) > 0 {
			query := `
				UNWIND $rels AS r
				MATCH (a:Letter {char: r.from}), (b:Letter {char: r.to})
				MERGE (a)-[rel:FASIL]->(b)
				ON CREATE SET rel.weight = r.weight
				ON MATCH SET rel.weight = rel.weight + r.weight
			`
			if _, err := tx.Run(ctx, query, map[string]interface{}{"rels": fasilRels}); err != nil {
				return nil, err
			}
		}

		if len(medRels) > 0 {
			query := `
				UNWIND $rels AS r
				MATCH (a:Letter {char: r.from}), (b:Letter {char: r.to})
				MERGE (a)-[rel:MED]->(b)
				ON CREATE SET rel.weight = r.weight
				ON MATCH SET rel.weight = rel.weight + r.weight
			`
			if _, err := tx.Run(ctx, query, map[string]interface{}{"rels": medRels}); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

type MacroCorrelations struct {
	TotalFrequency int `json:"total_frequency"`
	AlemCounts map[string]int `json:"alem_counts"`
	TopSequences []SequencePattern `json:"top_sequences"`
}

type SequencePattern struct {
	Pattern string `json:"pattern"`
	Count int `json:"count"`
}

func (repo *Neo4jRepository) GetMacroCorrelations(ctx context.Context) (*MacroCorrelations, error) {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	correlations := &MacroCorrelations{
		AlemCounts: make(map[string]int),
	}

	// 1. Get Alem density based on incoming weights
	_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH ()-[r]->(b:Letter)
			RETURN b.alem AS alem, sum(r.weight) AS count
		`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			record := result.Record()
			alem, _ := record.Get("alem")
			count, _ := record.Get("count")
			correlations.AlemCounts[alem.(string)] = int(count.(int64))
			
			// Compute approximate frequency directly here (using previously defined weights)
			weight := 0
			switch alem.(string) {
			case "Ceberut": weight = 3
			case "Melekut": weight = 2
			case "Orta": weight = 1
			}
			correlations.TotalFrequency += int(count.(int64)) * weight
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	// 2. Find most frequent 3-length Alem patterns
	_, err = session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// A simple 3-hop path finding query (approximated for demo/performance)
		query := `
			MATCH (a:Letter)-[r1]->(b:Letter)-[r2]->(c:Letter)
			WHERE a.alem IS NOT NULL AND b.alem IS NOT NULL AND c.alem IS NOT NULL
			WITH a.alem + ' ➔ ' + b.alem + ' ➔ ' + c.alem AS pattern, sum(r1.weight + r2.weight) AS w
			RETURN pattern, w
			ORDER BY w DESC LIMIT 5
		`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			record := result.Record()
			pattern, _ := record.Get("pattern")
			w, _ := record.Get("w")
			correlations.TopSequences = append(correlations.TopSequences, SequencePattern{
				Pattern: pattern.(string),
				Count: int(w.(int64)),
			})
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return correlations, nil
}

type Transition struct {
	FromAlem string `json:"from_alem"`
	ToAlem   string `json:"to_alem"`
	Count    int    `json:"count"`
}

func (repo *Neo4jRepository) GetAlemTransitionMatrix(ctx context.Context) ([]Transition, error) {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:Letter)-[r]->(b:Letter)
			WHERE a.alem IS NOT NULL AND b.alem IS NOT NULL
			RETURN a.alem AS from_alem, b.alem AS to_alem, sum(r.weight) AS count
			ORDER BY count DESC
		`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var transitions []Transition
		for result.Next(ctx) {
			record := result.Record()
			f, _ := record.Get("from_alem")
			t, _ := record.Get("to_alem")
			c, _ := record.Get("count")
			transitions = append(transitions, Transition{
				FromAlem: f.(string),
				ToAlem:   t.(string),
				Count:    int(c.(int64)),
			})
		}
		return transitions, nil
	})

	if err != nil {
		return nil, err
	}
	return res.([]Transition), nil
}

type RootJourney struct {
	Root string          `json:"root"`
	Path []models.Letter `json:"path"`
}

func (repo *Neo4jRepository) GetRootWordJourney(ctx context.Context, chars []string) (*RootJourney, error) {
	journey := &RootJourney{Root: ""}
	for _, ch := range chars {
		journey.Root += ch
		l, _ := repo.GetLetter(ctx, ch)
		if l != nil {
			journey.Path = append(journey.Path, *l)
		}
	}
	return journey, nil
}

// SaveSurahSignatures saves the signatures as Surah nodes
func (repo *Neo4jRepository) SaveSurahSignatures(ctx context.Context, signatures []models.SurahSignature) error {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			UNWIND $surahs AS s
			MERGE (node:Surah {number: s.number})
			SET node.name = s.name,
			    node.ceberut = s.ceberut,
			    node.melekut = s.melekut,
			    node.orta = s.orta,
			    node.asagi = s.asagi
		`
		var surahMaps []map[string]interface{}
		for _, s := range signatures {
			surahMaps = append(surahMaps, map[string]interface{}{
				"number":  s.Number,
				"name":    s.Name,
				"ceberut": s.Ceberut,
				"melekut": s.Melekut,
				"orta":    s.Orta,
				"asagi":   s.Asagi,
			})
		}
		_, err := tx.Run(ctx, query, map[string]interface{}{"surahs": surahMaps})
		return nil, err
	})
	return err
}

func (repo *Neo4jRepository) GetAllSurahSignatures(ctx context.Context) ([]models.SurahSignature, error) {
	session := repo.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `MATCH (s:Surah) RETURN s.number, s.name, s.ceberut, s.melekut, s.orta, s.asagi`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var sigs []models.SurahSignature
		for result.Next(ctx) {
			record := result.Record()
			number, _ := record.Get("s.number")
			name, _ := record.Get("s.name")
			ceberut, _ := record.Get("s.ceberut")
			melekut, _ := record.Get("s.melekut")
			orta, _ := record.Get("s.orta")
			asagi, _ := record.Get("s.asagi")

			sigs = append(sigs, models.SurahSignature{
				Number:  int(number.(int64)),
				Name:    name.(string),
				Ceberut: ceberut.(float64),
				Melekut: melekut.(float64),
				Orta:    orta.(float64),
				Asagi:   asagi.(float64),
			})
		}
		return sigs, nil
	})
	if err != nil {
		return nil, err
	}
	return res.([]models.SurahSignature), nil
}
