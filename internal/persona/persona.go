package persona

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Persona represents a complete AI personality with traits, memory, and behavior
type Persona struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Traits      *PersonalityTraits `json:"traits"`
	memoryMgr   *MemoryManager
	db          *sql.DB
	llmProvider LLMProvider
	logger      Logger
	createdAt   time.Time
	updatedAt   time.Time
}

// LLMProvider interface for AI model integration
type LLMProvider interface {
	GenerateResponse(ctx context.Context, req LLMRequest) (*LLMResponse, error)
	GetModelInfo() ModelInfo
}

// LLMRequest represents a request to the LLM
type LLMRequest struct {
	Prompt      string                 `json:"prompt"`
	SystemMsg   string                 `json:"system_message"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Context     map[string]interface{} `json:"context"`
}

// LLMResponse represents the response from the LLM
type LLMResponse struct {
	Content      string        `json:"content"`
	TokensUsed   int           `json:"tokens_used"`
	Model        string        `json:"model"`
	Duration     time.Duration `json:"duration"`
	FinishReason string        `json:"finish_reason"`
}

// ModelInfo contains information about the LLM model
type ModelInfo struct {
	Name      string  `json:"name"`
	Provider  string  `json:"provider"`
	MaxTokens int     `json:"max_tokens"`
	CostPer1K float64 `json:"cost_per_1k"`
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// ThinkingContext represents the context for a thinking session
type ThinkingContext struct {
	Topic               string                 `json:"topic"`
	ProjectContext      map[string]interface{} `json:"project_context"`
	BoardContext        map[string]interface{} `json:"board_context"`
	ConversationHistory []ConversationTurn     `json:"conversation_history"`
	EmotionalState      string                 `json:"emotional_state"`
	Focus               string                 `json:"focus"`
}

// ConversationTurn represents a single turn in a conversation
type ConversationTurn struct {
	Speaker   string    `json:"speaker"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ThinkingResult represents the output of a persona's thinking process
type ThinkingResult struct {
	Response        string             `json:"response"`
	Reasoning       string             `json:"reasoning"`
	Confidence      float64            `json:"confidence"`
	EmotionalTone   string             `json:"emotional_tone"`
	KeyInsights     []string           `json:"key_insights"`
	Questions       []string           `json:"questions"`
	Recommendations []string           `json:"recommendations"`
	MemoriesUsed    []string           `json:"memories_used"`
	TraitsInfluence map[string]float64 `json:"traits_influence"`
}

// New creates a new persona with the specified configuration
func New(id, name, description string, traitsConfig string, db *sql.DB, llmProvider LLMProvider, logger Logger) (*Persona, error) {
	// Load personality traits
	loader := NewTraitLoader("config")
	traits, err := loader.LoadPersonalityConfig(traitsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load personality traits: %w", err)
	}

	// Create memory system
	memory := NewMemory(id)
	memoryMgr := NewMemoryManager(memory)

	persona := &Persona{
		ID:          id,
		Name:        name,
		Description: description,
		Traits:      traits,
		memoryMgr:   memoryMgr,
		db:          db,
		llmProvider: llmProvider,
		logger:      logger,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	// Load existing memory from database
	if err := persona.loadMemoryFromDB(); err != nil {
		logger.Warn("Failed to load existing memory", "persona_id", id, "error", err)
	}

	// Initialize persona with core personality
	if err := persona.initializePersonality(); err != nil {
		return nil, fmt.Errorf("failed to initialize personality: %w", err)
	}

	return persona, nil
}

// NewFromJSONTraits creates a new persona with traits loaded from JSON string
func NewFromJSONTraits(id, name, description string, traitsJSON string, db *sql.DB, llmProvider LLMProvider, logger Logger) (*Persona, error) {
	// Load personality traits from JSON string
	traits, err := LoadPersonalityConfigFromJSONSimple(traitsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load personality traits from JSON: %w", err)
	}

	// Create memory system
	memory := NewMemory(id)
	memoryMgr := NewMemoryManager(memory)

	persona := &Persona{
		ID:          id,
		Name:        name,
		Description: description,
		Traits:      traits,
		memoryMgr:   memoryMgr,
		db:          db,
		llmProvider: llmProvider,
		logger:      logger,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	// Load existing memory from database
	if err := persona.loadMemoryFromDB(); err != nil {
		logger.Warn("Failed to load existing memory", "persona_id", id, "error", err)
	}

	// Initialize persona with core personality
	if err := persona.initializePersonality(); err != nil {
		return nil, fmt.Errorf("failed to initialize personality: %w", err)
	}

	return persona, nil
}

// Think is the main method for persona reasoning and response generation
func (p *Persona) Think(ctx context.Context, prompt string, context ThinkingContext) (*ThinkingResult, error) {
	startTime := time.Now()

	p.logger.Debug("Persona thinking started", "persona_id", p.ID, "prompt_length", len(prompt))

	// Update emotional state based on context
	emotionalState := p.determineEmotionalState(context)

	// Apply context-specific trait modifications
	workingTraits := p.applyContextualTraits(context, emotionalState)

	// Retrieve relevant memories
	relevantMemories := p.memoryMgr.RetrieveRelevant(prompt, 5)

	// Build the complete prompt for LLM
	enhancedPrompt, systemMessage := p.buildPrompt(prompt, context, relevantMemories, workingTraits, emotionalState)

	// Determine LLM parameters based on personality
	temperature := p.calculateTemperature(workingTraits)
	maxTokens := p.calculateMaxTokens(workingTraits)

	// Generate response using LLM
	llmReq := LLMRequest{
		Prompt:      enhancedPrompt,
		SystemMsg:   systemMessage,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Context:     context.ProjectContext,
	}

	llmResp, err := p.llmProvider.GenerateResponse(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse and enhance the response
	result, err := p.processLLMResponse(llmResp, workingTraits, relevantMemories)
	if err != nil {
		return nil, fmt.Errorf("failed to process LLM response: %w", err)
	}

	// Store the interaction in memory
	p.storeInteraction(prompt, result, context, emotionalState)

	// Log the interaction
	p.logInteraction(llmReq, llmResp, time.Since(startTime))

	p.logger.Debug("Persona thinking completed", "persona_id", p.ID, "duration", time.Since(startTime))

	return result, nil
}

// determineEmotionalState analyzes context to determine current emotional state
func (p *Persona) determineEmotionalState(context ThinkingContext) string {
	// Check for explicit emotional state
	if context.EmotionalState != "" {
		return context.EmotionalState
	}

	// Analyze conversation history for emotional cues
	if len(context.ConversationHistory) > 0 {
		recentTurns := context.ConversationHistory
		if len(recentTurns) > 3 {
			recentTurns = recentTurns[len(recentTurns)-3:]
		}

		// Check for emotional triggers
		for _, turn := range recentTurns {
			lowerContent := strings.ToLower(turn.Content)

			// Check energizers
			for _, energizer := range p.Traits.EmotionalTriggers.Energizers {
				if strings.Contains(lowerContent, strings.ToLower(energizer)) {
					return "excited"
				}
			}

			// Check frustrations
			for _, frustration := range p.Traits.EmotionalTriggers.Frustrations {
				if strings.Contains(lowerContent, strings.ToLower(frustration)) {
					return "frustrated"
				}
			}
		}
	}

	// Default to neutral state
	return "neutral"
}

// applyContextualTraits modifies traits based on current context and emotional state
func (p *Persona) applyContextualTraits(context ThinkingContext, emotionalState string) *PersonalityTraits {
	// Start with base traits
	workingTraits := p.Traits

	// Apply emotional state modifiers
	if _, exists := p.Traits.ResponseModifiers[emotionalState]; exists {
		workingTraits = p.Traits.ApplyContextModifier(emotionalState)
	}

	// Apply focus-specific modifiers
	if context.Focus != "" {
		if _, exists := p.Traits.ResponseModifiers[context.Focus]; exists {
			workingTraits = workingTraits.ApplyContextModifier(context.Focus)
		}
	}

	return workingTraits
}

// buildPrompt constructs the complete prompt for the LLM including personality context
func (p *Persona) buildPrompt(prompt string, context ThinkingContext, memories []MemoryEntry, traits *PersonalityTraits, emotionalState string) (string, string) {
	// Build system message with personality
	systemMessage := p.buildSystemMessage(traits, emotionalState)

	// Build the enhanced prompt
	var promptBuilder strings.Builder

	// Add conversation history context
	if len(context.ConversationHistory) > 0 {
		promptBuilder.WriteString("## Recent Conversation:\n")
		for _, turn := range context.ConversationHistory {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", turn.Speaker, turn.Content))
		}
		promptBuilder.WriteString("\n")
	}

	// Add relevant memories
	if len(memories) > 0 {
		promptBuilder.WriteString("## Relevant Context from Memory:\n")
		for _, memory := range memories {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", memory.Content))
		}
		promptBuilder.WriteString("\n")
	}

	// Add project context if available
	if len(context.ProjectContext) > 0 {
		promptBuilder.WriteString("## Project Context:\n")
		for key, value := range context.ProjectContext {
			promptBuilder.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		promptBuilder.WriteString("\n")
	}

	// Add the main prompt
	promptBuilder.WriteString("## Current Question/Topic:\n")
	promptBuilder.WriteString(prompt)
	promptBuilder.WriteString("\n\n")

	// Add personality-specific instruction
	promptBuilder.WriteString(p.buildPersonalityInstruction(traits, emotionalState))

	return promptBuilder.String(), systemMessage
}

// buildSystemMessage creates the system message that defines the persona's behavior
func (p *Persona) buildSystemMessage(traits *PersonalityTraits, emotionalState string) string {
	var msgBuilder strings.Builder

	// Basic identity
	msgBuilder.WriteString(fmt.Sprintf("You are %s, %s.\n\n", p.Name, p.Description))

	// Core personality traits
	msgBuilder.WriteString("## Your Personality:\n")

	// Communication style
	formality := traits.GetStringTrait("communication_style", "formality")
	directness := traits.GetStringTrait("communication_style", "directness")
	verbosity := traits.GetStringTrait("communication_style", "verbosity")

	msgBuilder.WriteString(fmt.Sprintf("- Communication: %s, %s, %s\n", formality, directness, verbosity))

	// Key traits (top traits that are high or low)
	creativity := traits.GetIntTrait("core_dimensions", "creativity")
	analytical := traits.GetIntTrait("core_dimensions", "analytical")
	optimism := traits.GetIntTrait("core_dimensions", "optimism")
	riskTolerance := traits.GetIntTrait("core_dimensions", "risk_tolerance")

	if creativity >= 8 {
		msgBuilder.WriteString("- You are highly creative and innovative\n")
	}
	if analytical >= 8 {
		msgBuilder.WriteString("- You are deeply analytical and logical\n")
	}
	if optimism >= 8 {
		msgBuilder.WriteString("- You maintain a very positive outlook\n")
	} else if optimism <= 3 {
		msgBuilder.WriteString("- You tend toward skepticism and caution\n")
	}
	if riskTolerance >= 8 {
		msgBuilder.WriteString("- You embrace uncertainty and calculated risks\n")
	}

	// Expertise areas
	if len(traits.ExpertiseAreas) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("- Your expertise: %s\n", strings.Join(traits.ExpertiseAreas, ", ")))
	}

	// Speaking patterns
	if len(traits.SpeakingPatterns.CommonPhrases) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("- You often say things like: %s\n", strings.Join(traits.SpeakingPatterns.CommonPhrases[:min(3, len(traits.SpeakingPatterns.CommonPhrases))], ", ")))
	}

	if len(traits.SpeakingPatterns.AvoidsPhrases) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("- You avoid saying: %s\n", strings.Join(traits.SpeakingPatterns.AvoidsPhrases[:min(2, len(traits.SpeakingPatterns.AvoidsPhrases))], ", ")))
	}

	// Current emotional state
	msgBuilder.WriteString(fmt.Sprintf("\n## Current State: %s\n", emotionalState))

	// Behavioral instructions
	msgBuilder.WriteString("\n## Instructions:\n")
	msgBuilder.WriteString("- Stay true to your personality traits and communication style\n")
	msgBuilder.WriteString("- Use your expertise and experience to provide valuable insights\n")
	msgBuilder.WriteString("- Be authentic to your character while being helpful\n")

	if traits.GetIntTrait("response_patterns", "question_tendency") >= 7 {
		msgBuilder.WriteString("- Ask probing questions to better understand the situation\n")
	}

	if traits.GetIntTrait("response_patterns", "example_usage") >= 7 {
		msgBuilder.WriteString("- Use relevant examples and analogies to illustrate your points\n")
	}

	return msgBuilder.String()
}

