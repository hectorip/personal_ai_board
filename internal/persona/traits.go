package persona

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TraitValue represents a flexible trait value that can be numeric or string
type TraitValue interface{}

// TraitDefinition defines the structure and constraints of a personality trait
type TraitDefinition struct {
	Type        string      `json:"type"`
	Range       []int       `json:"range,omitempty"`
	Options     []string    `json:"options,omitempty"`
	Default     TraitValue  `json:"default"`
	Description string      `json:"description"`
}

// BaseTraitConfig represents the base personality trait configuration
type BaseTraitConfig struct {
	Version          string                         `json:"version"`
	Description      string                         `json:"description"`
	CoreDimensions   map[string]TraitDefinition     `json:"core_dimensions"`
	CommunicationStyle map[string]TraitDefinition   `json:"communication_style"`
	ExpertiseAreas   TraitDefinition                `json:"expertise_areas"`
	BiasesAndTendencies map[string]TraitDefinition  `json:"biases_and_tendencies"`
	ResponsePatterns map[string]TraitDefinition     `json:"response_patterns"`
	DecisionMaking   map[string]TraitDefinition     `json:"decision_making"`
	TemporalOrientation map[string]TraitDefinition  `json:"temporal_orientation"`
	LearningStyle    map[string]TraitDefinition     `json:"learning_style"`
	Constraints      TraitConstraints               `json:"constraints"`
}

// PersonalityConfig represents a specific personality configuration
type PersonalityConfig struct {
	Extends              string                    `json:"extends"`
	PersonaType          string                    `json:"persona_type"`
	Name                 string                    `json:"name"`
	Description          string                    `json:"description"`
	CoreDimensions       map[string]TraitValue     `json:"core_dimensions"`
	CommunicationStyle   map[string]TraitValue     `json:"communication_style"`
	ExpertiseAreas       []string                  `json:"expertise_areas"`
	BiasesAndTendencies  map[string]TraitValue     `json:"biases_and_tendencies"`
	ResponsePatterns     map[string]TraitValue     `json:"response_patterns"`
	DecisionMaking       map[string]TraitValue     `json:"decision_making"`
	TemporalOrientation  map[string]TraitValue     `json:"temporal_orientation"`
	LearningStyle        map[string]TraitValue     `json:"learning_style"`
	CustomTraits         map[string]TraitValue     `json:"custom_traits"`
	SpeakingPatterns     SpeakingPatterns          `json:"speaking_patterns"`
	EmotionalTriggers    EmotionalTriggers         `json:"emotional_triggers"`
	ResponseModifiers    map[string]TraitModifier  `json:"response_modifiers"`
}

// TraitConstraints defines validation rules for trait combinations
type TraitConstraints struct {
	TraitSumLimits struct {
		Description string           `json:"description"`
		Rules       []ConstraintRule `json:"rules"`
	} `json:"trait_sum_limits"`
}

// ConstraintRule defines a specific constraint on trait combinations
type ConstraintRule struct {
	Traits      []string `json:"traits"`
	MaxTotal    int      `json:"max_total"`
	MinTotal    int      `json:"min_total"`
	Description string   `json:"description"`
}

// SpeakingPatterns defines language patterns for a personality
type SpeakingPatterns struct {
	CommonPhrases     []string `json:"common_phrases"`
	AvoidsPhrases     []string `json:"avoids_phrases"`
	FavoriteAnalogies []string `json:"favorite_analogies,omitempty"`
	FavoriteFrameworks []string `json:"favorite_frameworks,omitempty"`
	CreativeTechniques []string `json:"creative_techniques,omitempty"`
}

// EmotionalTriggers defines what energizes and frustrates a personality
type EmotionalTriggers struct {
	Energizers   []string `json:"energizers"`
	Frustrations []string `json:"frustrations"`
}

// TraitModifier defines how traits change in specific contexts
type TraitModifier map[string]TraitValue

// PersonalityTraits represents the complete personality profile
type PersonalityTraits struct {
	Base                 *BaseTraitConfig
	Config               *PersonalityConfig
	CoreDimensions       map[string]TraitValue
	CommunicationStyle   map[string]TraitValue
	ExpertiseAreas       []string
	BiasesAndTendencies  map[string]TraitValue
	ResponsePatterns     map[string]TraitValue
	DecisionMaking       map[string]TraitValue
	TemporalOrientation  map[string]TraitValue
	LearningStyle        map[string]TraitValue
	CustomTraits         map[string]TraitValue
	SpeakingPatterns     SpeakingPatterns
	EmotionalTriggers    EmotionalTriggers
	ResponseModifiers    map[string]TraitModifier
}

// TraitLoader handles loading and validation of personality traits
type TraitLoader struct {
	configPath string
	baseConfig *BaseTraitConfig
}

