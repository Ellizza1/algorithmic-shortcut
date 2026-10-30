package generator

import (
	"testing"

	"maze_visualizer/pkg/maze"
)

func countPassages(g *maze.Grid) int {
	passages := 0
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			c := g.Get(x, y)
			// Считаем проходы вправо и вниз (чтобы не дублировать)
			if x+1 < g.Width && (c.Walls&maze.WallRight) == 0 {
				passages++
			}
			if y+1 < g.Height && (c.Walls&maze.WallBottom) == 0 {
				passages++
			}
		}
	}
	return passages
}

func countReachable(g *maze.Grid) int {
	visited := make(map[[2]int]bool)
	queue := [][2]int{{0, 0}}
	visited[[2]int{0, 0}] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		c := g.Get(curr[0], curr[1])
		for _, n := range g.AccessibleNeighbors(c) {
			pos := [2]int{n.X, n.Y}
			if !visited[pos] {
				visited[pos] = true
				queue = append(queue, pos)
			}
		}
	}
	return len(visited)
}

func testSpanningTree(t *testing.T, gen Generator, w, h int) {
	grid := maze.NewGrid(w, h)
	eventsCount := 0

	gen.Generate(grid, func(e maze.Event) {
		eventsCount++
	})

	totalCells := w * h
	expectedEdges := totalCells - 1

	edges := countPassages(grid)
	if edges != expectedEdges {
		t.Fatalf("%s: expected %d carved edges in %dx%d spanning tree, got %d",
			gen.Name(), expectedEdges, w, h, edges)
	}

	reachable := countReachable(grid)
	if reachable != totalCells {
		t.Fatalf("%s: expected all %d cells to be reachable, got %d",
			gen.Name(), totalCells, reachable)
	}

	if eventsCount == 0 {
		t.Fatalf("%s: expected step events to be emitted", gen.Name())
	}
}

func TestGenerators(t *testing.T) {
	generators := []Generator{
		NewBacktracker(42),
		NewPrim(42),
		NewKruskal(42),
	}

	sizes := [][2]int{
		{5, 5},
		{10, 15},
		{21, 21},
	}

	for _, gen := range generators {
		for _, size := range sizes {
			w, h := size[0], size[1]
			t.Run(gen.Name(), func(t *testing.T) {
				testSpanningTree(t, gen, w, h)
			})
		}
	}
}
