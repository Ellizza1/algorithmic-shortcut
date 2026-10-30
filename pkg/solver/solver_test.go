package solver

import (
	"testing"

	"maze_visualizer/pkg/generator"
	"maze_visualizer/pkg/maze"
)

func validatePath(t *testing.T, grid *maze.Grid, path [][2]int, start, goal [2]int) {
	if len(path) == 0 {
		t.Fatal("path is empty")
	}
	if path[0] != start {
		t.Fatalf("path should start at %v, got %v", start, path[0])
	}
	if path[len(path)-1] != goal {
		t.Fatalf("path should end at %v, got %v", goal, path[len(path)-1])
	}

	for i := 0; i < len(path)-1; i++ {
		c1 := grid.Get(path[i][0], path[i][1])
		c2 := grid.Get(path[i+1][0], path[i+1][1])
		if !grid.AreConnected(c1, c2) {
			t.Fatalf("wall between step %d (%v) and step %d (%v)", i, path[i], i+1, path[i+1])
		}
	}
}

func TestSolversEquivalenceAndOptimality(t *testing.T) {
	w, h := 15, 15
	grid := maze.NewGrid(w, h)
	gen := generator.NewBacktracker(12345)
	gen.Generate(grid, nil)

	start := [2]int{0, 0}
	goal := [2]int{w - 1, h - 1}

	bfs := NewBFS()
	dijkstra := NewDijkstra()
	astar := NewAStar()

	var bfsVisited, dijkstraVisited, astarVisited int

	pathBFS := bfs.Solve(grid, start, goal, func(e maze.Event) {
		if e.Type == maze.EventVisit {
			bfsVisited++
		}
	})
	validatePath(t, grid, pathBFS, start, goal)

	pathDijkstra := dijkstra.Solve(grid, start, goal, func(e maze.Event) {
		if e.Type == maze.EventVisit {
			dijkstraVisited++
		}
	})
	validatePath(t, grid, pathDijkstra, start, goal)

	pathAStar := astar.Solve(grid, start, goal, func(e maze.Event) {
		if e.Type == maze.EventVisit {
			astarVisited++
		}
	})
	validatePath(t, grid, pathAStar, start, goal)

	// В остовном дереве путь единственный, поэтому длины должны строго совпадать
	if len(pathBFS) != len(pathDijkstra) || len(pathBFS) != len(pathAStar) {
		t.Fatalf("path lengths mismatch: BFS=%d, Dijkstra=%d, A*=%d",
			len(pathBFS), len(pathDijkstra), len(pathAStar))
	}

	t.Logf("Path length: %d", len(pathBFS))
	t.Logf("Nodes visited: BFS=%d, Dijkstra=%d, A*=%d",
		bfsVisited, dijkstraVisited, astarVisited)
}
