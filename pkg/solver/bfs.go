package solver

import (
	"time"

	"maze_visualizer/pkg/maze"
)

// BFS реализует поиск в ширину (радиальное волновое распространение)
type BFS struct{}

func NewBFS() *BFS {
	return &BFS{}
}

func (b *BFS) Name() string {
	return "Breadth-First Search (BFS)"
}

func (b *BFS) Solve(grid *maze.Grid, start, goal [2]int, onStep func(maze.Event)) [][2]int {
	startTime := time.Now()
	stepIdx := 0
	emit := func(e maze.Event) {
		stepIdx++
		e.StepIndex = stepIdx
		if onStep != nil {
			onStep(e)
		}
	}

	queue := [][2]int{start}
	visited := make(map[[2]int]bool)
	visited[start] = true
	parents := make(map[[2]int][2]int)

	emit(maze.Event{
		Type:    maze.EventVisit,
		X1:      start[0],
		Y1:      start[1],
		Message: "BFS: Старт волнового фронта",
	})

	found := false

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == goal {
			found = true
			break
		}

		c := grid.Get(curr[0], curr[1])
		neighbors := grid.AccessibleNeighbors(c)

		for _, n := range neighbors {
			pos := [2]int{n.X, n.Y}
			if !visited[pos] {
				visited[pos] = true
				parents[pos] = curr
				queue = append(queue, pos)

				emit(maze.Event{
					Type:      maze.EventVisit,
					X1:        n.X,
					Y1:        n.Y,
					Intensity: float64(len(visited)),
					Stats: map[string]any{
						"visited":    len(visited),
						"queue_size": len(queue),
					},
					Message: "BFS: раскрытие вершины",
				})
			}
		}
	}

	if !found {
		emit(maze.Event{
			Type:    maze.EventComplete,
			Message: "BFS: Путь к выходу не найден",
		})
		return nil
	}

	path := ReconstructPath(parents, start, goal)
	elapsed := time.Since(startTime)

	// Анимация финального пути
	for idx, pt := range path {
		emit(maze.Event{
			Type:      maze.EventPath,
			X1:        pt[0],
			Y1:        pt[1],
			Intensity: float64(idx) / float64(len(path)),
			Stats: map[string]any{
				"path_length": len(path),
				"visited":     len(visited),
			},
		})
	}

	emit(maze.Event{
		Type: maze.EventComplete,
		Path: path,
		Stats: map[string]any{
			"visited":     len(visited),
			"path_length": len(path),
			"time_ms":     elapsed.Milliseconds(),
		},
		Message: "BFS: Кратчайший путь успешно найден!",
	})

	return path
}
