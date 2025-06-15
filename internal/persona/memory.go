package persona

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MemoryEntry represents a single memory item with metadata
type MemoryEntry struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      []string               `json:"tags"`
	Weight    float64                `json:"weight"`      // Importance/relevance weight
	Context   map[string]interface{} `json:"context"`     // Additional context data
	Type      MemoryType             `json:"type"`        // Type of memory
	Decay     float64                `json:"decay"`       // Memory decay factor
}

// MemoryType defines different types of memories
type MemoryType string

const (
	MemoryTypeInteraction MemoryType = "interaction"
	MemoryTypeKnowledge   MemoryType = "knowledge"
	MemoryTypePersonal    MemoryType = "personal"
	MemoryTypeEmotional   MemoryType = "emotional"
	MemoryTypePattern     MemoryType = "pattern"
)

// Memory manages persona memory with short-term and long-term storage
type Memory struct {
	PersonaID     string                 `json:"persona_id"`
	Context       map[string]interface{} `json:"context"`        // Current conversation context
	ShortTerm     []MemoryEntry          `json:"short_term"`     // Recent memories
	LongTerm      []MemoryEntry          `json:"long_term"`      // Consolidated memories
	WorkingMemory []MemoryEntry          `json:"working_memory"` // Currently active memories
	
	// Configuration
	ShortTermLimit int     `json:"short_term_limit"` // Max short-term memories
	LongTermLimit  int     `json:"long_term_limit"`  // Max long-term memories
	DecayRate      float64 `json:"decay_rate"`       // Memory decay rate
}

// MemoryManager handles memory operations and consolidation
type MemoryManager struct {
	memory *Memory
}

