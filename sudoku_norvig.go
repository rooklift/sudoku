package main

// Sudoku solver with constraint propagation.
// This version more directly ports Norvig's implementation.
// But (when not IO-bound by the terminal) it's about 10x slower.

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

var digits string = "123456789"
var rows string = "ABCDEFGHI"		// Note this is quite unlike
var cols string = digits			// Chess square names

var name [9][9]string				// Lookup table: x, y --> name

var squares []string				// The 81 names of the squares, e.g. "A1"
var unitlist [][]string				// 27 units which are len-9 lists of squares

var units map[string][][]string		// Lookup table: square --> units containing it
var peers map[string][]string		// Lookup table: square --> peers it sees

func init() {

	// Names of squares

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			name[x][y] = fmt.Sprintf("%c%c", rows[y], digits[x])
			squares = append(squares, name[x][y])
		}
	}

	// List of all 27 units, each of which is a list of 9 squares

	for x := 0; x < 9; x++ {
		var unit []string
		for y := 0; y < 9; y++ {
			unit = append(unit, name[x][y])
		}
		unitlist = append(unitlist, unit)
	}

	for y := 0; y < 9; y++ {
		var unit []string
		for x := 0; x < 9; x++ {
			unit = append(unit, name[x][y])
		}
		unitlist = append(unitlist, unit)
	}

	for startx := 0; startx <= 6; startx += 3 {
		for starty := 0; starty <= 6; starty += 3 {
			var unit []string
			for x := startx; x < startx + 3; x++ {
				for y := starty; y < starty + 3; y++ {
					unit = append(unit, name[x][y])
				}
			}
			unitlist = append(unitlist, unit)
		}
	}

	// The lookup table: square --> list of units containing it

	units = make(map[string][][]string)

	for _, s := range squares {
		for _, unit := range unitlist {
			for _, z := range unit {
				if z == s {								// unit contains the square
					units[s] = append(units[s], unit)
					break
				}
			}
		}
	}

	// The lookup table: square --> list of peers that square sees

	peers = make(map[string][]string)

	for _, s := range squares {
		
		for _, s2 := range squares {

			if s == s2 {
				continue
			}

			for _, unit := range units[s] {

				contains_s := false
				contains_s2 := false

				for _, z := range unit {
					if z == s {
						contains_s = true
					}
					if z == s2 {
						contains_s2 = true
					}
				}

				if contains_s && contains_s2 {
					already_got_s2 := false
					for _, z := range peers[s] {
						if z == s2 {
							already_got_s2 = true
							break
						}
					}
					if !already_got_s2 {
						peers[s] = append(peers[s], s2)
					}
				}
			}
		}
	}

	run_tests()
}

func run_tests() {

	if len(squares) != 81 {
		panic("squares invalid")
	}
	if len(unitlist) != 27 {
		panic("unitlist invalid")
	}
	for _, unit := range unitlist {
		if len(unit) != 9 {
			panic("unit invalid")
		}
	}

	for _, s := range squares {
		if len(peers[s]) != 20 {
			panic("peer list invalid")
		}
		if len(units[s]) != 3 {
			panic("unit list invalid")
		}
		for _, unit := range units[s] {
			if len(unit) != 9 {
				panic("unit invalid")
			}
		}
	}
}

// ------------------------------------------------------------------------------------------------

func assign(values map[string]string, s, d string) map[string]string {

	if strings.Contains(values[s], d) == false {
		panic("Invalid call to assign()")
	}

	possibles := strings.Split(values[s], "")

	for _, d2 := range possibles {
		if d2 != d {
			if eliminate(values, s, d2) == nil {
				return nil
			}
		}
	}

	return values
}

func eliminate(values map[string]string, s, d string) map[string]string {

	if values == nil {
		return nil
	}

	if strings.Contains(values[s], d) == false {
		return values
	}

	values[s] = strings.ReplaceAll(values[s], d, "")

	if len(values[s]) == 0 {
		return nil
	}

	// Norvig strategy #1

	if len(values[s]) == 1 {
		d2 := values[s]
		for _, s2 := range peers[s] {
			if eliminate(values, s2, d2) == nil {
				return nil
			}
		}
	}

	// Norvig strategy #2

	for _, u := range units[s] {

		dplaces := 0
		place := ""

		for _, s2 := range u {
			if strings.Contains(values[s2], d) {
				dplaces++
				place = s2
			}
		}

		if dplaces == 0 {
			return nil
		}

		if dplaces == 1 {
			if assign(values, place, d) == nil {
				return nil
			}
		}
	}

	return values
}

func copy_map(m map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func search(values map[string]string) map[string]string {
	
	if values == nil {
		return nil
	}

	done := true

	lowest_count := 999
	search_square := ""

	for _, s := range squares {

		count := len(values[s])

		if count == 0 {
			return nil
		}

		if count > 1 {
			done = false
			if count < lowest_count {
				lowest_count = count
				search_square = s
			}
		}
	}

	if done {
		return values
	}

	for _, d := range values[search_square] {
		foo := copy_map(values)
		assign(foo, search_square, string(d))
		result := search(foo)
		if result != nil {
			return result
		}
	}

	return nil
}

func new_map() map[string]string {
	ret := make(map[string]string)
	for _, s := range squares {
		ret[s] = digits
	}
	return ret
}

func parse_string(s string) map[string]string {

	var numstrings []string

	for _, c := range s {
		if c == '.' || c == '0' {
			numstrings = append(numstrings, "")
		} else if c >= '1' && c <= '9' {
			numstrings = append(numstrings, string(c))
		} else {
			continue
		}
	}

	if len(numstrings) != 81 {
		panic("Bad puzzle string")
	}

	ret := new_map()

	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			index := y * 9 + x
			if numstrings[index] == "" {
				continue
			} else {
				assign(ret, name[x][y], numstrings[index])
			}
		}
	}

	return ret
}

func print(values map[string]string) {
	for y := 0; y < 9; y++ {
		if y == 3 || y == 6 {
			fmt.Printf(" ------+-------+------\n")
		}
		for x := 0; x < 9; x++ {
			if x == 3 || x == 6 {
				fmt.Printf(" |")
			}
			s := "?"
			if len(values[name[x][y]]) > 1 {
				s = "."
			} else if len(values[name[x][y]]) == 1 {
				s = values[name[x][y]]
			}
			fmt.Printf(" %s", s)
		}
		fmt.Printf("\n")
	}
}

func validate(values map[string]string) bool {

	for _, d := range values {
		if len(d) != 1 {
			return false
		}
	}

	for _, unit := range unitlist {
		set := make(map[string]bool)
		for _, d := range unit {
			set[d] = true
		}
		if len(set) != 9 {
			return false
		}
	}

	return true
}

func main() {

	f, err := ioutil.ReadFile("puzzles.txt")

	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(f), "\n")

	puzzle_id := 0
	var fails []int

	start_time := time.Now()

	for _, line := range lines {

		if len(line) < 81 {
			continue
		}

		puzzle_id++
		grid := parse_string(line)
		fmt.Printf("%d. New puzzle...\n", puzzle_id)
		print(grid)

		solution := search(grid)
		
		if solution == nil {
			fmt.Printf("No solution found!\n")
			fails = append(fails, puzzle_id)
		} else if validate(solution) == false {
			panic("Solution failed validation")
		} else {
			fmt.Printf("Solution found...\n")
			print(solution)
		}
	}

	if len(fails) > 0 {
		fmt.Printf("\nFailures: %v\n", fails)
	}

	fmt.Printf("\nElapsed time: %v\n", time.Now().Sub(start_time))
}