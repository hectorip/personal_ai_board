# Personal AI Board - Simplified Go Architecture

## Architecture Philosophy

Following John Ousterhout's principles of deep modules and software simplicity:
- **Deep Modules**: Simple interfaces hiding complex implementation
- **Information Hiding**: Encapsulate complexity within modules
- **Minimize Abstractions**: Avoid unnecessary layers and indirection
- **Concurrent by Design**: Leveraging Go's goroutines and channels

## Directory Structure

```
personal-ai-board/
├── cmd/
│   ├── cli/main.go            # CLI entry point
│   └── web/main.go            # Web server entry point
├── internal/
│   ├── persona/               # Deep module: Persona management
│   │   ├── persona.go         # Core persona logic
│   │   ├── memory.go          # Memory management
│   │   ├── traits.go          # Personality trait system
│   │   └── storage.go         # Direct database access
│   ├── board/                 # Deep module: Board management
│   │   ├── board.go           # Board logic and composition
│   │   ├── templates.go       # Template management
│   │   └── storage.go         # Direct database access
│   ├── analysis/              # Deep module: Analysis engine
│   │   ├── engine.go          # Analysis orchestration
│   │   ├── modes.go           # Analysis mode implementations
│   │   ├── pipeline.go        # Concurrent processing
│   │   └── storage.go         # Direct database access
│   ├── project/               # Deep module: Project management
│   │   ├── project.go         # Project and ideas
│   │   ├── documents.go       # Document processing
│   │   └── storage.go         # Direct database access
│   ├── llm/                   # Deep module: LLM integration
│   │   ├── llm.go             # LLM orchestration
│   │   ├── providers/         # Provider implementations
│   │   │   ├── openai.go
│   │   │   ├── anthropic.go
│   │   │   └── ollama.go
│   │   └── logger.go          # LLM interaction logging
│   └── db/                    # Database utilities
│       ├── db.go              # Connection management
│       └── migrations.go      # Schema management
├── web/
│   ├── cli/                   # CLI interface (BubbleTea)
│   │   └── app.go
│   └── http/                  # HTTP interface (Gin + HTMX)
│       ├── server.go
│       ├── handlers.go
│       └── templates/
├── config/
│   ├── traits/                # Personality trait configurations
│   │   ├── base.json          # Base personality traits
│   │   ├── creative.json      # Creative personality template
│   │   ├── analytical.json    # Analytical personality template
│   │   └── visionary.json     # Visionary personality template
│   └── boards/                # Board templates
│       └── startup_advisors.json
├── go.mod
└── go.sum
```

## Deep Modules Design

### 1. Persona Module (`internal/persona/`)

**Simple Interface, Complex Implementation**

```go
// persona.go
package persona

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
)

// Simple public interface
type Persona struct {
    ID          string
    Name        string
    Description string
    Traits      PersonalityTraits
    memory      *Memory
    db          *sql.DB
}

type PersonalityTraits map[string]interface{}

type Memory struct {
    Context   map[string]interface{}
    History   []MemoryEntry
    shortTerm []MemoryEntry
    longTerm  []MemoryEntry
}

type MemoryEntry struct {
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Tags      []string  `json:"tags"`
    Weight    float64   `json:"weight"`
}

// Deep module - hides all complexity
func New(id, name, description string, traitsConfig string, db *sql.DB) (*Persona, error) {
    var traits PersonalityTraits
    if err := json.Unmarshal([]byte(traitsConfig), &traits); err != nil {
        return nil, err
    }

    p := &Persona{
        ID:          id,
        Name:        name,
        Description: description,
        Traits:      traits,
        memory:      &Memory{Context: make(map[string]interface{})},
        db:          db,
    }

    // Load existing memory from database
    p.loadMemory()
    return p, nil
}

func (p *Persona) Think(ctx context.Context, prompt string, context map[string]interface{}) (string, error) {
    // Complex internal logic hidden from caller:
    // 1. Retrieve relevant memories
    // 2. Apply personality traits to thinking process
    // 3. Generate contextual response
    // 4. Update memory
    // 5. Log interaction
    
    relevantMemories := p.memory.retrieveRelevant(prompt)
    personalizedPrompt := p.applyPersonality(prompt, context, relevantMemories)
    
    // This calls the LLM internally
    response, err := p.generateResponse(ctx, personalizedPrompt)
    if err != nil {
        return "", err
    }
    
    p.updateMemory(prompt, response, context)
    p.saveMemory()
    
    return response, nil
}

func (p *Persona) UpdateTraits(newTraits string) error {
    var traits PersonalityTraits
    if err := json.Unmarshal([]byte(newTraits), &traits); err != nil {
        return err
    }
    p.Traits = traits
    return p.saveTraits()
}

// All complex implementation details hidden inside the module
func (p *Persona) applyPersonality(prompt string, context map[string]interface{}, memories []MemoryEntry) string {
    // Complex personality application logic
    // Uses traits configuration to modify prompt
    // Applies communication style, expertise, biases, etc.
}

func (p *Persona) loadMemory() error {
    // Direct database access - no repository abstraction
    query := `SELECT context, history FROM persona_memory WHERE persona_id = ?`
    // Implementation details hidden
}

func (p *Persona) saveMemory() error {
    // Direct database save
    // Memory compression logic
    // Cleanup old memories
}
```