// buildPersonalityInstruction creates personality-specific response guidance
func (p *Persona) buildPersonalityInstruction(traits *PersonalityTraits, emotionalState string) string {
	var instrBuilder strings.Builder

	instrBuilder.WriteString("## Response Guidance:\n")

	// Response style based on traits
	contrarian := traits.GetIntTrait("response_patterns", "contrarian_level")
	solutionOriented := traits.GetIntTrait("response_patterns", "solution_orientation")

	if contrarian >= 7 {
		instrBuilder.WriteString("- Challenge assumptions and explore alternative perspectives\n")
	}

	if solutionOriented >= 7 {
		instrBuilder.WriteString("- Focus on actionable solutions and next steps\n")
	}

	// Decision making style
	dataVsIntuition := traits.GetIntTrait("decision_making", "data_vs_intuition")
	if dataVsIntuition >= 8 {
		instrBuilder.WriteString("- Support your points with data and evidence\n")
	} else if dataVsIntuition <= 3 {
		instrBuilder.WriteString("- Trust your instincts and share intuitive insights\n")
	}

	// Emotional state specific guidance
	switch emotionalState {
	case "excited":
		instrBuilder.WriteString("- You're feeling energized and enthusiastic about this topic\n")
	case "frustrated":
		instrBuilder.WriteString("- You're feeling frustrated and may be more direct than usual\n")
	case "focused":
		instrBuilder.WriteString("- You're in deep focus mode and thinking systematically\n")
	}

	return instrBuilder.String()
}

