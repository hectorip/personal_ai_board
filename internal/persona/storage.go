package persona

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Storage handles database operations for personas
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance
func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// SavePersona saves a persona to the database
func (s *Storage) SavePersona(persona *Persona) error {
	// Serialize traits
	traitsData, err := json.Marshal(persona.Traits)
	if err != nil {
		return fmt.Errorf("failed to serialize traits: %w", err)
	}

	// Export memory
	memoryData, err := persona.memoryMgr.ExportMemory()
	if err != nil {
		return fmt.Errorf("failed to export memory: %w", err)
	}

	// Insert or update persona
	query := `
		INSERT OR REPLACE INTO personas (
			id, name, description, traits_config, memory_data,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		persona.ID,
		persona.Name,
		persona.Description,
		string(traitsData),
		string(memoryData),
		persona.createdAt,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save persona: %w", err)
	}

	return nil
}

// LoadPersona loads a persona from the database
func (s *Storage) LoadPersona(id string, llmProvider LLMProvider, logger Logger) (*Persona, error) {
	query := `
		SELECT id, name, description, traits_config, memory_data,
		       created_at, updated_at
		FROM personas WHERE id = ?
	`

	var persona Persona
	var traitsData, memoryData string
	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(query, id).Scan(
		&persona.ID,
		&persona.Name,
		&persona.Description,
		&traitsData,
		&memoryData,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load persona: %w", err)
	}

	// Deserialize traits
	var traits PersonalityTraits
	if err := json.Unmarshal([]byte(traitsData), &traits); err != nil {
		return nil, fmt.Errorf("failed to deserialize traits: %w", err)
	}

	// Create memory system
	memory := NewMemory(id)
	memoryMgr := NewMemoryManager(memory)

	// Import memory data
	if memoryData != "" {
		if err := memoryMgr.ImportMemory([]byte(memoryData)); err != nil {
			logger.Warn("Failed to import memory data", "persona_id", id, "error", err)
		}
	}

	persona.Traits = &traits
	persona.memoryMgr = memoryMgr
	persona.db = s.db
	persona.llmProvider = llmProvider
	persona.logger = logger
	persona.createdAt = createdAt
	persona.updatedAt = updatedAt

	return &persona, nil
}

// ListPersonas returns all personas
func (s *Storage) ListPersonas() ([]PersonaInfo, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM personas
		ORDER BY updated_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query personas: %w", err)
	}
	defer rows.Close()

	var personas []PersonaInfo
	for rows.Next() {
		var info PersonaInfo
		err := rows.Scan(
			&info.ID,
			&info.Name,
			&info.Description,
			&info.CreatedAt,
			&info.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, info)
	}

	return personas, nil
}

// DeletePersona removes a persona from the database
func (s *Storage) DeletePersona(id string) error {
	// Delete from personas table
	_, err := s.db.Exec("DELETE FROM personas WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete persona: %w", err)
	}

	// Delete related LLM logs
	_, err = s.db.Exec("DELETE FROM llm_interaction_logs WHERE persona_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete LLM logs: %w", err)
	}

	return nil
}

// PersonaInfo contains basic information about a persona
type PersonaInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Database operations for persona memory and interactions
func (p *Persona) loadMemoryFromDB() error {
	if p.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Load memory is handled in the storage layer during persona loading
	// This method is called during initialization but actual loading
	// happens in LoadPersona method
	return nil
}

// saveMemoryToDB saves memory data to database
func (p *Persona) saveMemoryToDB() error {
	if p.db == nil {
		return fmt.Errorf("database connection not available")
	}

	memoryData, err := p.memoryMgr.ExportMemory()
	if err != nil {
		return fmt.Errorf("failed to export memory: %w", err)
	}

	query := `UPDATE personas SET memory_data = ?, updated_at = ? WHERE id = ?`
	_, err = p.db.Exec(query, string(memoryData), time.Now(), p.ID)
	if err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}

	return nil
}

// logInteraction logs the LLM interaction to database
func (p *Persona) logInteraction(req LLMRequest, resp *LLMResponse, duration time.Duration) {
	if p.db == nil {
		p.logger.Warn("Database not available for logging interaction")
		return
	}

	// Prepare context data
	contextData, _ := json.Marshal(req.Context)

	query := `
		INSERT INTO llm_interaction_logs (
			id, persona_id, prompt, system_message, response,
			model_name, temperature, max_tokens, tokens_used,
			duration_ms, context_data, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	logID := fmt.Sprintf("%s_%d", p.ID, time.Now().UnixNano())

	_, err := p.db.Exec(query,
		logID,
		p.ID,
		req.Prompt,
		req.SystemMsg,
		resp.Content,
		resp.Model,
		req.Temperature,
		req.MaxTokens,
		resp.TokensUsed,
		duration.Milliseconds(),
		string(contextData),
		time.Now(),
	)

	if err != nil {
		p.logger.Error("Failed to log LLM interaction", "persona_id", p.ID, "error", err)
	}
}

// initializePersonality sets up the persona's core personality patterns
func (p *Persona) initializePersonality() error {
	// Add core personality memories based on traits
	p.addCorePersonalityMemories()

	// Set up initial context
	p.memoryMgr.UpdateContext("persona_name", p.Name)
	p.memoryMgr.UpdateContext("persona_type", p.Traits.Config.PersonaType)
	p.memoryMgr.UpdateContext("expertise_areas", p.Traits.ExpertiseAreas)
	p.memoryMgr.UpdateContext("initialized_at", time.Now())

	return nil
}

// addCorePersonalityMemories adds fundamental personality patterns to memory
func (p *Persona) addCorePersonalityMemories() {
	// Add expertise areas as knowledge memories
	for _, area := range p.Traits.ExpertiseAreas {
		p.memoryMgr.AddMemory(
			fmt.Sprintf("I have expertise in %s", area),
			MemoryTypeKnowledge,
			0.9,
			[]string{"expertise", area, "core_knowledge"},
			map[string]interface{}{"type": "expertise"},
		)
	}

	// Add speaking patterns as behavioral memories
	for _, phrase := range p.Traits.SpeakingPatterns.CommonPhrases {
		p.memoryMgr.AddMemory(
			fmt.Sprintf("I often express ideas like: %s", phrase),
			MemoryTypePersonal,
			0.8,
			[]string{"communication", "speaking_pattern", "phrase"},
			map[string]interface{}{"type": "speaking_pattern"},
		)
	}

	// Add emotional triggers as personal memories
	for _, energizer := range p.Traits.EmotionalTriggers.Energizers {
		p.memoryMgr.AddMemory(
			fmt.Sprintf("I get energized by %s", energizer),
			MemoryTypeEmotional,
			0.7,
			[]string{"emotional", "energizer", "motivation"},
			map[string]interface{}{"type": "energizer", "valence": "positive"},
		)
	}

	for _, frustration := range p.Traits.EmotionalTriggers.Frustrations {
		p.memoryMgr.AddMemory(
			fmt.Sprintf("I get frustrated by %s", frustration),
			MemoryTypeEmotional,
			0.7,
			[]string{"emotional", "frustration", "trigger"},
			map[string]interface{}{"type": "frustration", "valence": "negative"},
		)
	}

	// Add core trait patterns
	creativity := p.Traits.GetIntTrait("core_dimensions", "creativity")
	if creativity >= 8 {
		p.memoryMgr.AddMemory(
			"I naturally think creatively and look for innovative solutions",
			MemoryTypePattern,
			0.9,
			[]string{"creativity", "innovation", "core_trait"},
			map[string]interface{}{"trait": "creativity", "level": creativity},
		)
	}

	analytical := p.Traits.GetIntTrait("core_dimensions", "analytical")
	if analytical >= 8 {
		p.memoryMgr.AddMemory(
			"I approach problems systematically and rely on logical analysis",
			MemoryTypePattern,
			0.9,
			[]string{"analytical", "logic", "systematic", "core_trait"},
			map[string]interface{}{"trait": "analytical", "level": analytical},
		)
	}

	riskTolerance := p.Traits.GetIntTrait("core_dimensions", "risk_tolerance")
	if riskTolerance >= 8 {
		p.memoryMgr.AddMemory(
			"I'm comfortable with uncertainty and willing to take calculated risks",
			MemoryTypePattern,
			0.8,
			[]string{"risk", "uncertainty", "bold", "core_trait"},
			map[string]interface{}{"trait": "risk_tolerance", "level": riskTolerance},
		)
	} else if riskTolerance <= 3 {
		p.memoryMgr.AddMemory(
			"I prefer careful, well-planned approaches and avoid unnecessary risks",
			MemoryTypePattern,
			0.8,
			[]string{"caution", "planning", "conservative", "core_trait"},
			map[string]interface{}{"trait": "risk_tolerance", "level": riskTolerance},
		)
	}
}

// UpdateTraits updates the persona's traits configuration
func (p *Persona) UpdateTraits(newTraitsConfig string) error {
	loader := NewTraitLoader("config")
	newTraits, err := loader.LoadPersonalityConfig(newTraitsConfig)
	if err != nil {
		return fmt.Errorf("failed to load new traits: %w", err)
	}

	p.Traits = newTraits
	p.updatedAt = time.Now()

	// Save to database
	if err := p.saveMemoryToDB(); err != nil {
		p.logger.Warn("Failed to save updated traits to database", "persona_id", p.ID, "error", err)
	}

	// Add memory about the update
	p.memoryMgr.AddMemory(
		"My personality traits were updated",
		MemoryTypePersonal,
		0.7,
		[]string{"personality", "update", "traits"},
		map[string]interface{}{
			"event":     "traits_update",
			"timestamp": time.Now(),
		},
	)

	p.logger.Info("Persona traits updated", "persona_id", p.ID)
	return nil
}

// GetPersonalityProfile returns a summary of the persona's personality
func (p *Persona) GetPersonalityProfile() map[string]interface{} {
	profile := map[string]interface{}{
		"id":           p.ID,
		"name":         p.Name,
		"description":  p.Description,
		"persona_type": p.Traits.Config.PersonaType,
		"created_at":   p.createdAt,
		"updated_at":   p.updatedAt,
	}

	// Add key traits
	profile["core_traits"] = map[string]interface{}{
		"creativity":     p.Traits.GetIntTrait("core_dimensions", "creativity"),
		"analytical":     p.Traits.GetIntTrait("core_dimensions", "analytical"),
		"optimism":       p.Traits.GetIntTrait("core_dimensions", "optimism"),
		"risk_tolerance": p.Traits.GetIntTrait("core_dimensions", "risk_tolerance"),
		"empathy":        p.Traits.GetIntTrait("core_dimensions", "empathy"),
		"assertiveness":  p.Traits.GetIntTrait("core_dimensions", "assertiveness"),
	}

	// Add communication style
	profile["communication_style"] = map[string]interface{}{
		"formality":  p.Traits.GetStringTrait("communication_style", "formality"),
		"directness": p.Traits.GetStringTrait("communication_style", "directness"),
		"verbosity":  p.Traits.GetStringTrait("communication_style", "verbosity"),
	}

	// Add expertise and memory stats
	profile["expertise_areas"] = p.Traits.ExpertiseAreas
	profile["memory_stats"] = p.memoryMgr.GetMemoryStats()

	return profile
}

// Clone creates a copy of the persona with slight variations
func (p *Persona) Clone(newID, newName string) (*Persona, error) {
	// Create a copy of traits with small random variations
	clonedTraits := *p.Traits

	// Apply small random variations to core dimensions (±1 point)
	for key := range clonedTraits.CoreDimensions {
		if intVal := clonedTraits.GetIntTrait("core_dimensions", key); intVal > 0 {
			variation := (time.Now().UnixNano() % 3) - 1 // -1, 0, or 1
			newVal := intVal + int(variation)
			if newVal < 1 {
				newVal = 1
			} else if newVal > 10 {
				newVal = 10
			}
			clonedTraits.CoreDimensions[key] = float64(newVal)
		}
	}

	// Create new persona with varied traits
	traitsData, err := json.Marshal(&clonedTraits)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize cloned traits: %w", err)
	}

	// Create new persona instance
	newPersona, err := New(newID, newName, p.Description, string(traitsData), p.db, p.llmProvider, p.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloned persona: %w", err)
	}

	// Add memory about being cloned
	newPersona.memoryMgr.AddMemory(
		fmt.Sprintf("I am a variation of %s with similar but unique traits", p.Name),
		MemoryTypePersonal,
		0.8,
		[]string{"identity", "clone", "origin"},
		map[string]interface{}{
			"original_persona": p.ID,
			"clone_event":      time.Now(),
		},
	)

	return newPersona, nil
}