// NewTraitLoader creates a new trait loader with the specified config path
func NewTraitLoader(configPath string) *TraitLoader {
	return &TraitLoader{
		configPath: configPath,
	}
}

// LoadBaseConfig loads the base trait configuration
func (tl *TraitLoader) LoadBaseConfig() error {
	basePath := filepath.Join(tl.configPath, "traits", "base.json")
	
	file, err := os.Open(basePath)
	if err != nil {
		return fmt.Errorf("failed to open base config: %w", err)
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read base config: %w", err)
	}
	
	var baseConfig BaseTraitConfig
	if err := json.Unmarshal(data, &baseConfig); err != nil {
		return fmt.Errorf("failed to parse base config: %w", err)
	}
	
	tl.baseConfig = &baseConfig
	return nil
}

// LoadPersonalityConfig loads and validates a specific personality configuration
func (tl *TraitLoader) LoadPersonalityConfig(filename string) (*PersonalityTraits, error) {
	if tl.baseConfig == nil {
		if err := tl.LoadBaseConfig(); err != nil {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
	}
	
	configPath := filepath.Join(tl.configPath, "traits", filename)
	
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open personality config: %w", err)
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read personality config: %w", err)
	}
	
	var config PersonalityConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse personality config: %w", err)
	}
	
	// Merge base config with personality config
	traits, err := tl.mergeTraits(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to merge traits: %w", err)
	}
	
	// Validate the merged traits
	if err := tl.validateTraits(traits); err != nil {
		return nil, fmt.Errorf("trait validation failed: %w", err)
	}
	
	return traits, nil
}

// mergeTraits combines base configuration with personality-specific overrides
func (tl *TraitLoader) mergeTraits(config *PersonalityConfig) (*PersonalityTraits, error) {
	traits := &PersonalityTraits{
		Base:                tl.baseConfig,
		Config:              config,
		CoreDimensions:      make(map[string]TraitValue),
		CommunicationStyle:  make(map[string]TraitValue),
		ExpertiseAreas:      config.ExpertiseAreas,
		BiasesAndTendencies: make(map[string]TraitValue),
		ResponsePatterns:    make(map[string]TraitValue),
		DecisionMaking:      make(map[string]TraitValue),
		TemporalOrientation: make(map[string]TraitValue),
		LearningStyle:       make(map[string]TraitValue),
		CustomTraits:        config.CustomTraits,
		SpeakingPatterns:    config.SpeakingPatterns,
		EmotionalTriggers:   config.EmotionalTriggers,
		ResponseModifiers:   config.ResponseModifiers,
	}
	
	// Merge core dimensions
	for key, def := range tl.baseConfig.CoreDimensions {
		if override, exists := config.CoreDimensions[key]; exists {
			traits.CoreDimensions[key] = override
		} else {
			traits.CoreDimensions[key] = def.Default
		}
	}
	
	// Merge communication style
	for key, def := range tl.baseConfig.CommunicationStyle {
		if override, exists := config.CommunicationStyle[key]; exists {
			traits.CommunicationStyle[key] = override
		} else {
			traits.CommunicationStyle[key] = def.Default
		}
	}
	
	// Merge other trait categories
	tl.mergeTraitCategory(tl.baseConfig.BiasesAndTendencies, config.BiasesAndTendencies, traits.BiasesAndTendencies)
	tl.mergeTraitCategory(tl.baseConfig.ResponsePatterns, config.ResponsePatterns, traits.ResponsePatterns)
	tl.mergeTraitCategory(tl.baseConfig.DecisionMaking, config.DecisionMaking, traits.DecisionMaking)
	tl.mergeTraitCategory(tl.baseConfig.TemporalOrientation, config.TemporalOrientation, traits.TemporalOrientation)
	tl.mergeTraitCategory(tl.baseConfig.LearningStyle, config.LearningStyle, traits.LearningStyle)
	
	return traits, nil
}

// mergeTraitCategory merges a specific category of traits
func (tl *TraitLoader) mergeTraitCategory(baseDefs map[string]TraitDefinition, configValues map[string]TraitValue, result map[string]TraitValue) {
	for key, def := range baseDefs {
		if override, exists := configValues[key]; exists {
			result[key] = override
		} else {
			result[key] = def.Default
		}
	}
}

