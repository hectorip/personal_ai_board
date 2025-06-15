# Personal AI Board - System Design Document

## 1. System Overview

The Personal AI Board is a decision-support system that simulates advisory boards using AI personas to analyze ideas, projects, and decisions. The system enables users to create custom boards of AI personas, each with unique characteristics and persistent memory, to provide diverse perspectives on user-submitted topics.

## 2. Core Concepts

### 2.1 Persona
- **Definition**: A unique AI entity with defined personality traits, expertise areas, and behavioral patterns
- **Memory**: Persistent context storage unique to each persona instance
- **Variation**: Small random variations in responses to simulate natural human variability

### 2.2 Board
- **Definition**: A collection of personas assembled for specific analysis purposes
- **Types**: Template boards (pre-configured) and custom boards (user-created)
- **Composition**: 3-12 personas with complementary or diverse perspectives

### 2.3 Project
- **Definition**: A collection of related ideas with a common objective
- **Components**: Ideas, documents, files (images, videos, audio)
- **Knowledge Integration**: Documents processed into a knowledge graph for context

### 2.4 Analysis Modes
- **Discussion**: Interactive dialogue between personas about the topic
- **Simulation**: Predictive modeling of outcomes and scenarios
- **Analysis**: Structured breakdown and evaluation of components
- **Comparison**: Side-by-side evaluation against alternatives
- **Evaluation**: Scoring and ranking based on defined criteria
- **Prediction**: Forecasting potential outcomes and probabilities

## 3. System Architecture