### 2. Board Module (`internal/board/`)

```go
// board.go
package board

import (
    "context"
    "database/sql"
    "personal-ai-board/internal/persona"
)

type Board struct {
    ID          string
    Name        string
    Description string
    personas    []*persona.Persona
    db          *sql.DB
}

// Simple interface - complex implementation hidden
func New(id, name, description string, db *sql.DB) *Board {
    return &Board{
        ID:          id,
        Name:        name,
        Description: description,
        personas:    make([]*persona.Persona, 0),
        db:          db,
    }
}

func (b *Board) AddPersona(p *persona.Persona) error {
    // Validation logic hidden
    // Database persistence hidden
    b.personas = append(b.personas, p)
    return b.save()
}

func (b *Board) Discuss(ctx context.Context, topic string, context map[string]interface{}) (*Discussion, error) {
    // Complex orchestration logic hidden:
    // 1. Coordinate multiple personas
    // 2. Manage turn-taking and interaction
    // 3. Handle concurrent processing
    // 4. Synthesize results
    
    discussion := &Discussion{
        Topic:   topic,
        Context: context,
        Turns:   make([]Turn, 0),
    }
    
    // Concurrent persona processing
    return b.orchestrateDiscussion(ctx, discussion)
}

type Discussion struct {
    Topic   string
    Context map[string]interface{}
    Turns   []Turn
    Summary string
}

type Turn struct {
    PersonaID string
    Response  string
    Timestamp time.Time
}
```

### 3. Analysis Module (`internal/analysis/`)

```go
// engine.go
package analysis

import (
    "context"
    "personal-ai-board/internal/board"
    "personal-ai-board/internal/project"
)

type Engine struct {
    db *sql.DB
}

type AnalysisRequest struct {
    Project *project.Project
    Board   *board.Board
    Mode    string
    Context map[string]interface{}
}

type AnalysisResult struct {
    ID             string
    Summary        string
    Recommendations []string
    PersonaInsights map[string]string
    Confidence     float64
    CompletedAt    time.Time
}

// Simple interface - all complexity hidden
func (e *Engine) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
    // Complex orchestration hidden:
    // 1. Determine analysis strategy based on mode
    // 2. Coordinate board personas
    // 3. Process project documents
    // 4. Run concurrent analysis pipeline
    // 5. Synthesize results
    // 6. Generate recommendations
    
    pipeline := e.createPipeline(req.Mode)
    return pipeline.Execute(ctx, req)
}

func NewEngine(db *sql.DB) *Engine {
    return &Engine{db: db}
}

// All implementation details hidden in private methods
func (e *Engine) createPipeline(mode string) analysisPipeline {
    // Complex pipeline creation logic
}
```

## Personality Traits Configuration System

### Base Traits Configuration (`config/traits/base.json`)