// NewMemory creates a new memory instance for a persona
func NewMemory(personaID string) *Memory {
	return &Memory{
		PersonaID:      personaID,
		Context:        make(map[string]interface{}),
		ShortTerm:      make([]MemoryEntry, 0),
		LongTerm:       make([]MemoryEntry, 0),
		WorkingMemory:  make([]MemoryEntry, 0),
		ShortTermLimit: 50,  // Keep last 50 interactions in short-term
		LongTermLimit:  200, // Keep up to 200 consolidated long-term memories
		DecayRate:      0.95, // Memory strength decays by 5% over time
	}
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(memory *Memory) *MemoryManager {
	return &MemoryManager{
		memory: memory,
	}
}

// AddMemory adds a new memory entry
func (mm *MemoryManager) AddMemory(content string, memType MemoryType, weight float64, tags []string, context map[string]interface{}) {
	entry := MemoryEntry{
		ID:        mm.generateMemoryID(),
		Content:   content,
		Timestamp: time.Now(),
		Tags:      tags,
		Weight:    weight,
		Context:   context,
		Type:      memType,
		Decay:     1.0, // Fresh memory starts at full strength
	}
	
	// Add to short-term memory
	mm.memory.ShortTerm = append(mm.memory.ShortTerm, entry)
	
	// Trigger consolidation if short-term is full
	if len(mm.memory.ShortTerm) >= mm.memory.ShortTermLimit {
		mm.consolidateMemories()
	}
	
	// Update working memory with relevant entries
	mm.updateWorkingMemory(content, tags)
}

// RetrieveRelevant finds memories relevant to the given prompt
func (mm *MemoryManager) RetrieveRelevant(prompt string, limit int) []MemoryEntry {
	allMemories := make([]MemoryEntry, 0)
	
	// Combine all memory sources
	allMemories = append(allMemories, mm.memory.WorkingMemory...)
	allMemories = append(allMemories, mm.memory.ShortTerm...)
	allMemories = append(allMemories, mm.memory.LongTerm...)
	
	// Score memories by relevance
	scored := make([]struct {
		entry MemoryEntry
		score float64
	}, 0)
	
	promptLower := strings.ToLower(prompt)
	promptWords := strings.Fields(promptLower)
	
	for _, memory := range allMemories {
		score := mm.calculateRelevanceScore(memory, promptLower, promptWords)
		if score > 0.1 { // Only include memories with minimum relevance
			scored = append(scored, struct {
				entry MemoryEntry
				score float64
			}{memory, score})
		}
	}
	
	// Sort by relevance score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	
	// Return top results
	result := make([]MemoryEntry, 0)
	maxResults := limit
	if maxResults > len(scored) {
		maxResults = len(scored)
	}
	
	for i := 0; i < maxResults; i++ {
		result = append(result, scored[i].entry)
	}
	
	return result
}

// calculateRelevanceScore computes how relevant a memory is to the current prompt
func (mm *MemoryManager) calculateRelevanceScore(memory MemoryEntry, promptLower string, promptWords []string) float64 {
	score := 0.0
	contentLower := strings.ToLower(memory.Content)
	
	// Direct text similarity
	for _, word := range promptWords {
		if strings.Contains(contentLower, word) {
			score += 0.5
		}
	}
	
	// Tag matching
	for _, tag := range memory.Tags {
		tagLower := strings.ToLower(tag)
		for _, word := range promptWords {
			if strings.Contains(tagLower, word) || strings.Contains(word, tagLower) {
				score += 0.3
			}
		}
	}
	
	// Context relevance
	for key, value := range memory.Context {
		keyLower := strings.ToLower(key)
		valueStr := fmt.Sprintf("%v", value)
		valueLower := strings.ToLower(valueStr)
		
		if strings.Contains(promptLower, keyLower) || strings.Contains(promptLower, valueLower) {
			score += 0.2
		}
	}
	
	// Apply memory strength (decay factor)
	score *= memory.Decay
	
	// Apply base weight
	score *= memory.Weight
	
	// Recency bonus (more recent memories get slight boost)
	timeSince := time.Since(memory.Timestamp)
	if timeSince < time.Hour {
		score *= 1.2
	} else if timeSince < time.Hour*24 {
		score *= 1.1
	}
	
	// Memory type bonuses
	switch memory.Type {
	case MemoryTypeEmotional:
		score *= 1.15 // Emotional memories are more impactful
	case MemoryTypePattern:
		score *= 1.1  // Pattern recognition is valuable
	case MemoryTypePersonal:
		score *= 1.05 // Personal memories add character
	}
	
	return score
}

// consolidateMemories moves old short-term memories to long-term storage
func (mm *MemoryManager) consolidateMemories() {
	if len(mm.memory.ShortTerm) < mm.memory.ShortTermLimit/2 {
		return // Not enough memories to consolidate
	}
	
	// Sort short-term memories by importance and recency
	sort.Slice(mm.memory.ShortTerm, func(i, j int) bool {
		scoreI := mm.memory.ShortTerm[i].Weight * mm.memory.ShortTerm[i].Decay
		scoreJ := mm.memory.ShortTerm[j].Weight * mm.memory.ShortTerm[j].Decay
		
		if scoreI == scoreJ {
			return mm.memory.ShortTerm[i].Timestamp.After(mm.memory.ShortTerm[j].Timestamp)
		}
		return scoreI > scoreJ
	})
	
	// Move bottom half to long-term after consolidation
	midPoint := len(mm.memory.ShortTerm) / 2
	toConsolidate := mm.memory.ShortTerm[midPoint:]
	mm.memory.ShortTerm = mm.memory.ShortTerm[:midPoint]
	
	// Consolidate similar memories
	consolidated := mm.consolidateSimilarMemories(toConsolidate)
	
	// Add to long-term memory
	mm.memory.LongTerm = append(mm.memory.LongTerm, consolidated...)
	
	// Trim long-term memory if needed
	if len(mm.memory.LongTerm) > mm.memory.LongTermLimit {
		// Keep most important memories
		sort.Slice(mm.memory.LongTerm, func(i, j int) bool {
			scoreI := mm.memory.LongTerm[i].Weight * mm.memory.LongTerm[i].Decay
			scoreJ := mm.memory.LongTerm[j].Weight * mm.memory.LongTerm[j].Decay
			return scoreI > scoreJ
		})
		mm.memory.LongTerm = mm.memory.LongTerm[:mm.memory.LongTermLimit]
	}
	
	// Apply memory decay
	mm.applyMemoryDecay()
}

// consolidateSimilarMemories merges similar memories to reduce redundancy
func (mm *MemoryManager) consolidateSimilarMemories(memories []MemoryEntry) []MemoryEntry {
	if len(memories) == 0 {
		return memories
	}
	
	consolidated := make([]MemoryEntry, 0)
	used := make(map[int]bool)
	
	for i, memory := range memories {
		if used[i] {
			continue
		}
		
		similar := []MemoryEntry{memory}
		used[i] = true
		
		// Find similar memories
		for j := i + 1; j < len(memories); j++ {
			if used[j] {
				continue
			}
			
			if mm.areMemoriesSimilar(memory, memories[j]) {
				similar = append(similar, memories[j])
				used[j] = true
			}
		}
		
		// If we found similar memories, consolidate them
		if len(similar) > 1 {
			consolidated = append(consolidated, mm.mergeMemories(similar))
		} else {
			consolidated = append(consolidated, memory)
		}
	}
	
	return consolidated
}

// areMemoriesSimilar determines if two memories are similar enough to merge
func (mm *MemoryManager) areMemoriesSimilar(m1, m2 MemoryEntry) bool {
	// Same type and recent timing
	if m1.Type != m2.Type {
		return false
	}
	
	// Within 1 hour of each other
	if m1.Timestamp.Sub(m2.Timestamp).Abs() > time.Hour {
		return false
	}
	
	// Similar tags
	commonTags := 0
	for _, tag1 := range m1.Tags {
		for _, tag2 := range m2.Tags {
			if tag1 == tag2 {
				commonTags++
				break
			}
		}
	}
	
	if len(m1.Tags) > 0 && len(m2.Tags) > 0 {
		similarity := float64(commonTags) / float64(len(m1.Tags)+len(m2.Tags)-commonTags)
		return similarity > 0.3
	}
	
	// Similar content (simple word overlap check)
	words1 := strings.Fields(strings.ToLower(m1.Content))
	words2 := strings.Fields(strings.ToLower(m2.Content))
	
	commonWords := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 && len(w1) > 3 { // Only count significant words
				commonWords++
				break
			}
		}
	}
	
	if len(words1) > 0 && len(words2) > 0 {
		similarity := float64(commonWords) / float64(len(words1)+len(words2)-commonWords)
		return similarity > 0.2
	}
	
	return false
}