// validateTraits validates that trait values conform to their definitions and constraints
func (tl *TraitLoader) validateTraits(traits *PersonalityTraits) error {
	// Validate core dimensions
	for key, value := range traits.CoreDimensions {
		if def, exists := tl.baseConfig.CoreDimensions[key]; exists {
			if err := tl.validateTraitValue(value, def, key); err != nil {
				return err
			}
		}
	}
	
	// Validate communication style
	for key, value := range traits.CommunicationStyle {
		if def, exists := tl.baseConfig.CommunicationStyle[key]; exists {
			if err := tl.validateTraitValue(value, def, key); err != nil {
				return err
			}
		}
	}
	
	// Validate constraint rules
	for _, rule := range tl.baseConfig.Constraints.TraitSumLimits.Rules {
		if err := tl.validateConstraintRule(traits, rule); err != nil {
			return err
		}
	}
	
	return nil
}

// validateTraitValue validates a single trait value against its definition
func (tl *TraitLoader) validateTraitValue(value TraitValue, def TraitDefinition, traitName string) error {
	switch def.Type {
	case "scale":
		if numValue, ok := value.(float64); ok {
			intValue := int(numValue)
			if len(def.Range) >= 2 && (intValue < def.Range[0] || intValue > def.Range[1]) {
				return fmt.Errorf("trait %s value %d is outside valid range %v", traitName, intValue, def.Range)
			}
		} else {
			return fmt.Errorf("trait %s should be numeric, got %T", traitName, value)
		}
	case "enum":
		if strValue, ok := value.(string); ok {
			valid := false
			for _, option := range def.Options {
				if strValue == option {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("trait %s value '%s' is not in valid options %v", traitName, strValue, def.Options)
			}
		} else {
			return fmt.Errorf("trait %s should be string, got %T", traitName, value)
		}
	}
	
	return nil
}

// validateConstraintRule validates a constraint rule across multiple traits
func (tl *TraitLoader) validateConstraintRule(traits *PersonalityTraits, rule ConstraintRule) error {
	total := 0
	for _, traitName := range rule.Traits {
		if value, exists := traits.CoreDimensions[traitName]; exists {
			if numValue, ok := value.(float64); ok {
				total += int(numValue)
			}
		}
	}
	
	if rule.MaxTotal > 0 && total > rule.MaxTotal {
		return fmt.Errorf("constraint violation: %s - total %d exceeds max %d", rule.Description, total, rule.MaxTotal)
	}
	
	if rule.MinTotal > 0 && total < rule.MinTotal {
		return fmt.Errorf("constraint violation: %s - total %d below min %d", rule.Description, total, rule.MinTotal)
	}
	
	return nil
}

// GetTraitValue retrieves a trait value with type safety
func (pt *PersonalityTraits) GetTraitValue(category, traitName string) (TraitValue, bool) {
	switch category {
	case "core_dimensions":
		value, exists := pt.CoreDimensions[traitName]
		return value, exists
	case "communication_style":
		value, exists := pt.CommunicationStyle[traitName]
		return value, exists
	case "biases_and_tendencies":
		value, exists := pt.BiasesAndTendencies[traitName]
		return value, exists
	case "response_patterns":
		value, exists := pt.ResponsePatterns[traitName]
		return value, exists
	case "decision_making":
		value, exists := pt.DecisionMaking[traitName]
		return value, exists
	case "temporal_orientation":
		value, exists := pt.TemporalOrientation[traitName]
		return value, exists
	case "learning_style":
		value, exists := pt.LearningStyle[traitName]
		return value, exists
	case "custom_traits":
		value, exists := pt.CustomTraits[traitName]
		return value, exists
	default:
		return nil, false
	}
}

// GetIntTrait safely retrieves an integer trait value
func (pt *PersonalityTraits) GetIntTrait(category, traitName string) int {
	if value, exists := pt.GetTraitValue(category, traitName); exists {
		if numValue, ok := value.(float64); ok {
			return int(numValue)
		}
	}
	return 5 // Default middle value
}

// GetStringTrait safely retrieves a string trait value
func (pt *PersonalityTraits) GetStringTrait(category, traitName string) string {
	if value, exists := pt.GetTraitValue(category, traitName); exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return "" // Default empty string
}

// ApplyContextModifier applies trait modifications based on context
func (pt *PersonalityTraits) ApplyContextModifier(context string) *PersonalityTraits {
	if modifier, exists := pt.ResponseModifiers[context]; exists {
		// Create a copy of traits with modifications
		modified := *pt
		
		// Apply modifications to relevant trait categories
		for traitName, newValue := range modifier {
			// Try to find and update the trait in appropriate category
			if _, exists := modified.CoreDimensions[traitName]; exists {
				modified.CoreDimensions[traitName] = newValue
			} else if _, exists := modified.CommunicationStyle[traitName]; exists {
				modified.CommunicationStyle[traitName] = newValue
			} else if _, exists := modified.ResponsePatterns[traitName]; exists {
				modified.ResponsePatterns[traitName] = newValue
			}
			// Add other categories as needed
		}
		
		return &modified
	}
	
	return pt
}