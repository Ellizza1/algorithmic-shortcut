package maze

// Битовые маски для 4 стен каждой ячейки
const (
	WallTop    uint8 = 1 << 0 // 1  (Север)
	WallRight  uint8 = 1 << 1 // 2  (Восток)
	WallBottom uint8 = 1 << 2 // 4  (Юг)
	WallLeft   uint8 = 1 << 3 // 8  (Запад)
	WallAll    uint8 = WallTop | WallRight | WallBottom | WallLeft // 15 (все 4 стены)
)

// Cell представляет одну ячейку координатной сетки
type Cell struct {
	X       int   `json:"x"`
	Y       int   `json:"y"`
	Walls   uint8 `json:"walls"`   // Битовая маска оставшихся стен
	Visited bool  `json:"visited"` // Флаг посещения для алгоритмов
}

// Direction представляет смещение и битовую стену
type Direction struct {
	DX   int
	DY   int
	Wall uint8
}

// Directions - 4 ортогональных направления
var Directions = []Direction{
	{DX: 0, DY: -1, Wall: WallTop},
	{DX: 1, DY: 0, Wall: WallRight},
	{DX: 0, DY: 1, Wall: WallBottom},
	{DX: -1, DY: 0, Wall: WallLeft},
}

// OppositeWall возвращает противоположную стену
func OppositeWall(wall uint8) uint8 {
	switch wall {
	case WallTop:
		return WallBottom
	case WallRight:
		return WallLeft
	case WallBottom:
		return WallTop
	case WallLeft:
		return WallRight
	default:
		return 0
	}
}

// Grid представляет двумерное поле лабиринта
type Grid struct {
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Cells  [][]Cell `json:"cells"`
}

// NewGrid инициализирует поле с закрытыми со всех сторон ячейками
func NewGrid(width, height int) *Grid {
	g := &Grid{
		Width:  width,
		Height: height,
		Cells:  make([][]Cell, height),
	}
	for y := 0; y < height; y++ {
		g.Cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			g.Cells[y][x] = Cell{
				X:       x,
				Y:       y,
				Walls:   WallAll,
				Visited: false,
			}
		}
	}
	return g
}

// InBounds проверяет, лежат ли координаты внутри поля
func (g *Grid) InBounds(x, y int) bool {
	return x >= 0 && x < g.Width && y >= 0 && y < g.Height
}

// Get возвращает указатель на ячейку
func (g *Grid) Get(x, y int) *Cell {
	if !g.InBounds(x, y) {
		return nil
	}
	return &g.Cells[y][x]
}

// CarvePassage симметрично удаляет стену между двумя смежными ячейками
func (g *Grid) CarvePassage(c1, c2 *Cell) bool {
	if c1 == nil || c2 == nil {
		return false
	}
	dx := c2.X - c1.X
	dy := c2.Y - c1.Y

	for _, d := range Directions {
		if d.DX == dx && d.DY == dy {
			c1.Walls &= ^d.Wall
			c2.Walls &= ^OppositeWall(d.Wall)
			return true
		}
	}
	return false
}

// AreConnected проверяет, пробита ли стена между двумя смежными ячейками
func (g *Grid) AreConnected(c1, c2 *Cell) bool {
	if c1 == nil || c2 == nil {
		return false
	}
	dx := c2.X - c1.X
	dy := c2.Y - c1.Y

	for _, d := range Directions {
		if d.DX == dx && d.DY == dy {
			return (c1.Walls & d.Wall) == 0
		}
	}
	return false
}

// AllNeighbors возвращает всех соседей ячейки в пределах границ
func (g *Grid) AllNeighbors(c *Cell) []*Cell {
	var list []*Cell
	for _, d := range Directions {
		nx, ny := c.X+d.DX, c.Y+d.DY
		if g.InBounds(nx, ny) {
			list = append(list, g.Get(nx, ny))
		}
	}
	return list
}

// AccessibleNeighbors возвращает соседей, путь к которым свободен (нет стены)
func (g *Grid) AccessibleNeighbors(c *Cell) []*Cell {
	var list []*Cell
	for _, d := range Directions {
		if (c.Walls & d.Wall) == 0 { // Стены нет
			nx, ny := c.X+d.DX, c.Y+d.DY
			if g.InBounds(nx, ny) {
				list = append(list, g.Get(nx, ny))
			}
		}
	}
	return list
}

// Reset сбрасывает лабиринт в исходное состояние (все стены возведены)
func (g *Grid) Reset() {
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			g.Cells[y][x].Walls = WallAll
			g.Cells[y][x].Visited = false
		}
	}
}

// Clone создает глубокую копию лабиринта
func (g *Grid) Clone() *Grid {
	clone := NewGrid(g.Width, g.Height)
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			clone.Cells[y][x] = g.Cells[y][x]
		}
	}
	return clone
}