```json
{
  "version": "1.0",
  "core_dimensions": {
    "creativity": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Level of creative thinking and innovation"
    },
    "analytical": {
      "type": "scale", 
      "range": [1, 10],
      "default": 5,
      "description": "Depth of analytical and logical reasoning"
    },
    "optimism": {
      "type": "scale",
      "range": [1, 10], 
      "default": 5,
      "description": "Tendency toward positive or negative outlook"
    },
    "risk_tolerance": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Comfort with uncertainty and risk-taking"
    },
    "detail_orientation": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Focus on details vs big picture thinking"
    }
  },
  "communication_style": {
    "formality": {
      "type": "enum",
      "options": ["very_formal", "formal", "neutral", "casual", "very_casual"],
      "default": "neutral"
    },
    "directness": {
      "type": "enum", 
      "options": ["very_direct", "direct", "diplomatic", "indirect", "very_indirect"],
      "default": "direct"
    },
    "verbosity": {
      "type": "enum",
      "options": ["terse", "concise", "balanced", "detailed", "verbose"],
      "default": "balanced"
    },
    "emotion_level": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Emotional expressiveness in communication"
    }
  },
  "expertise_areas": {
    "type": "array",
    "items": "string",
    "description": "Areas of domain expertise",
    "examples": ["technology", "business", "psychology", "finance", "design"]
  },
  "biases_and_tendencies": {
    "confirmation_bias": {
      "type": "scale",
      "range": [1, 10],
      "default": 5
    },
    "optimism_bias": {
      "type": "scale", 
      "range": [1, 10],
      "default": 5
    },
    "authority_deference": {
      "type": "scale",
      "range": [1, 10], 
      "default": 5
    }
  },
  "response_patterns": {
    "question_tendency": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Likelihood to ask questions vs make statements"
    },
    "example_usage": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Tendency to use examples and analogies"
    },
    "contrarian_level": {
      "type": "scale",
      "range": [1, 10],
      "default": 5,
      "description": "Tendency to challenge ideas vs agree"
    }
  }
}
```

### Specific Personality Template (`config/traits/visionary.json`)

```json
{
  "extends": "base",
  "persona_type": "visionary_tech_leader",
  "core_dimensions": {
    "creativity": 9,
    "analytical": 8,
    "optimism": 8,
    "risk_tolerance": 9,
    "detail_orientation": 3
  },
  "communication_style": {
    "formality": "casual",
    "directness": "very_direct", 
    "verbosity": "concise",
    "emotion_level": 7
  },
  "expertise_areas": [
    "technology",
    "innovation",
    "product_strategy", 
    "market_disruption",
    "leadership"
  ],
  "biases_and_tendencies": {
    "confirmation_bias": 6,
    "optimism_bias": 8,
    "authority_deference": 2
  },
  "response_patterns": {
    "question_tendency": 7,
    "example_usage": 8,
    "contrarian_level": 7
  },
  "custom_traits": {
    "humor_level": 8,
    "storytelling": 9,
    "future_focus": 10,
    "disruption_mindset": 9,
    "impatience_with_status_quo": 8
  },
  "speaking_patterns": {
    "common_phrases": [
      "Think different",
      "What if we...",
      "The future is...",
      "Let me tell you a story"
    ],
    "avoids_phrases": [
      "That's impossible",
      "It's never been done",
      "Let's be realistic"
    ]
  },
  "decision_making": {
    "speed": 9,
    "gut_vs_data": 7,
    "consensus_seeking": 3
  }
}
```

## Module Usage Examples

### Simple API Usage

```go
// Using the deep modules - complexity is hidden
func main() {
    db := db.Connect("board.db")
    
    // Load personality from config
    traitsConfig, _ := os.ReadFile("config/traits/visionary.json")
    steve := persona.New("steve", "Steve Jobs", "Tech visionary", string(traitsConfig), db)
    
    // Create board
    board := board.New("advisors", "Startup Advisory Board", "Tech startup advisors", db)
    board.AddPersona(steve)
    
    // Run analysis - all complexity hidden
    engine := analysis.NewEngine(db)
    result, err := engine.Analyze(ctx, analysis.AnalysisRequest{
        Project: project,
        Board:   board, 
        Mode:    "discussion",
        Context: map[string]interface{}{"focus": "product_strategy"},
    })
}
```
