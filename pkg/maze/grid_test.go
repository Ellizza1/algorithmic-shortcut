package maze

import "testing"

func TestNewGrid(t *testing.T) {
	w, h := 10, 10
	grid := NewGrid(w, h)

	if grid.Width != w || grid.Height != h {
		t.Fatalf("expected size %dx%d, got %dx%d", w, h, grid.Width, grid.Height)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := grid.Get(x, y)
			if c == nil {
				t.Fatalf("cell (%d, %d) is nil", x, y)
			}
			if c.Walls != WallAll {
				t.Fatalf("expected all walls in cell (%d, %d), got %d", x, y, c.Walls)
			}
		}
	}
}

func TestCarvePassageSymmetry(t *testing.T) {
	grid := NewGrid(5, 5)
	c1 := grid.Get(1, 1)
	c2 := grid.Get(2, 1) // Сосед справа

	carved := grid.CarvePassage(c1, c2)
	if !carved {
		t.Fatal("expected CarvePassage to succeed")
	}

	// У c1 должна пропасть стена на восток (WallRight)
	if (c1.Walls & WallRight) != 0 {
		t.Errorf("expected c1 to have no right wall, walls: %04b", c1.Walls)
	}
	// У c2 должна симметрично пропасть стена на запад (WallLeft)
	if (c2.Walls & WallLeft) != 0 {
		t.Errorf("expected c2 to have no left wall, walls: %04b", c2.Walls)
	}

	if !grid.AreConnected(c1, c2) {
		t.Error("expected c1 and c2 to be connected")
	}
}
