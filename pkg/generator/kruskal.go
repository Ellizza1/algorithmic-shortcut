package generator

import (
	"math/rand"
	"time"

	"maze_visualizer/pkg/maze"
)

// DSU структура системы непересекающихся множеств с эвристиками сжатия пути и ранга
type DSU struct {
	parent []int
	rank   []int
	count  int
}

func NewDSU(size int) *DSU {
	parent := make([]int, size)
	rank := make([]int, size)
	for i := 0; i < size; i++ {
		parent[i] = i
		rank[i] = 0
	}
	return &DSU{
		parent: parent,
		rank:   rank,
		count:  size,
	}
}

func (d *DSU) Find(x int) int {
	if d.parent[x] != x {
		d.parent[x] = d.Find(d.parent[x]) // Path compression
	}
	return d.parent[x]
}

func (d *DSU) Union(x, y int) bool {
	rootX := d.Find(x)
	rootY := d.Find(y)
	if rootX == rootY {
		return false // Уже в одном множестве (цикл!)
	}

	// Union by rank
	if d.rank[rootX] < d.rank[rootY] {
		d.parent[rootX] = rootY
	} else if d.rank[rootX] > d.rank[rootY] {
		d.parent[rootY] = rootX
	} else {
		d.parent[rootY] = rootX
		d.rank[rootX]++
	}
	d.count--
	return true
}

type KruskalEdge struct {
	C1 *maze.Cell
	C2 *maze.Cell
}

// Kruskal реализует рандомизированный алгоритм Краскала (глобальное слияние компонент)
type Kruskal struct {
	rnd *rand.Rand
}

func NewKruskal(seed int64) *Kruskal {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Kruskal{rnd: rand.New(rand.NewSource(seed))}
}

func (k *Kruskal) Name() string {
	return "Randomized Kruskal's Algorithm"
}

func (k *Kruskal) Generate(grid *maze.Grid, onStep func(maze.Event)) {
	grid.Reset()

	stepIdx := 0
	emit := func(e maze.Event) {
		stepIdx++
		e.StepIndex = stepIdx
		if onStep != nil {
			onStep(e)
		}
	}

	totalCells := grid.Width * grid.Height
	dsu := NewDSU(totalCells)

	cellID := func(c *maze.Cell) int {
		return c.Y*grid.Width + c.X
	}

	// Собираем все внутренние перегородки (ребра графа сетки)
	var edges []KruskalEdge
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			c := grid.Get(x, y)
			// Горизонтальное ребро направо
			if x+1 < grid.Width {
				edges = append(edges, KruskalEdge{C1: c, C2: grid.Get(x+1, y)})
			}
			// Вертикальное ребро вниз
			if y+1 < grid.Height {
				edges = append(edges, KruskalEdge{C1: c, C2: grid.Get(x, y+1)})
			}
		}
	}

	// Равномерно перемешиваем ребра (Fisher-Yates)
	k.rnd.Shuffle(len(edges), func(i, j int) {
		edges[i], edges[j] = edges[j], edges[i]
	})

	emit(maze.Event{
		Type:    maze.EventInfo,
		Message: "Начало алгоритма Краскала: ребра перемешаны, DSU инициализирована",
		Stats: map[string]any{
			"components": dsu.count,
			"edges":      len(edges),
		},
	})

	// Перебираем перемешанные перегородки
	for _, edge := range edges {
		id1 := cellID(edge.C1)
		id2 := cellID(edge.C2)

		// Если ячейки в разных компонентах — разрушаем стену и объединяем деревья
		if dsu.Union(id1, id2) {
			grid.CarvePassage(edge.C1, edge.C2)

			emit(maze.Event{
				Type:      maze.EventCarve,
				X1:        edge.C1.X,
				Y1:        edge.C1.Y,
				X2:        edge.C2.X,
				Y2:        edge.C2.Y,
				Intensity: float64(dsu.count) / float64(totalCells),
				Stats: map[string]any{
					"components": dsu.count,
				},
				Message: "Слияние двух компонент связности",
			})

			// Если осталась ровно 1 компонента — остовное дерево построено
			if dsu.count == 1 {
				break
			}
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
		Message: "Генерация лабиринта (алгоритм Краскала) успешно завершена",
	})
}
