package engine

import (
	"encoding/json"
	"log"
	"sync"

	"maze_visualizer/pkg/generator"
	"maze_visualizer/pkg/maze"
	"maze_visualizer/pkg/solver"
)

// Command входящая команда от веб-клиента
type Command struct {
	Action    string `json:"action"`    // "generate", "solve", "reset", "init"
	Algorithm string `json:"algorithm"` // "backtracker", "prim", "kruskal", "astar", "bfs", "dijkstra"
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Seed      int64  `json:"seed"`
	StartX    int    `json:"start_x"`
	StartY    int    `json:"start_y"`
	GoalX     int    `json:"goal_x"`
	GoalY     int    `json:"goal_y"`
}

// Session управляет состоянием одного клиента и его симуляцией
type Session struct {
	mu        sync.Mutex
	grid      *maze.Grid
	sendEvent func(event maze.Event)
	cancel    chan struct{}
}

func NewSession(sendEvent func(event maze.Event)) *Session {
	return &Session{
		grid:      maze.NewGrid(25, 25),
		sendEvent: sendEvent,
	}
}

// HandleCommand обрабатывает действие пользователя
func (s *Session) HandleCommand(cmd Command) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Останавливаем предыдущую выполняющуюся операцию, если была
	if s.cancel != nil {
		close(s.cancel)
		s.cancel = nil
	}

	switch cmd.Action {
	case "init", "reset":
		w := cmd.Width
		h := cmd.Height
		if w < 5 {
			w = 25
		}
		if h < 5 {
			h = 25
		}
		s.grid = maze.NewGrid(w, h)
		s.sendEvent(maze.Event{
			Type: maze.EventReset,
			Stats: map[string]any{
				"width":  w,
				"height": h,
			},
			Message: "Сетка инициализирована",
		})

	case "generate":
		w := cmd.Width
		h := cmd.Height
		if w < 5 {
			w = 25
		}
		if h < 5 {
			h = 25
		}
		s.grid = maze.NewGrid(w, h)

		var gen generator.Generator
		switch cmd.Algorithm {
		case "prim":
			gen = generator.NewPrim(cmd.Seed)
		case "kruskal":
			gen = generator.NewKruskal(cmd.Seed)
		default:
			gen = generator.NewBacktracker(cmd.Seed)
		}

		s.sendEvent(maze.Event{
			Type: maze.EventReset,
			Stats: map[string]any{
				"width":     w,
				"height":    h,
				"generator": gen.Name(),
			},
			Message: "Старт генерации: " + gen.Name(),
		})

		gen.Generate(s.grid, func(e maze.Event) {
			s.sendEvent(e)
		})

	case "solve":
		if s.grid == nil {
			return
		}

		startX, startY := cmd.StartX, cmd.StartY
		goalX, goalY := cmd.GoalX, cmd.GoalY
		if !s.grid.InBounds(startX, startY) {
			startX, startY = 0, 0
		}
		if !s.grid.InBounds(goalX, goalY) {
			goalX, goalY = s.grid.Width-1, s.grid.Height-1
		}

		var solv solver.Solver
		switch cmd.Algorithm {
		case "bfs":
			solv = solver.NewBFS()
		case "dijkstra":
			solv = solver.NewDijkstra()
		default:
			solv = solver.NewAStar()
		}

		s.sendEvent(maze.Event{
			Type: maze.EventInfo,
			Stats: map[string]any{
				"solver":  solv.Name(),
				"start":   [2]int{startX, startY},
				"goal":    [2]int{goalX, goalY},
			},
			Message: "Старт поиска пути: " + solv.Name(),
		})

		solv.Solve(s.grid, [2]int{startX, startY}, [2]int{goalX, goalY}, func(e maze.Event) {
			s.sendEvent(e)
		})

	default:
		log.Printf("Неизвестное действие: %s", cmd.Action)
	}
}

// ParseCommand десериализует JSON команду
func ParseCommand(data []byte) (Command, error) {
	var cmd Command
	err := json.Unmarshal(data, &cmd)
	return cmd, err
}
