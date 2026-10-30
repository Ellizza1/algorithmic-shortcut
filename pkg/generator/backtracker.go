package generator

import (
	"math/rand"
	"time"

	"maze_visualizer/pkg/maze"
)

// Backtracker реализует рандомизированный поиск в глубину (DFS) со стеком
type Backtracker struct {
	rnd *rand.Rand
}

func NewBacktracker(seed int64) *Backtracker {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Backtracker{rnd: rand.New(rand.NewSource(seed))}
}

func (b *Backtracker) Name() string {
	return "Recursive Backtracker (DFS)"
}

func (b *Backtracker) Generate(grid *maze.Grid, onStep func(maze.Event)) {
	grid.Reset()

	stepIdx := 0
	emit := func(e maze.Event) {
		stepIdx++
		e.StepIndex = stepIdx
		if onStep != nil {
			onStep(e)
		}
	}

	// Стек для обхода в глубину
	stack := make([]*maze.Cell, 0, grid.Width*grid.Height)

	// Начинаем со стартовой клетки (0, 0)
	start := grid.Get(0, 0)
	start.Visited = true
	stack = append(stack, start)

	emit(maze.Event{
		Type:    maze.EventVisit,
		X1:      start.X,
		Y1:      start.Y,
		Message: "Начало генерации: стартовая ячейка помещена в стек DFS",
	})

	for len(stack) > 0 {
		curr := stack[len(stack)-1]

		// Ищем непосещенных соседей
		allNeighbors := grid.AllNeighbors(curr)
		unvisited := make([]*maze.Cell, 0, len(allNeighbors))
		for _, n := range allNeighbors {
			if !n.Visited {
				unvisited = append(unvisited, n)
			}
		}

		if len(unvisited) > 0 {
			// Выбираем случайного соседа
			next := unvisited[b.rnd.Intn(len(unvisited))]

			// Пробиваем стену между curr и next
			grid.CarvePassage(curr, next)
			next.Visited = true

			emit(maze.Event{
				Type:      maze.EventCarve,
				X1:        curr.X,
				Y1:        curr.Y,
				X2:        next.X,
				Y2:        next.Y,
				Intensity: float64(len(stack)),
				Stats: map[string]any{
					"stack_depth": len(stack) + 1,
				},
				Message: "Пробитие стены в глубину",
			})

			stack = append(stack, next)
		} else {
			// Тупик: выталкиваем из стека (backtrack)
			stack = stack[:len(stack)-1]

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				emit(maze.Event{
					Type: maze.EventBacktrack,
					X1:   curr.X,
					Y1:   curr.Y,
					X2:   parent.X,
					Y2:   parent.Y,
					Stats: map[string]any{
						"stack_depth": len(stack),
					},
					Message: "Откат назад из тупика",
				})
			}
		}
	}

	// Сбрасываем флаги Visited для дальнейшего решения лабиринта
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			grid.Cells[y][x].Visited = false
		}
	}

	emit(maze.Event{
		Type:    maze.EventComplete,
		Message: "Генерация лабиринта (DFS Backtracker) успешно завершена",
	})
}
