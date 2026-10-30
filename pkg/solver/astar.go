package solver

import (
	"container/heap"
	"math"
	"time"

	"maze_visualizer/pkg/maze"
)

// AStarItem элемент приоритетной очереди для A*
type AStarItem struct {
	Pos   [2]int
	G     float64 // Точная стоимость пути от старта
	F     float64 // f = g + h
	Index int
}

type AStarPQ []*AStarItem

func (pq AStarPQ) Len() int { return len(pq) }
func (pq AStarPQ) Less(i, j int) bool {
	if pq[i].F == pq[j].F {
		return pq[i].G > pq[j].G // При равенстве f предпочитаем узел ближе к цели (больший g)
	}
	return pq[i].F < pq[j].F
}
func (pq AStarPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *AStarPQ) Push(x any) {
	n := len(*pq)
	item := x.(*AStarItem)
	item.Index = n
	*pq = append(*pq, item)
}
func (pq *AStarPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// ManhattanDistance вычисляет манхэттенское расстояние |dx| + |dy|
func ManhattanDistance(a, b [2]int) float64 {
	dx := math.Abs(float64(a[0] - b[0]))
	dy := math.Abs(float64(a[1] - b[1]))
	return dx + dy
}

// AStar реализует алгоритм A* с эвристикой Манхэттена
type AStar struct{}

func NewAStar() *AStar {
	return &AStar{}
}

func (a *AStar) Name() string {
	return "A* Pathfinding (Manhattan Heuristic)"
}

func (a *AStar) Solve(grid *maze.Grid, start, goal [2]int, onStep func(maze.Event)) [][2]int {
	startTime := time.Now()
	stepIdx := 0
	emit := func(e maze.Event) {
		stepIdx++
		e.StepIndex = stepIdx
		if onStep != nil {
			onStep(e)
		}
	}

	gScore := make(map[[2]int]float64)
	parents := make(map[[2]int][2]int)
	closedSet := make(map[[2]int]bool)

	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			gScore[[2]int{x, y}] = math.Inf(1)
		}
	}
	gScore[start] = 0

	pq := &AStarPQ{}
	heap.Init(pq)
	heap.Push(pq, &AStarItem{
		Pos: start,
		G:   0,
		F:   ManhattanDistance(start, goal),
	})

	found := false

	for pq.Len() > 0 {
		currItem := heap.Pop(pq).(*AStarItem)
		curr := currItem.Pos

		if closedSet[curr] {
			continue
		}
		closedSet[curr] = true

		emit(maze.Event{
			Type:      maze.EventVisit,
			X1:        curr[0],
			Y1:        curr[1],
			Intensity: currItem.F,
			Stats: map[string]any{
				"visited": len(closedSet),
				"pq_size": pq.Len(),
				"f_score": currItem.F,
				"g_score": currItem.G,
			},
			Message: "A*: направленное раскрытие узла",
		})

		if curr == goal {
			found = true
			break
		}

		c := grid.Get(curr[0], curr[1])
		for _, n := range grid.AccessibleNeighbors(c) {
			neighborPos := [2]int{n.X, n.Y}
			if closedSet[neighborPos] {
				continue
			}

			tentativeG := gScore[curr] + 1.0
			if tentativeG < gScore[neighborPos] {
				parents[neighborPos] = curr
				gScore[neighborPos] = tentativeG
				f := tentativeG + ManhattanDistance(neighborPos, goal)

				heap.Push(pq, &AStarItem{
					Pos: neighborPos,
					G:   tentativeG,
					F:   f,
				})

				emit(maze.Event{
					Type:      maze.EventFrontier,
					X1:        neighborPos[0],
					Y1:        neighborPos[1],
					Intensity: f,
					Stats: map[string]any{
						"visited": len(closedSet),
						"pq_size": pq.Len(),
						"f_score": f,
					},
					Message: "A*: добавление узла в открытый список с эвристикой",
				})
			}
		}
	}

	if !found {
		emit(maze.Event{
			Type:    maze.EventComplete,
			Message: "A*: Путь к выходу не найден",
		})
		return nil
	}

	path := ReconstructPath(parents, start, goal)
	elapsed := time.Since(startTime)

	for idx, pt := range path {
		emit(maze.Event{
			Type:      maze.EventPath,
			X1:        pt[0],
			Y1:        pt[1],
			Intensity: float64(idx) / float64(len(path)),
			Stats: map[string]any{
				"path_length": len(path),
				"visited":     len(closedSet),
			},
		})
	}

	emit(maze.Event{
		Type: maze.EventComplete,
		Path: path,
		Stats: map[string]any{
			"visited":     len(closedSet),
			"path_length": len(path),
			"time_ms":     elapsed.Milliseconds(),
		},
		Message: "A*: Кратчайший путь найден с минимальным раскрытием узлов!",
	})

	return path
}