### 3.1 Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│             Interfaces Layer            │
│  ┌─────────────┐  ┌─────────────────────┐│
│  │     CLI     │  │    Web (Gin+HTMX)  ││
│  └─────────────┘  └─────────────────────┘│
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│           Application Layer             │
│  ┌─────────────────────────────────────┐ │
│  │          Use Cases                  │ │
│  └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│            Domain Layer                 │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐│
│  │Persona  │ │ Board   │ │   Project   ││
│  └─────────┘ └─────────┘ └─────────────┘│
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│          Infrastructure Layer           │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐│
│  │SQLite DB│ │LLM Proxy│ │Vector Store ││
│  └─────────┘ └─────────┘ └─────────────┘│
└─────────────────────────────────────────┘
```

### 3.2 Core Components

#### Domain Entities
- **Persona**: Personality traits, memory context, response patterns
- **Board**: Persona collection, analysis configuration
- **Project**: Ideas, documents, metadata
- **Analysis**: Results, reports, insights
- **Session**: Conversation state, analysis progress

#### Repositories (Interfaces)
- **PersonaRepository**: CRUD operations for personas and their memory
- **BoardRepository**: Board management and composition
- **ProjectRepository**: Project and document management
- **AnalysisRepository**: Results storage and retrieval

#### Services (Domain Logic)
- **PersonaService**: Persona behavior and memory management
- **BoardService**: Board orchestration and analysis coordination
- **LLMService**: AI model communication and response processing
- **DocumentService**: File processing and knowledge graph integration

## 4. Data Model

### 4.1 Database Schema

```sql
-- Personas
CREATE TABLE personas (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    personality_traits JSON,
    expertise_areas JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Boards
CREATE TABLE boards (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_template BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Board-Persona relationships
CREATE TABLE board_personas (
    board_id TEXT REFERENCES boards(id),
    persona_id TEXT REFERENCES personas(id),
    role TEXT,
    PRIMARY KEY (board_id, persona_id)
);

-- Projects
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    metadata JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Ideas within projects
CREATE TABLE ideas (
    id TEXT PRIMARY KEY,
    project_id TEXT REFERENCES projects(id),
    title TEXT NOT NULL,
    content TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Analysis sessions
CREATE TABLE analysis_sessions (
    id TEXT PRIMARY KEY,
    project_id TEXT REFERENCES projects(id),
    board_id TEXT REFERENCES boards(id),
    mode TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    results JSON,
    created_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Persona memory (conversation history)
CREATE TABLE persona_memory (
    id TEXT PRIMARY KEY,
    persona_id TEXT REFERENCES personas(id),
    session_id TEXT REFERENCES analysis_sessions(id),
    context JSON,
    created_at TIMESTAMP
);

-- LLM interaction logs
CREATE TABLE llm_logs (
    id TEXT PRIMARY KEY,
    persona_id TEXT,
    session_id TEXT,
    prompt TEXT,
    response TEXT,
    model_name TEXT,
    tokens_used INTEGER,
    created_at TIMESTAMP
);
```

### 4.2 Knowledge Graph (Vector Store)
- **Documents**: Processed files with embeddings
- **Concepts**: Extracted entities and relationships
- **Context**: Project-specific knowledge for retrieval

## 5. Concurrency Design

### 5.1 Analysis Pipeline
```go
type AnalysisPipeline struct {
    PersonaPool    chan *Persona
    TaskQueue      chan AnalysisTask
    ResultCollector chan AnalysisResult
    ErrorHandler   chan error
}
```

### 5.2 Concurrent Patterns
- **Worker Pool**: Fixed number of goroutines processing persona interactions
- **Fan-out/Fan-in**: Distribute analysis tasks to personas, collect results
- **Pipeline**: Sequential processing stages with buffered channels
- **Circuit Breaker**: LLM failure handling and retry logic

## 6. LLM Integration Strategy

### 6.1 Abstraction Layer
```go
type LLMProvider interface {
    GenerateResponse(ctx context.Context, prompt string, config LLMConfig) (*Response, error)
    GetModels() []ModelInfo
    ValidateConfig(config LLMConfig) error
}
```

### 6.2 Supported Providers
- OpenAI GPT models
- Anthropic Claude
- Local models (Ollama)
- Custom API endpoints

### 6.3 Response Processing
- **Parsing**: Extract structured data from free-form responses
- **Validation**: Ensure responses meet expected format
- **Logging**: Complete interaction history for debugging

## 7. Memory Management

### 7.1 Persona Memory Strategy
- **Session Context**: Temporary memory for current analysis
- **Long-term Memory**: Persistent personality evolution
- **Memory Compression**: Summarize old conversations to manage size
- **Memory Retrieval**: Context-aware memory injection

### 7.2 Memory Storage
- JSON documents in SQLite for structured data
- Vector embeddings for semantic memory search
- File system for large context files

## 8. API Design

### 8.1 Core Use Cases
```go
// Board Management
CreateBoard(name, description string, personaIDs []string) (*Board, error)
GetBoard(id string) (*Board, error)
ListBoards() ([]*Board, error)

// Analysis Execution
RunAnalysis(projectID, boardID string, mode AnalysisMode) (*AnalysisSession, error)
GetAnalysisResults(sessionID string) (*AnalysisResult, error)

// Persona Management
CreatePersona(spec PersonaSpec) (*Persona, error)
UpdatePersonaMemory(personaID string, context MemoryContext) error
```

### 8.2 Web API Endpoints
```
POST   /api/v1/boards
GET    /api/v1/boards
GET    /api/v1/boards/{id}
POST   /api/v1/boards/{id}/analyze
GET    /api/v1/analysis/{sessionId}
POST   /api/v1/personas
GET    /api/v1/personas
POST   /api/v1/projects
GET    /api/v1/projects/{id}/documents
```

## 9. Security Considerations

### 9.1 Data Protection
- **Encryption**: Sensitive persona memory encrypted at rest
- **Access Control**: User-based isolation of boards and projects
- **Audit Trail**: Complete logging of all LLM interactions

### 9.2 LLM Safety
- **Input Sanitization**: Prevent prompt injection attacks
- **Output Filtering**: Remove potentially harmful content
- **Rate Limiting**: Prevent API abuse

## 10. Performance Requirements

### 10.1 Scalability Targets
- Support 100+ concurrent analysis sessions
- Handle boards with up to 12 personas
- Process documents up to 100MB
- Response time < 30 seconds for standard analysis

### 10.2 Resource Management
- **Connection Pooling**: Database and LLM API connections
- **Caching**: Frequently accessed persona data
- **Background Processing**: Document indexing and memory compression

## 11. Monitoring and Observability

### 11.1 Metrics
- Analysis session completion rates
- LLM API response times and error rates
- Resource utilization (CPU, memory, storage)
- User engagement patterns

### 11.2 Logging
- Structured logging with correlation IDs
- LLM interaction traces
- Error tracking and alerting
- Performance profiling data

## 12. Deployment Architecture

### 12.1 Single Binary Deployment
- Self-contained Go binary
- Embedded SQLite database
- Local file system storage
- Configuration via environment variables

### 12.2 Distributed Deployment (Future)
- PostgreSQL for primary data
- Redis for session state
- S3-compatible storage for documents
- Container orchestration ready