// calculateTemperature determines LLM temperature based on personality traits
func (p *Persona) calculateTemperature(traits *PersonalityTraits) float64 {
	baseTemp := 0.7

	// Creativity increases temperature
	creativity := float64(traits.GetIntTrait("core_dimensions", "creativity")) / 10.0

	// Analytical thinking decreases temperature
	analytical := float64(traits.GetIntTrait("core_dimensions", "analytical")) / 10.0

	// Risk tolerance affects temperature
	riskTolerance := float64(traits.GetIntTrait("core_dimensions", "risk_tolerance")) / 10.0

	// Calculate weighted temperature
	temperature := baseTemp + (creativity * 0.3) - (analytical * 0.2) + (riskTolerance * 0.1)

	// Clamp to valid range
	if temperature < 0.1 {
		temperature = 0.1
	} else if temperature > 1.0 {
		temperature = 1.0
	}

	return temperature
}

// calculateMaxTokens determines max tokens based on verbosity and other traits
func (p *Persona) calculateMaxTokens(traits *PersonalityTraits) int {
	baseTokens := 500

	verbosity := traits.GetStringTrait("communication_style", "verbosity")

	switch verbosity {
	case "terse":
		return baseTokens / 2
	case "concise":
		return int(float64(baseTokens) * 0.7)
	case "balanced":
		return baseTokens
	case "detailed":
		return int(float64(baseTokens) * 1.5)
	case "verbose":
		return baseTokens * 2
	default:
		return baseTokens
	}
}

