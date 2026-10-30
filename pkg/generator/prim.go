package generator

import (
	"math/rand"
	"time"

	"maze_visualizer/pkg/maze"
)

type PrimWall struct {
	From *maze.Cell
	To   *maze.Cell
}

// Prim реализует рандомизированный алгоритм Прима (минимальный остов от очага)
type Prim struct {
	rnd *rand.Rand
}

func NewPrim(seed int64) *Prim {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Prim{rnd: rand.New(rand.NewSource(seed))}
}

func (p *Prim) Name() string {
	return "Randomized Prim's Algorithm"
}

func (p *Prim) Generate(grid *maze.Grid, onStep func(maze.Event)) {
	grid.Reset()

	stepIdx := 0
	emit := func(e maze.Event) {
		stepIdx++
		e.StepIndex = stepIdx
		if onStep != nil {
			onStep(e)
		}
	}

	inMaze := make([][]bool, grid.Height)
	for y := 0; y < grid.Height; y++ {
		inMaze[y] = make([]bool, grid.Width)
	}

	// Начинаем с центра или (0,0)
	startX, startY := grid.Width/2, grid.Height/2
	start := grid.Get(startX, startY)
	inMaze[startY][startX] = true

	emit(maze.Event{
		Type:    maze.EventVisit,
		X1:      start.X,
		Y1:      start.Y,
		Message: "Старт алгоритма Прима: ядро остовного дерева заложено",
	})

	// Список стен фронтира
	var frontier []PrimWall

	addFrontier := func(c *maze.Cell) {
		for _, n := range grid.AllNeighbors(c) {
			if !inMaze[n.Y][n.X] {
				frontier = append(frontier, PrimWall{From: c, To: n})
				emit(maze.Event{
					Type: maze.EventFrontier,
					X1:   n.X,
					Y1:   n.Y,
					Stats: map[string]any{
						"frontier_size": len(frontier),
					},
					Message: "Добавление стены во фронтир",
				})
			}
		}
	}

	addFrontier(start)

	for len(frontier) > 0 {
		// Случайно выбираем стену из фронтира
		idx := p.rnd.Intn(len(frontier))
		wall := frontier[idx]

		// Быстрое удаление за O(1) перестановкой с последним элементом
		frontier[idx] = frontier[len(frontier)-1]
		frontier = frontier[:len(frontier)-1]

		if !inMaze[wall.To.Y][wall.To.X] {
			// Пробиваем стену
			grid.CarvePassage(wall.From, wall.To)
			inMaze[wall.To.Y][wall.To.X] = true

			emit(maze.Event{
				Type:      maze.EventCarve,
				X1:        wall.From.X,
				Y1:        wall.From.Y,
				X2:        wall.To.X,
				Y2:        wall.To.Y,
				Intensity: float64(len(frontier)),
				Stats: map[string]any{
					"frontier_size": len(frontier),
				},
				Message: "Пробитие стены во фронтире Прима",
			})

			// Добавляем новые стены в фронтир
			addFrontier(wall.To)
		}
	}

	// Сброс флагов
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			grid.Cells[y][x].Visited = false
		}
	}

	emit(maze.Event{
		Type:    maze.EventComplete,
		Message: "Генерация лабиринта (алгоритм Прима) успешно завершена",
	})
}
