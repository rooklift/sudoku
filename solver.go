package main

// Sudoku solver with constraint propagation.
// Loosely inspired by http://norvig.com/sudoku.html
//
// We do a depth first search. The trick really is that, at each step of the search:
//
//		- Setting a value at x,y eliminates the value as a possibility among peers.
//		- This can resolve some of the peers... which then eliminate possibles from *their* peers, recursively.
//		- NOTE: this recursive procedure is not the search itself, but merely one iteration of the search.
//
// At no point is it necessary to check for contradictions (e.g. the same number appearing twice in a row)
// because if this ever did occur, one cell would eliminate the number from its peers leaving the other cell
// with no possibles at all (which is then detected and treated as a fail).
//
// NOTE: internally we do Sudoku with numbers 0-8.
// The number 9 in puzzle source is converted to 0.

import (
	"fmt"
	"strconv"
	"strings"
)

const PUZZLE = "..5.2.6...9...4.1.2..5....3..6.3.......8.1.......9.4..3....2..7.1.9...5...4.6.8.."

type Grid struct {
	cells	[9][9][9]bool							// Bools say whether their index is possible for the cell
}

func NewGrid() *Grid {
	ret := new(Grid)
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			for n := 0; n < 9; n++ {
				ret.cells[x][y][n] = true
			}
		}
	}
	return ret
}

func (self *Grid) Copy() *Grid {
	ret := new(Grid)
	ret.cells = self.cells							// This works to copy the cells since we are only using
	return ret										// actual arrays (if it was slices it wouldn't work)
}

func (self *Grid) Set(x, y, val int) {				// Set the val as the only possibility for x,y
	for n := 0; n < 9; n++ {						// then call Restrain() to propagate the consequences
		if n == val {
			self.cells[x][y][n] = true
		} else {
			self.cells[x][y][n] = false
		}
	}
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

	// Eliminate along the vertical...

	for y2 := 0; y2 < 9; y2++ {
		if y != y2 {
			self.Disallow(x, y2, val)
		}
	}

	// Work out the boundaries of the 3x3 area we are in...

	startx := (x / 3) * 3
	starty := (y / 3) * 3

	// Eliminate from the 3x3 area...

	for x3 := startx; x3 < startx + 3; x3++ {
		for y3 := starty; y3 < starty + 3; y3++ {
			if x3 != x || y3 != y {
				self.Disallow(x3, y3, val)
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

	// Check if that resolves the cell...

	count := 0
	goodval := -1

	for n := 0; n < 9; n++ {
		if self.cells[x][y][n] {
			count++;
			goodval = n
		}
	}

	if count == 1 {
		self.Restrain(x, y, goodval)				// Cell is resolved, so propagate the consequences
	}
}

func (self *Grid) Solve() *Grid {					// Returns the solved grid, or nil if there was no solution

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
	if len(s) != 81 {
		panic("Bad puzzle string")
	}
	len_1_strings := strings.Split(s, "")
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			index := y * 9 + x
			val, err := strconv.Atoi(len_1_strings[index])
			if err != nil || val == 0 {
				continue
			}
			if val == 9 {							// Internally we use 0 instead of 9
				val = 0
			}
			self.Set(x, y, val)
		}
	}
}

func main() {

	p := NewGrid()

	p.SetFromString(PUZZLE)

	p.Print()

	solution := p.Solve()
	if solution == nil {
		fmt.Printf("No solution found\n")
	} else {
		fmt.Printf("Solution found...\n")
		solution.Print()
	}
}

