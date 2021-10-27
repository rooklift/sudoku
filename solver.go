package main

// Sudoku solver with constraint propagation.
// Loosely inspired by http://norvig.com/sudoku.html
//
// We do a depth first search. The trick really is that, at each step of the search:
//
//		- Setting a value at x,y eliminates the value as a possibility among peers.
//		- This can resolve some of the peers... which then eliminate possibles from *their* peers, recursively.
//
//		- We also check whether a number is forced into a box by being eliminated from the other 8 boxes in a unit.
//		- This likewise progagates constraints recursively.
//
//		- All of this happens during a single search iteration.
//
// At no point is it necessary to check for contradictions (e.g. the same number appearing twice in a row)
// because if this ever did occur, one cell would eliminate the number from its peers leaving the other cell
// with no possibles at all (which is then detected and treated as a fail).
//
// NOTE: internally we do Sudoku with numbers 0-8.
// The number 9 in puzzle source is converted to 0.

import (
	"fmt"
	"strings"
)

const PUZZLES = `
4.....8.5.3..........7......2.....6.....8.4......1.......6.3.7.5..2.....1.4......
52...6.........7.13...........4..8..6......5...........418.........3..2...87.....
6.....8.3.4.7.................5.4.7.3..2.....1.6.......2.....5.....8.6......1....
48.3............71.2.......7.5....6....2..8.............1.76...3.....4......5....
....14....3....2...7..........9...3.6.1.............8.2.....1.4....5.6.....7.8...
......52..8.4......3...9...5.1...6..2..7........3.....6...1..........7.4.......3.
6.2.5.........3.4..........43...8....1....2........7..5..27...........81...6.....
.524.........7.1..............8.2...3.....6...9.5.....1.6.3...........897........
6.2.5.........4.3..........43...8....1....2........7..5..27...........81...6.....
.923.........8.1...........1.7.4...........658.........6.5.2...4.....7.....9.....
6..3.2....5.....1..........7.26............543.........8.15........4.2........7..
.6.5.1.9.1...9..539....7....4.8...7.......5.8.817.5.3.....5.2............76..8...
..5...987.4..5...1..7......2...48....9.1.....6..2.....3..6..2.......9.7.......5..
3.6.7...........518.........1.4.5...7.....6.....2......2.....4.....8.3.....5.....
1.....3.8.7.4..............2.3.1...........958.........5.6...7.....8.2...4.......
6..3.2....4.....1..........7.26............543.........8.15........4.2........7..
....3..9....2....1.5.9..............1.2.8.4.6.8.5...2..75......4.1..6..3.....4.6.
45.....3....8.1....9...........5..9.2..7.....8.........1..4..........7.2...6..8..
.237....68...6.59.9.....7......4.97.3.7.96..2.........5..47.........2....8.......
..84...3....3.....9....157479...8........7..514.....2...9.6...2.5....4......9..56
.98.1....2......6.............3.2.5..84.........6.........4.8.93..5...........1..
..247..58..............1.4.....2...9528.9.4....9...1.........3.3....75..685..2...
4.....8.5.3..........7......2.....6.....5.4......1.......6.3.7.5..2.....1.9......
.2.3......63.....58.......15....9.3....7........1....8.879..26......6.7...6..7..4
1.....7.9.4...72..8.........7..1..6.3.......5.6..4..2.........8..53...7.7.2....46
4.....3.....8.2......7........1...8734.......6........5...6........1.4...82......
.......71.2.8........4.3...7...6..5....2..3..9........6...7.....8....4......5....
6..3.2....4.....8..........7.26............543.........8.15........8.2........7..
.47.8...1............6..7..6....357......5....1..6....28..4.....9.1...4.....2.69.
......8.17..2........5.6......7...5..1....3...8.......5......2..4..8....6...3....
38.6.......9.......2..3.51......5....3..1..6....4......17.5..8.......9.......7.32
...5...........5.697.....2...48.2...25.1...3..8..3.........4.7..13.5..9..2...31..
.2.......3.5.62..9.68...3...5..........64.8.2..47..9....3.....1.....6...17.43....
.8..4....3......1........2...5...4.69..1..8..2...........3.9....6....5.....2.....
..8.9.1...6.5...2......6....3.1.7.5.........9..4...3...5....2...7...3.8.2..7....4
4.....5.8.3..........7......2.....6.....5.8......1.......6.3.7.5..2.....1.8......
1.....3.8.6.4..............2.3.1...........958.........5.6...7.....8.2...4.......
1....6.8..64..........4...7....9.6...7.4..5..5...7.1...5....32.3....8...4........
249.6...3.3....2..8.......5.....6......2......1..4.82..9.5..7....4.....1.7...3...
...8....9.873...4.6..7.......85..97...........43..75.......3....3...145.4....2..1
...5.1....9....8...6.......4.1..........7..9........3.8.....1.5...2..4.....36....
......8.16..2........7.5......6...2..1....3...8.......2......7..3..8....5...4....
.476...5.8.3.....2.....9......8.5..6...1.....6.24......78...51...6....4..9...4..7
.....7.95.....1...86..2.....2..73..85......6...3..49..3.5...41724................
.4.5.....8...9..3..76.2.....146..........9..7.....36....1..4.5..6......3..71..2..
.834.........7..5...........4.1.8..........27...3.....2.6.5....5.....8........1..
..9.....3.....9...7.....5.6..65..4.....3......28......3..75.6..6...........12.3.8
.26.39......6....19.....7.......4..9.5....2....85.....3..2..9..4....762.........4
2.3.8....8..7...........1...6.5.7...4......3....1............82.5....6...1.......
6..3.2....1.....5..........7.26............843.........8.15........8.2........7..
1.....9...64..1.7..7..4.......3.....3.89..5....7....2.....6.7.9.....4.1....129.3.
.........9......84.623...5....6...453...1...6...9...7....1.....4.5..2....3.8....9
.2....5938..5..46.94..6...8..2.3.....6..8.73.7..2.........4.38..7....6..........5
9.4..5...25.6..1..31......8.7...9...4..26......147....7.......2...3..8.6.4.....9.
...52.....9...3..4......7...1.....4..8..453..6...1...87.2........8....32.4..8..1.
53..2.9...24.3..5...9..........1.827...7.........981.............64....91.2.5.43.
1....786...7..8.1.8..2....9........24...1......9..5...6.8..........5.9.......93.4
....5...11......7..6.....8......4.....9.1.3.....596.2..8..62..7..7......3.5.7.2..
.47.2....8....1....3....9.2.....5...6..81..5.....4.....7....3.4...9...1.4..27.8..
......94.....9...53....5.7..8.4..1..463...........7.8.8..7.....7......28.5.26....
.2......6....41.....78....1......7....37.....6..412....1..74..5..8.5..7......39..
1.....3.8.6.4..............2.3.1...........758.........7.5...6.....8.2...4.......
2....1.9..1..3.7..9..8...2.......85..6.4.........7...3.2.3...6....5.....1.9...2.5
..7..8.....6.2.3...3......9.1..5..6.....1.....7.9....2........4.83..4...26....51.
...36....85.......9.4..8........68.........17..9..45...1.5...6.4....9..2.....3...
34.6.......7.......2..8.57......5....7..1..2....4......36.2..1.......9.......7.82
......4.18..2........6.7......8...6..4....3...1.......6......2..5..1....7...3....
.4..5..67...1...4....2.....1..8..3........2...6...........4..5.3.....8..2........
.......4...2..4..1.7..5..9...3..7....4..6....6..1..8...2....1..85.9...6.....8...3
8..7....4.5....6............3.97...8....43..5....2.9....6......2...6...7.71..83.2
.8...4.5....7..3............1..85...6.....2......4....3.26............417........
....7..8...6...5...2...3.61.1...7..2..8..534.2..9.......2......58...6.3.4...1....
......8.16..2........7.5......6...2..1....3...8.......2......7..4..8....5...3....
.2..........6....3.74.8.........3..2.8..4..1.6..5.........1.78.5....9..........4.
.52..68.......7.2.......6....48..9..2..41......1.....8..61..38.....9...63..6..1.9
....1.78.5....9..........4..2..........6....3.74.8.........3..2.8..4..1.6..5.....
1.......3.6.3..7...7...5..121.7...9...7........8.1..2....8.64....9.2..6....4.....
4...7.1....19.46.5.....1......7....2..2.3....847..6....14...8.6.2....3..6...9....
......8.17..2........5.6......7...5..1....3...8.......5......2..3..8....6...4....
963......1....8......2.5....4.8......1....7......3..257......3...9.2.4.7......9..
15.3......7..4.2....4.72.....8.........9..1.8.1..8.79......38...........6....7423
..........5724...98....947...9..3...5..9..12...3.1.9...6....25....56.....7......6
....75....1..2.....4...3...5.....3.2...8...1.......6.....1..48.2........7........
6.....7.3.4.8.................5.4.8.7..2.....1.3.......2.....5.....7.9......1....
....6...4..6.3....1..4..5.77.....8.5...8.....6.8....9...2.9....4....32....97..1..
.32.....58..3.....9.428...1...4...39...6...5.....1.....2...67.8.....4....95....6.
...5.3.......6.7..5.8....1636..2.......4.1.......3...567....2.8..4.7.......2..5..
.5.3.7.4.1.........3.......5.8.3.61....8..5.9.6..1........4...6...6927....2...9..
..5..8..18......9.......78....4.....64....9......53..2.6.........138..5....9.714.
..........72.6.1....51...82.8...13..4.........37.9..1.....238..5.4..9.........79.
...658.....4......12............96.7...3..5....2.8...3..19..8..3.6.....4....473..
.2.3.......6..8.9.83.5........2...8.7.9..5........6..4.......1...1...4.22..7..8.9
.5..9....1.....6.....3.8.....8.4...9514.......3....2..........4.8...6..77..15..6.
.....2.......7...17..3...9.8..7......2.89.6...13..6....9..5.824.....891..........
3...8.......7....51..............36...2..4....7...........6.13..452...........8..
`

