package main

// Sudoku solver with constraint propagation.
// Loosely inspired by http://norvig.com/sudoku.html
//
// We do a depth first search. The trick really is that, at each step of the search:
//
//		- The fundamental operation is eliminating a value as a possiblity.
//		- Eliminating a value can cause a cell to be solved, which then eliminates it from its peers.
//		- Eliminating a value can cause that value to be forced into some other cell (the last remaining option).
//		- The Eliminate() function is recursive, i.e. one elimination can trigger more eliminations.
//
// Note: internally we do Sudoku with numbers 0-8. The number nine in puzzles becomes our zero.

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type Point struct {
	x		int
	y		int
}

var lookup_units [9][9][][]Point					// Can retrieve the 3 units a cell belongs to.
var lookup_peers [9][9][]Point						// Can retrieve the 20 peers a cell has.

var all_units [][]Point

// ------------------------------------------------------------------------------------------------
// Unit lookup tables - a unit is a set of 9 cells. Each cell belongs to 3 units.
// There are a total of 27 units.

func build_unit_tables() {

	// Columns...

	for x := 0; x < 9; x++ {
		var unit []Point
		for y := 0; y < 9; y++ {
			unit = append(unit, Point{x, y})
		}
		all_units = append(all_units, unit)
	}

	// Rows...

	for y := 0; y < 9; y++ {
		var unit []Point
		for x := 0; x < 9; x++ {
			unit = append(unit, Point{x, y})
		}
		all_units = append(all_units, unit)
	}

	// 3x3 squares...

	for startx := 0; startx <= 6; startx += 3 {
		for starty := 0; starty <= 6; starty += 3 {
			var unit []Point
			for x := startx; x < startx + 3; x++ {
				for y := starty; y < starty + 3; y++ {
					unit = append(unit, Point{x, y})
				}
			}
			all_units = append(all_units, unit)
		}
	}

	if len(all_units) != 27 {
		panic("Wat?")
	}

	unit_contains := func(unit []Point, x, y int) bool {		// Helper function
		for _, point := range unit {
			if point.x == x && point.y == y {
				return true
			}
		}
		return false
	}

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			for _, unit := range all_units {
				if unit_contains(unit, x, y) {
					lookup_units[x][y] = append(lookup_units[x][y], unit)
				}
			}
			if len(lookup_units[x][y]) != 3 {
				panic("Wat?")
			}
		}
	}
}

// ------------------------------------------------------------------------------------------------
// Peer lookup tables - a peer is a cell which is "seen" by a cell. Every cell sees 20 other cells.

func build_peer_tables() {

	for x := 0; x < 9; x++ {

		for y := 0; y < 9; y++ {

			var peers []Point

			for x2 := 0; x2 < 9; x2++ {
				if x2 != x {
					peers = append(peers, Point{x2, y})
				}
			}

			for y2 := 0; y2 < 9; y2++ {
				if y2 != y {
					peers = append(peers, Point{x, y2})
				}
			}

			square_startx := (x / 3) * 3
			square_starty := (y / 3) * 3

			for x3 := square_startx; x3 < square_startx + 3; x3++ {
				for y3 := square_starty; y3 < square_starty + 3; y3++ {
					if x3 != x && y3 != y {
						peers = append(peers, Point{x3, y3})
					}
				}
			}

			if len(peers) != 20 {
				panic("Wat?")
			}

			lookup_peers[x][y] = peers
		}
	}
}

// ------------------------------------------------------------------------------------------------
// Grid - our main data structure, definition, creation, and validation...

type Grid struct {
	cells	[9][9][9]bool							// Bools say whether their index is possible for the cell.
	steps	*int									// How many times Solve() is called. Shared between grids with the same origin.
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
	ret.steps = new(int)
	return ret
}

func (self *Grid) Copy() *Grid {
	ret := new(Grid)
	ret.cells = self.cells							// This works to copy the cells since we are only using actual arrays (if it was slices it wouldn't work)
	ret.steps = self.steps							// Same pointer
	return ret										
}

func (self *Grid) Validate() bool {					// Complete test of whether the solution is valid. Only used for sanity checking, not during search.

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if self.Count(x, y) != 1 {
				return false
			}
		}
	}

	for _, unit := range all_units {
		set := make(map[int]bool)
		for _, point := range unit {
			set[self.Value(point.x, point.y)] = true
		}
		if len(set) != 9 {
			return false
		}
	}

	return true
}