// mergeMemories combines multiple similar memories into one consolidated memory
func (mm *MemoryManager) mergeMemories(memories []MemoryEntry) MemoryEntry {
	if len(memories) == 1 {
		return memories[0]
	}
	
	// Use the most recent memory as base
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Timestamp.After(memories[j].Timestamp)
	})
	
	base := memories[0]
	
	// Combine content
	contents := make([]string, len(memories))
	for i, memory := range memories {
		contents[i] = memory.Content
	}
	
	// Create consolidated memory
	consolidated := MemoryEntry{
		ID:        mm.generateMemoryID(),
		Content:   fmt.Sprintf("Consolidated: %s", strings.Join(contents, " | ")),
		Timestamp: base.Timestamp,
		Tags:      mm.mergeTags(memories),
		Weight:    mm.calculateAverageWeight(memories),
		Context:   base.Context, // Use most recent context
		Type:      base.Type,
		Decay:     mm.calculateAverageDecay(memories),
	}
	
	return consolidated
}

// mergeTags combines tags from multiple memories, removing duplicates
func (mm *MemoryManager) mergeTags(memories []MemoryEntry) []string {
	tagMap := make(map[string]bool)
	
	for _, memory := range memories {
		for _, tag := range memory.Tags {
			tagMap[tag] = true
		}
	}
	
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	
	return tags
}

// calculateAverageWeight computes weighted average of memory weights
func (mm *MemoryManager) calculateAverageWeight(memories []MemoryEntry) float64 {
	if len(memories) == 0 {
		return 0.5
	}
	
	totalWeight := 0.0
	for _, memory := range memories {
		totalWeight += memory.Weight
	}
	
	return totalWeight / float64(len(memories))
}