var steps int = 0

type Grid struct {
	cells	[9][9][9]bool							// Bools say whether their index is possible for the cell
	count	[9][9]int								// The number of possibles (trues) for the cell (kept up to date)
}

func NewGrid() *Grid {
	ret := new(Grid)
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			for n := 0; n < 9; n++ {
				ret.cells[x][y][n] = true
			}
			ret.count[x][y] = 9
		}
	}
	return ret
}

func (self *Grid) Copy() *Grid {
	ret := new(Grid)
	ret.cells = self.cells							// This works to copy the cells since we are only using
	ret.count = self.count							// actual arrays (if it was slices it wouldn't work)
	return ret										
}

func (self *Grid) Set(x, y, val int) {				// Set the val as the only possibility for x,y
	for n := 0; n < 9; n++ {						// then call Restrain() to propagate the consequences
		if n == val {
			self.cells[x][y][n] = true
		} else {
			self.cells[x][y][n] = false
		}
	}
	self.count[x][y] = 1
	self.Restrain(x, y, val)
}

func (self *Grid) Restrain(x, y, val int) {			// Eliminates val as a possibility from the *peers* of x,y
													// - calls Disallow() which can recursively call Restrain()
	// Eliminate along the horizontal...

	for x2 := 0; x2 < 9; x2++ {
		if x != x2 {
			self.Disallow(x2, y, val)
		}
	}

	// Check whether doing that forced any number to be in a certain place (because it is allowed in only 1 box)...

	for n := 0; n < 9; n++ {

		maybe_x := -1

		for x2 := 0; x2 < 9; x2++ {
			if self.cells[x2][y][n] {				// This cell is a possible location for n...
				if self.count[x2][y] == 1 {			// This is already the certain location of n
					maybe_x = -1
					break
				} else if maybe_x == -1 {			// First box we've seen where we might go
					maybe_x = x2
				} else {							// We've already seen one box where we could go
					maybe_x = -1
					break
				}
			}
		}

		if maybe_x != -1 {
			if self.count[maybe_x][y] > 1 {
				self.Set(maybe_x, y, n)
			}
		}
	}

	// Eliminate along the vertical...

	for y2 := 0; y2 < 9; y2++ {
		if y != y2 {
			self.Disallow(x, y2, val)
		}
	}

	// Check whether doing that forced any number to be in a certain place (because it is allowed in only 1 box)...

	for n := 0; n < 9; n++ {

		maybe_y := -1

		for y2 := 0; y2 < 9; y2++ {
			if self.cells[x][y2][n] {				// This cell is a possible location for n...
				if self.count[x][y2] == 1 {			// This is already the certain location of n
					maybe_y = -1
					break
				} else if maybe_y == -1 {			// First box we've seen where we might go
					maybe_y = y2
				} else {							// We've already seen one box where we could go
					maybe_y = -1
					break
				}
			}
		}

		if maybe_y != -1 {
			if self.count[x][maybe_y] > 1 {
				self.Set(x, maybe_y, n)
			}
		}
	}

	// Eliminate from the 3x3 area...

	startx := (x / 3) * 3
	starty := (y / 3) * 3

	for x3 := startx; x3 < startx + 3; x3++ {
		for y3 := starty; y3 < starty + 3; y3++ {
			if x3 != x || y3 != y {
				self.Disallow(x3, y3, val)
			}
		}
	}

	// Check whether doing that forced any number to be in a certain place (because it is allowed in only 1 box)...

	for n := 0; n < 9; n++ {

		maybe_x := -1
		maybe_y := -1

		Outer:
		for x3 := startx; x3 < startx + 3; x3++ {
			for y3 := starty; y3 < starty + 3; y3++ {
				if self.cells[x3][y3][n] {				// This cell is a possible location for n...
					if self.count[x3][y3] == 1 {		// This is already the certain location of n
						maybe_x = -1
						maybe_y = -1
						break Outer
					} else if maybe_x == -1 && maybe_y == -1 {		// First box we've seen where we might go
						maybe_x = x3
						maybe_y = y3
					} else {							// We've already seen one box where we could go
						maybe_x = -1
						maybe_y = -1
						break Outer
					}
				}
			}
		}

		if maybe_x != -1 && maybe_y != -1 {
			if self.count[maybe_x][maybe_y] > 1 {
				self.Set(maybe_x, maybe_y, n)
			}
		}
	}
}