// processLLMResponse analyzes and enhances the raw LLM response
func (p *Persona) processLLMResponse(llmResp *LLMResponse, traits *PersonalityTraits, memories []MemoryEntry) (*ThinkingResult, error) {
	// Extract key insights and questions from the response
	insights := p.extractInsights(llmResp.Content)
	questions := p.extractQuestions(llmResp.Content)
	recommendations := p.extractRecommendations(llmResp.Content)

	// Calculate confidence based on response and personality
	confidence := p.calculateConfidence(llmResp, traits)

	// Determine emotional tone
	emotionalTone := p.analyzeEmotionalTone(llmResp.Content, traits)

	// Track which memories were used
	memoriesUsed := make([]string, len(memories))
	for i, memory := range memories {
		memoriesUsed[i] = memory.ID
	}

	// Analyze trait influences
	traitInfluence := p.analyzeTraitInfluence(traits)

	return &ThinkingResult{
		Response:        llmResp.Content,
		Reasoning:       p.extractReasoning(llmResp.Content),
		Confidence:      confidence,
		EmotionalTone:   emotionalTone,
		KeyInsights:     insights,
		Questions:       questions,
		Recommendations: recommendations,
		MemoriesUsed:    memoriesUsed,
		TraitsInfluence: traitInfluence,
	}, nil
}

// storeInteraction saves the thinking session to memory
func (p *Persona) storeInteraction(prompt string, result *ThinkingResult, context ThinkingContext, emotionalState string) {
	// Create memory entry for the interaction
	memoryContext := map[string]interface{}{
		"topic":           context.Topic,
		"emotional_state": emotionalState,
		"confidence":      result.Confidence,
		"focus":           context.Focus,
	}

	// Store the prompt and response
	promptContent := fmt.Sprintf("Question: %s", prompt)
	p.memoryMgr.AddMemory(
		promptContent,
		MemoryTypeInteraction,
		0.8, // High weight for interactions
		[]string{"interaction", "question", context.Topic},
		memoryContext,
	)

	responseContent := fmt.Sprintf("Response: %s", result.Response)
	p.memoryMgr.AddMemory(
		responseContent,
		MemoryTypeInteraction,
		0.8,
		[]string{"interaction", "response", context.Topic},
		memoryContext,
	)

	// Store key insights as separate memories
	for _, insight := range result.KeyInsights {
		p.memoryMgr.AddMemory(
			insight,
			MemoryTypeKnowledge,
			0.9, // High weight for insights
			[]string{"insight", "knowledge", context.Topic},
			memoryContext,
		)
	}

	// Update conversation context
	p.memoryMgr.UpdateContext("last_interaction_time", time.Now())
	p.memoryMgr.UpdateContext("last_topic", context.Topic)
	p.memoryMgr.UpdateContext("last_emotional_state", emotionalState)
}

