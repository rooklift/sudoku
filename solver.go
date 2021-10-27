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
	"io/ioutil"
	"strings"
)

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

	f, err := ioutil.ReadFile("puzzles.txt")

	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(f), "\n")

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