func (self *Grid) Disallow(x, y, val int) {			// Disallow the value from x,y and check if this resolves
													// the cell (reduces it to 1 possible value) and if so,
													// call Restrain() to propagate the consequences
	if self.cells[x][y][val] == false {
		return										// Do nothing if already forbidden
	}

	self.cells[x][y][val] = false
	self.count[x][y]--

	if self.count[x][y] == 1 {						// Cell is resolved, so propagate the consequences
		goodval := -1
		for n := 0; n < 9; n++ {
			if self.cells[x][y][n] {
				goodval = n
				break
			}
		}
		self.Restrain(x, y, goodval)
	}
}

func (self *Grid) Solve() *Grid {					// Returns the solved grid, or nil if there was no solution

	steps++

	x_index := -1
	y_index := -1
	got_zero := false
	got_above_one := false
	lowest_above_one := 999

	// If there is a cell with zero possibles, our grid is illegal.
	// If there are no cells with more than one possible, our grid is solved.
	// Otherwise, we find the cell with the smallest number of possibles so we can test each in turn.

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			count := 0
			for n := 0; n < 9; n++ {
				if self.cells[x][y][n] {
					count++
				}
			}
			if count == 0 {
				got_zero = true
			}
			if count > 1 {
				got_above_one = true
				if count < lowest_above_one {
					lowest_above_one = count
					x_index = x
					y_index = y
				}
			}
		}
	}

	if got_zero {									// We have a cell with zero possibles
		return nil
	} else if !got_above_one {						// The puzzle is solved
		return self
	} else {										// Try each possible for the chosen x,y in turn...

		var possibles []int

		for n := 0; n < 9; n++ {
			if self.cells[x_index][y_index][n] {
				possibles = append(possibles, n)
			}
		}

		for _, n := range possibles {
			foo := self.Copy()
			foo.Set(x_index, y_index, n)
			result := foo.Solve()
			if result != nil {
				return result
			}
		}
	}

	return nil
}