// ------------------------------------------------------------------------------------------------
// Grid - manipulation and solving...

func (self *Grid) Count(x, y int) int {				// The number of possibles at x,y - maybe optimise this away later
	ret := 0
	for n := 0; n < 9; n++ {
		if self.cells[x][y][n] {
			ret++
		}
	}
	return ret
}

func (self *Grid) Value(x, y int) int {				// The value locked in to x,y, only valid iff Count(x,y) == 1
	for n := 0; n < 9; n++ {
		if self.cells[x][y][n] {
			return n
		}
	}
	panic("Value() called but cell had zero possibles")
}

func (self *Grid) Possibles(x, y int) []int {		// List of all possible values for x,y
	var ret []int
	for n := 0; n < 9; n++ {
		if self.cells[x][y][n] {
			ret = append(ret, n)
		}
	}
	return ret
}

func (self *Grid) Set(x, y, val int) {
	if self.cells[x][y][val] == false {
		panic("Set() tried to set a value already ruled out")
	}
	for n := 0; n < 9; n++ {
		if n != val {
			self.Eliminate(x, y, n)
		}
	}
}

func (self *Grid) Eliminate(x, y, val int) {

	if self.cells[x][y][val] == false {
		return
	}

	self.cells[x][y][val] = false

	// Norvig strategy #1...
	// If the cell now has only 1 value, it is fixed here and must be removed from all the peers...

	if self.Count(x, y) == 1 {
		fixed_value := self.Value(x, y)
		peers := lookup_peers[x][y]
		for _, peer := range peers {
			self.Eliminate(peer.x, peer.y, fixed_value)
		}
	}

	// Norvig strategy #2...
	// For each unit containing x,y, the elimination may have forced val into some other square (if it's val's last option)

	units := lookup_units[x][y]

	for _, unit := range units {

		options := 0
		for _, point := range unit {
			if self.cells[point.x][point.y][val] {
				options++
			}
		}

		if options == 1 {
			for _, point := range unit {						// Find it again! Could optimise this away.
				if self.cells[point.x][point.y][val] {
					if self.Count(point.x, point.y) > 1 {		// i.e. this cell wasn't already solved
						self.Set(point.x, point.y, val)
					}
				}
			}
		}
	}
}

func (self *Grid) Solve() *Grid {					// Returns the solved grid, or nil if there was no solution

	*self.steps++

	x_index := -1
	y_index := -1
	got_above_one := false
	lowest_above_one := 999

	// Some counting of possibilities in cells...
	// If we need to search, we find the cell with the smallest number of possibles so we can test each in turn.

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			count := self.Count(x, y)
			if count == 0 {
				return nil							// We have a cell with zero possibles - grid is illegal
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

	if !got_above_one {								// Every cell has exactly 1 possible - the puzzle is solved
		return self
	}

	// Try each possible for the chosen x,y in turn...

	possibles := self.Possibles(x_index, y_index)

	for _, n := range possibles {
		foo := self.Copy()
		foo.Set(x_index, y_index, n)
		result := foo.Solve()
		if result != nil {
			return result
		}
	}

	return nil
}

// ------------------------------------------------------------------------------------------------
// Grid - utility methods...

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

// ------------------------------------------------------------------------------------------------

func init() {
	build_unit_tables()
	build_peer_tables()
}

func main() {

	f, err := ioutil.ReadFile("puzzles.txt")

	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(f), "\n")

	puzzle_id := 0
	var fails []int

	for _, line := range lines {

		if len(line) < 81 {
			continue
		}

		puzzle_id++
		grid := NewGrid()
		grid.SetFromString(line)
		fmt.Printf("%d. New puzzle...\n", puzzle_id)
		grid.Print()

		solution := grid.Solve()
		
		if solution == nil {
			fmt.Printf("No solution found! (search tree size was %d)\n", *grid.steps)
			fails = append(fails, puzzle_id)
		} else if solution.Validate() == false {
			panic("Solution failed validation")
		} else {
			fmt.Printf("Solution found... (search tree size was %d)\n", *solution.steps)
			solution.Print()
		}
	}

	if len(fails) > 0 {
		fmt.Printf("\nFailures: %v\n", fails)
	}

}

