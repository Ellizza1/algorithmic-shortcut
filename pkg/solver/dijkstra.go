package solver

import (
	"container/heap"
	"math"
	"time"

	"maze_visualizer/pkg/maze"
)

// DijkstraItem элемент очереди с приоритетом
type DijkstraItem struct {
	Pos   [2]int
	Dist  float64
	Index int
}

type DijkstraPQ []*DijkstraItem

func (pq DijkstraPQ) Len() int           { return len(pq) }
func (pq DijkstraPQ) Less(i, j int) bool { return pq[i].Dist < pq[j].Dist }
func (pq DijkstraPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *DijkstraPQ) Push(x any) {
	n := len(*pq)
	item := x.(*DijkstraItem)
	item.Index = n
	*pq = append(*pq, item)
}
func (pq *DijkstraPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// Dijkstra реализует поиск кратчайшего пути с учетом весов
type Dijkstra struct{}

func NewDijkstra() *Dijkstra {
	return &Dijkstra{}
}

func (d *Dijkstra) Name() string {
	return "Dijkstra's Algorithm"
}

func (d *Dijkstra) Solve(grid *maze.Grid, start, goal [2]int, onStep func(maze.Event)) [][2]int {
	startTime := time.Now()
	stepIdx := 0
	emit := func(e maze.Event) {
		stepIdx++
		e.StepIndex = stepIdx
		if onStep != nil {
			onStep(e)
		}
	}

	dist := make(map[[2]int]float64)
	parents := make(map[[2]int][2]int)
	visited := make(map[[2]int]bool)

	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			dist[[2]int{x, y}] = math.Inf(1)
		}
	}
	dist[start] = 0

	pq := &DijkstraPQ{}
	heap.Init(pq)
	heap.Push(pq, &DijkstraItem{Pos: start, Dist: 0})

	found := false

	for pq.Len() > 0 {
		currItem := heap.Pop(pq).(*DijkstraItem)
		curr := currItem.Pos

		if visited[curr] {
			continue
		}
		visited[curr] = true

		emit(maze.Event{
			Type:      maze.EventVisit,
			X1:        curr[0],
			Y1:        curr[1],
			Intensity: currItem.Dist,
			Stats: map[string]any{
				"visited": len(visited),
				"pq_size": pq.Len(),
				"dist":    currItem.Dist,
			},
			Message: "Дейкстра: извлечение узла с минимальным расстоянием",
		})

		if curr == goal {
			found = true
			break
		}

		c := grid.Get(curr[0], curr[1])
		for _, n := range grid.AccessibleNeighbors(c) {
			neighborPos := [2]int{n.X, n.Y}
			if visited[neighborPos] {
				continue
			}

			newDist := dist[curr] + 1.0 // В решетке единичный вес шага
			if newDist < dist[neighborPos] {
				dist[neighborPos] = newDist
				parents[neighborPos] = curr
				heap.Push(pq, &DijkstraItem{Pos: neighborPos, Dist: newDist})

				emit(maze.Event{
					Type:      maze.EventFrontier,
					X1:        neighborPos[0],
					Y1:        neighborPos[1],
					Intensity: newDist,
					Stats: map[string]any{
						"visited": len(visited),
						"pq_size": pq.Len(),
					},
					Message: "Дейкстра: релаксация ребра",
				})
			}
		}
	}

	if !found {
		emit(maze.Event{
			Type:    maze.EventComplete,
			Message: "Дейкстра: Выход недостижим",
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
		Message: "Дейкстра: Кратчайший путь найден!",
	})

	return path
}
