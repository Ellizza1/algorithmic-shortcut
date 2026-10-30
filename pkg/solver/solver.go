package solver

import "maze_visualizer/pkg/maze"

// Solver интерфейс алгоритмов решения лабиринта
type Solver interface {
	Name() string
	Solve(grid *maze.Grid, start, goal [2]int, onStep func(maze.Event)) [][2]int
}

// ReconstructPath восстанавливает путь от цели к старту
func ReconstructPath(parents map[[2]int][2]int, start, goal [2]int) [][2]int {
	var path [][2]int
	curr := goal

	for {
		path = append([][2]int{curr}, path...) // prepend
		if curr == start {
			break
		}
		p, exists := parents[curr]
		if !exists {
			return nil // Путь не найден
		}
		curr = p
	}
	return path
}
