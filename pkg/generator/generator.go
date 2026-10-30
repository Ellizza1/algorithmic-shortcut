package generator

import "maze_visualizer/pkg/maze"

// Generator интерфейс для алгоритмов построения лабиринта
type Generator interface {
	Name() string
	Generate(grid *maze.Grid, onStep func(maze.Event))
}
