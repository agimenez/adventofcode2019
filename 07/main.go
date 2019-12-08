package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

const (
	debug = false
)

func dbg(fmt string, v ...interface{}) {
	if debug {
		log.Printf(fmt, v...)
	}
}

type program struct {
	mem    []int
	pc     int
	output []int
}

func newProgram(p string) *program {

	pr := &program{
		mem:    []int{},
		pc:     0,
		output: []int{},
	}

	pSlice := strings.Split(p, ",")
	for _, b := range pSlice {
		i, err := strconv.Atoi(b)
		if err != nil {
			log.Fatal(err)
		}

		pr.mem = append(pr.mem, i)
	}

	return pr
}

func (p *program) run(input []int) {

	for op := p.mem[p.pc]; op != 99; {
		dbg("MEM = %v", p.mem)
		dbg("pc = %d; op = %v", p.pc, op)

		opcode := op % 100
		dbg("opcode = %v", opcode)

		switch opcode {
		case 1: // Add
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+4])
			a, b, c := p.fetchParameter(1), p.fetchParameter(2), p.mem[p.pc+3]
			dbg(" ADD %d %d -> %d", a, b, c)
			p.mem[c] = a + b
			p.pc += 4
		case 2: // Mul
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+4])
			a, b, c := p.fetchParameter(1), p.fetchParameter(2), p.mem[p.pc+3]
			dbg(" MUL %d %d -> %d", a, b, c)
			p.mem[c] = a * b
			p.pc += 4
		case 3: // In
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+2])
			var in, dst int
			in, input, dst = input[0], input[1:], p.mem[p.pc+1]
			dbg(" IN  %d -> %d", in, dst)
			p.mem[dst] = in
			p.pc += 2
		case 4: // Out
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+2])
			src := p.fetchParameter(1)
			dbg(" OUT %d", src)

			p.output = append(p.output, src)
			p.pc += 2

		case 5: // JMP IF TRUE
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+3])
			tst, newpc := p.fetchParameter(1), p.fetchParameter(2)
			dbg(" JMP %d if %d != 0", newpc, tst)
			if tst != 0 {
				p.pc = newpc
			} else {
				p.pc += 3
			}

		case 6: // JMP IF FALSE
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+3])
			tst, newpc := p.fetchParameter(1), p.fetchParameter(2)
			dbg(" JMP %d if %d == 0", newpc, tst)
			if tst == 0 {
				p.pc = newpc
			} else {
				p.pc += 3
			}

		case 7: // LT
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+4])
			first, second, dst := p.fetchParameter(1), p.fetchParameter(2), p.mem[p.pc+3]
			dbg(" LT %d %d %d", first, second, dst)
			if first < second {
				p.mem[dst] = 1
			} else {
				p.mem[dst] = 0
			}

			p.pc += 4

		case 8: // EQ
			dbg(" INSTR = %v", p.mem[p.pc:p.pc+4])
			first, second, dst := p.fetchParameter(1), p.fetchParameter(2), p.mem[p.pc+3]
			dbg(" EQ %d %d %d", first, second, dst)
			if first == second {
				p.mem[dst] = 1
			} else {
				p.mem[dst] = 0
			}
			p.pc += 4

		default:
			log.Fatalf("Bad opcode = %v", op)
		}
		dbg("MEM = %v", p.mem)

		op = p.mem[p.pc]
	}
}

func (p *program) fetchParameter(n int) int {
	opcode := p.mem[p.pc]
	parameter := p.mem[p.pc+n]
	mode := opcode / int(math.Pow10(n+1)) % 10

	dbg("  (fetch) param[%d](%d) mode: %d", n, parameter, mode)
	if mode == 0 {
		return p.mem[parameter]
	}

	return parameter
}

// Heap's algorith
// Credits to https://stackoverflow.com/a/30226442/4735682
func slicePermutations(s []int) [][]int {
	perms := [][]int{}

	// fwd declaration to be able to call itself recursively
	var helper func(arr []int, n int)
	helper = func(arr []int, n int) {
		if n == 1 {
			tmp := make([]int, len(s))
			copy(tmp, s)
			perms = append(perms, tmp)
			return
		}

		for i := 0; i < n; i++ {
			helper(s, n-1)

			if n%2 == 1 {
				arr[i], arr[n-1] = arr[n-1], arr[i]
			} else {
				arr[0], arr[n-1] = arr[n-1], arr[0]
			}
		}
	}

	helper(s, len(s))

	return perms
}

func main() {
	var in string
	fmt.Scan(&in)

	maxOutput := 0
	perms := slicePermutations([]int{0, 1, 2, 3, 4})
	for _, perm := range perms {
		thrustersSignal := runPermutation(&in, perm)
		log.Printf("Got signal from perm %v: %v", perm, thrustersSignal)
		if thrustersSignal > maxOutput {
			maxOutput = thrustersSignal
		}
	}

	log.Printf("Max output: %v", maxOutput)

}

func runPermutation(in *string, phaseSettings []int) int {
	input := 0
	for _, ampSetting := range phaseSettings {
		program := newProgram(*in)
		program.run([]int{ampSetting, input})

		// input for the next phase
		input = program.output[0]

	}

	return input
}