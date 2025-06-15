# Personal AI board

This projecto objetive is to create a personal AI board to help the user make decisions and analyze topics.

## Features

- Assemble a board of simulated personas, for example you can include in your board a man that is a visionary
in technology, is ruthless, and has a strong sense of humor. He is also very creative and innovative. Each personality
is a unique individual with their own set of skills and abilities. Each persona is an instance of a personality, has
its own memory and small random variations when the person is instantiated.
- There are some template boards, pre-designed with a set custom personalities.
- You can also create your own board by assembling your own personas.
- You can manage your ideas and projects (a project is a set of ideas that are related to each other and have a common goal) and run against several boards. A project can also include documents and files like images, videos, and audio files, these are processed and uploaded to a knowledge graph to be used in the analysis.
- When you run an idea against a board it writes a report with the results of the analysis.
- There are several modes:
    - **Discussion**: The board will discuss the idea with the personas and provide a report with the results of the discussion.
    - **Simulation**: Run the idea against the board and get a report with the results of the analysis.
    - **Analysis**: Analyze the idea against the board and get a report with the results of the analysis.
    - **Comparison**: Compare the idea against the board and get a report with the results of the analysis.
    - **Evaluation**: Evaluate the idea against the board and get a report with the results of the analysis.
    - **Prediction**: Predict the outcome of the idea against the board and get a report with the results of the analysis.

## Technology

- This is a Go project
- Each person has is unique memory, this could be a folder with a unique identifier.
- The project uses an LLM to simulate the personas. Make easy to switch between different LLM's.
- The core is a pure Go Library exposing functions, using sqlite as a backend database, but making easy to switch to other relational databases.
- The project runs analysis concurrently, using goroutines and channels
- The core library logs every response from the LLM.
- For processing the documents, we can use Weaviate, but we can evaluate Rust's Cocoindex
- The project has a CLI interface using a nice CLI library such as bubbletea
- The project has a web interface using a nice web framework such as Gin
- The project has a web interface using tailwind htmx

## Pasos del proyecto para Claude

1. Analiza este documento y mejoralo, crea un nuevo documento de diseño del sistema y si hay dudas agregalas.
2. Crea un documento de arquitectura, es muy importante que respetes las reglas de diseño: un core limpio en Go, diseño modular y comunicación con el exterior (CLI y HTTP) como modulos separados
3. Investiga las mejores prácitas de diseño de sistemas con IA
4. Programa el diseño
5. Crea un conjunto de datos de pruebas