// Helper functions for response processing
func (p *Persona) extractInsights(content string) []string {
	// Simple insight extraction - look for key patterns
	insights := make([]string, 0)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "key insight") ||
			strings.Contains(strings.ToLower(line), "important") ||
			strings.Contains(strings.ToLower(line), "crucial") {
			insights = append(insights, line)
		}
	}

	return insights
}

func (p *Persona) extractQuestions(content string) []string {
	questions := make([]string, 0)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "?") {
			questions = append(questions, line)
		}
	}

	return questions
}

func (p *Persona) extractRecommendations(content string) []string {
	recommendations := make([]string, 0)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "recommend") ||
			strings.Contains(strings.ToLower(line), "suggest") ||
			strings.Contains(strings.ToLower(line), "should") {
			recommendations = append(recommendations, line)
		}
	}

	return recommendations
}

func (p *Persona) extractReasoning(content string) string {
	// Extract reasoning patterns from the response
	if strings.Contains(strings.ToLower(content), "because") ||
		strings.Contains(strings.ToLower(content), "therefore") ||
		strings.Contains(strings.ToLower(content), "given that") {
		return "Logical reasoning detected"
	}

	if strings.Contains(strings.ToLower(content), "i feel") ||
		strings.Contains(strings.ToLower(content), "intuitively") {
		return "Intuitive reasoning detected"
	}

	return "Mixed reasoning approach"
}

func (p *Persona) calculateConfidence(llmResp *LLMResponse, traits *PersonalityTraits) float64 {
	baseConfidence := 0.7

	// Adjust based on response length and quality
	if len(llmResp.Content) > 200 {
		baseConfidence += 0.1
	}

	// Adjust based on personality traits
	confidence := traits.GetIntTrait("core_dimensions", "assertiveness")
	emotionalStability := traits.GetIntTrait("core_dimensions", "emotional_stability")

	adjustment := (float64(confidence+emotionalStability) / 20.0) - 0.5

	finalConfidence := baseConfidence + adjustment

	// Clamp to valid range
	if finalConfidence < 0.1 {
		finalConfidence = 0.1
	} else if finalConfidence > 1.0 {
		finalConfidence = 1.0
	}

	return finalConfidence
}

func (p *Persona) analyzeEmotionalTone(content string, traits *PersonalityTraits) string {
	contentLower := strings.ToLower(content)

	// Check for emotional indicators
	if strings.Contains(contentLower, "excited") || strings.Contains(contentLower, "amazing") {
		return "enthusiastic"
	}

	if strings.Contains(contentLower, "concerned") || strings.Contains(contentLower, "worried") {
		return "cautious"
	}

	if strings.Contains(contentLower, "confident") || strings.Contains(contentLower, "certain") {
		return "confident"
	}

	// Default based on personality
	optimism := traits.GetIntTrait("core_dimensions", "optimism")
	if optimism >= 7 {
		return "optimistic"
	} else if optimism <= 3 {
		return "realistic"
	}

	return "balanced"
}

func (p *Persona) analyzeTraitInfluence(traits *PersonalityTraits) map[string]float64 {
	influence := make(map[string]float64)

	// Calculate how much each major trait influenced the response
	influence["creativity"] = float64(traits.GetIntTrait("core_dimensions", "creativity")) / 10.0
	influence["analytical"] = float64(traits.GetIntTrait("core_dimensions", "analytical")) / 10.0
	influence["optimism"] = float64(traits.GetIntTrait("core_dimensions", "optimism")) / 10.0
	influence["risk_tolerance"] = float64(traits.GetIntTrait("core_dimensions", "risk_tolerance")) / 10.0
	influence["empathy"] = float64(traits.GetIntTrait("core_dimensions", "empathy")) / 10.0

	return influence
}

// Utility functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Additional methods continued in next part...