func (self *Grid) Print() {
	for y := 0; y < 9; y++ {
		if y == 3 || y == 6 {
			fmt.Printf(" ------+-------+------\n")
		}
		for x := 0; x < 9; x++ {
			if x == 3 || x == 6 {
				fmt.Printf(" |")
			}
			s := ""
			for n := 0; n < 9; n++ {
				if self.cells[x][y][n] {
					if s == "" {
						s = fmt.Sprintf("%d", n)
						if s == "0" {
							s = "9"					// Internally we use 0 instead of 9
						}
					} else {
						s = "."
					}
				}
			}
			if s == "" {
				s = "?"
			}
			fmt.Printf(" %s", s)
		}
		fmt.Printf("\n")
	}
}

func (self *Grid) SetFromString(s string) {

	var numbers []int

	for _, c := range s {
		if c == '.' || c == '0' {
			numbers = append(numbers, -1)
		} else if c >= '1' && c <= '9' {
			numbers = append(numbers, int(c) - 48)
		} else {
			continue
		}
	}

	if len(numbers) != 81 {
		panic("Bad puzzle string")
	}

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			index := y * 9 + x
			if numbers[index] <= 0 {
				continue
			} else if numbers[index] == 9 {			// Internally we use 0 instead of 9
				self.Set(x, y, 0)
			} else {
				self.Set(x, y, numbers[index])
			}
		}
	}
}

func main() {

	lines := strings.Split(PUZZLES, "\n")

	for n, line := range lines {

		if len(line) < 81 {
			continue
		}

		grid := NewGrid()
		grid.SetFromString(line)
		fmt.Printf("%d. New puzzle...\n", n)
		grid.Print()
		solution := grid.Solve()
		if solution == nil {
			panic("No solution found")
		} else {
			fmt.Printf("Solution found... (search tree size was %d)\n", steps)
			steps = 0
			solution.Print()
		}
	}
}