// calculateAverageDecay computes weighted average of memory decay values
func (mm *MemoryManager) calculateAverageDecay(memories []MemoryEntry) float64 {
	if len(memories) == 0 {
		return 1.0
	}
	
	totalDecay := 0.0
	for _, memory := range memories {
		totalDecay += memory.Decay
	}
	
	return totalDecay / float64(len(memories))
}

// updateWorkingMemory refreshes the working memory with currently relevant entries
func (mm *MemoryManager) updateWorkingMemory(prompt string, tags []string) {
	// Clear current working memory
	mm.memory.WorkingMemory = make([]MemoryEntry, 0)
	
	// Add recent short-term memories
	recentLimit := 5
	if len(mm.memory.ShortTerm) < recentLimit {
		recentLimit = len(mm.memory.ShortTerm)
	}
	
	if recentLimit > 0 {
		recent := mm.memory.ShortTerm[len(mm.memory.ShortTerm)-recentLimit:]
		mm.memory.WorkingMemory = append(mm.memory.WorkingMemory, recent...)
	}
	
	// Add relevant long-term memories
	relevant := mm.RetrieveRelevant(prompt, 3) // Get top 3 relevant long-term memories
	for _, memory := range relevant {
		// Avoid duplicates
		isDuplicate := false
		for _, working := range mm.memory.WorkingMemory {
			if working.ID == memory.ID {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			mm.memory.WorkingMemory = append(mm.memory.WorkingMemory, memory)
		}
	}
}

// applyMemoryDecay reduces the strength of older memories
func (mm *MemoryManager) applyMemoryDecay() {
	// Apply decay to long-term memories
	for i := range mm.memory.LongTerm {
		timeSince := time.Since(mm.memory.LongTerm[i].Timestamp)
		daysSince := timeSince.Hours() / 24
		
		// Apply exponential decay based on days
		decayFactor := 1.0
		if daysSince > 0 {
			decayFactor = 1.0 / (1.0 + daysSince*0.1) // Gradual decay over time
		}
		
		mm.memory.LongTerm[i].Decay *= mm.memory.DecayRate * decayFactor
		
		// Remove very weak memories
		if mm.memory.LongTerm[i].Decay < 0.1 {
			mm.memory.LongTerm = append(mm.memory.LongTerm[:i], mm.memory.LongTerm[i+1:]...)
		}
	}
}

// UpdateContext updates the current conversation context
func (mm *MemoryManager) UpdateContext(key string, value interface{}) {
	mm.memory.Context[key] = value
}

// GetContext retrieves a value from the current context
func (mm *MemoryManager) GetContext(key string) (interface{}, bool) {
	value, exists := mm.memory.Context[key]
	return value, exists
}

// ClearContext clears the current conversation context
func (mm *MemoryManager) ClearContext() {
	mm.memory.Context = make(map[string]interface{})
}

// generateMemoryID creates a unique ID for a memory entry
func (mm *MemoryManager) generateMemoryID() string {
	return fmt.Sprintf("%s_%d", mm.memory.PersonaID, time.Now().UnixNano())
}

// GetMemoryStats returns statistics about the memory system
func (mm *MemoryManager) GetMemoryStats() map[string]interface{} {
	return map[string]interface{}{
		"short_term_count":    len(mm.memory.ShortTerm),
		"long_term_count":     len(mm.memory.LongTerm),
		"working_memory_count": len(mm.memory.WorkingMemory),
		"total_memories":      len(mm.memory.ShortTerm) + len(mm.memory.LongTerm),
		"context_keys":        len(mm.memory.Context),
	}
}

// ExportMemory exports memory data for persistence
func (mm *MemoryManager) ExportMemory() ([]byte, error) {
	return json.Marshal(mm.memory)
}

// ImportMemory imports memory data from persistence
func (mm *MemoryManager) ImportMemory(data []byte) error {
	return json.Unmarshal(data, mm.memory)
}