package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

var steps int = 0

type Point struct {
	x		int
	y		int
}

// ------------------------------------------------------------------------------------------------
// Peer lookup tables - a peer is a cell which is "seen" by a cell. Every cell sees 20 other cells.

var lookup_peers [9][9][20]Point

func init() {

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

			for n := 0; n < 20; n++ {
				lookup_peers[x][y][n] = peers[n]
			}
		}
	}
}

// ------------------------------------------------------------------------------------------------
// Grid - definition and creation...

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
	ret.cells = self.cells							// This works to copy the cells since we are only using actual arrays (if it was slices it wouldn't work)
	return ret										
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
		panic("Set() tried to set a value already ruled out.")
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

	if self.Count(x, y) == 1 {
		fixed_value := self.Value(x, y)
		peers := lookup_peers[x][y]
		for _, peer := range peers {
			self.Eliminate(peer.x, peer.y, fixed_value)
		}
	}

	// Norvig strategy #2...
	// TODO
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
			count := self.Count(x, y)
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

		possibles := self.Possibles(x_index, y_index)

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

