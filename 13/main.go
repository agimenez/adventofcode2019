package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

var (
	debug int
)

func dbg(level int, fmt string, v ...interface{}) {
	if debug >= level {
		log.Printf(fmt+"\n", v...)
	}
}

type program struct {
	mem  []int
	pc   int
	base int
}

func newProgram(p string) *program {

	pr := &program{
		mem:  []int{},
		pc:   0,
		base: 0,
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

// ensureAddr ensures that a given address is reachable, by growing the memory slice
// if neccessary
func (p *program) ensureAddr(addr int) {
	extend := addr - len(p.mem) + 1
	if addr >= len(p.mem) {
		p.mem = append(p.mem, make([]int, extend)...)
	}
}

func (p *program) setMem(addr, val int) {
	p.ensureAddr(addr)

	dbg(2, "  (setMem) p.mem[%d] = %d", addr, val)
	p.mem[addr] = val
}

func (p *program) getMem(addr int) int {
	p.ensureAddr(addr)

	dbg(2, "  (get) p.mem[%d] = %d", addr, p.mem[addr])
	return p.mem[addr]
}
func (p *program) run(input <-chan int, output chan<- int, wantinput chan<- bool) {

	for op := p.mem[p.pc]; op != 99; {
		dbg(4, "MEM = %v", p.mem)
		dbg(3, "pc = %d; op = %v", p.pc, op)

		opcode := op % 100
		dbg(3, "opcode = %v", opcode)

		switch opcode {
		case 1: // Add
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+4])
			a, b, c := p.fetchParameter(1), p.fetchParameter(2), p.getAddrIndex(3)
			dbg(3, " ADD %d %d -> %d", a, b, c)
			p.setMem(c, a+b)
			p.pc += 4
		case 2: // Mul
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+4])
			a, b, c := p.fetchParameter(1), p.fetchParameter(2), p.getAddrIndex(3)
			dbg(3, " MUL %d %d -> %d", a, b, c)
			p.setMem(c, a*b)
			p.pc += 4
		case 3: // In
			dbg(2, " INSTR = %v", p.mem[p.pc:p.pc+2])
			var in, dst int
			wantinput <- true
			in, dst = <-input, p.getAddrIndex(1)
			dbg(2, " IN  %d -> mem[%d]", in, dst)
			p.setMem(dst, in)
			p.pc += 2
		case 4: // Out
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+2])
			src := p.fetchParameter(1)
			dbg(3, " OUT %d", src)

			output <- src
			p.pc += 2

		case 5: // JMP IF TRUE
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+3])
			tst, newpc := p.fetchParameter(1), p.fetchParameter(2)
			dbg(3, " JMP %d if %d != 0", newpc, tst)
			if tst != 0 {
				p.pc = newpc
			} else {
				p.pc += 3
			}

		case 6: // JMP IF FALSE
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+3])
			tst, newpc := p.fetchParameter(1), p.fetchParameter(2)
			dbg(3, " JMP %d if %d == 0", newpc, tst)
			if tst == 0 {
				p.pc = newpc
			} else {
				p.pc += 3
			}

		case 7: // LT
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+4])
			first, second, dst := p.fetchParameter(1), p.fetchParameter(2), p.getAddrIndex(3)
			dbg(3, " LT %d %d %d", first, second, dst)
			if first < second {
				p.setMem(dst, 1)
			} else {
				p.setMem(dst, 0)
			}

			p.pc += 4

		case 8: // EQ
			dbg(3, " INSTR = %v", p.mem[p.pc:p.pc+4])
			first, second, dst := p.fetchParameter(1), p.fetchParameter(2), p.getAddrIndex(3)
			dbg(3, " EQ %d %d %d", first, second, dst)
			if first == second {
				p.setMem(dst, 1)
			} else {
				p.setMem(dst, 0)
			}
			p.pc += 4

		case 9: // RELBASE
			dbg(2, " INSTR = %v", p.mem[p.pc:p.pc+2])
			offset := p.fetchParameter(1)
			dbg(2, " RELBASE  %d", offset)

			p.base += offset
			p.pc += 2

		default:
			log.Fatalf("Bad opcode = %v", op)
		}
		dbg(4, " MEM = %v", p.mem)

		op = p.mem[p.pc]
	}
}

func (p *program) instructionMode(offset int) int {
	opcode := p.mem[p.pc]
	return opcode / int(math.Pow10(offset+1)) % 10
}

// This is for writing to mem operations (IN, etc)
func (p *program) getAddrIndex(n int) int {
	parameter := p.mem[p.pc+n]
	mode := p.instructionMode(n)

	if mode == 0 {
		return parameter
	} else if mode == 2 {
		return p.base + parameter
	} else {
		panic("unsupported immediate mode for writing")
	}
}

func (p *program) fetchParameter(n int) int {
	mode := p.instructionMode(n)
	parameter := p.mem[p.pc+n]
	dbg(2, "   param[%d](%d) mode: %d, base %d, memsize %d", n, parameter, mode, p.base, len(p.mem))

	if mode == 0 {
		// position mode
		val := p.getMem(parameter)
		dbg(2, "   (posmode) -> mem[%d] = %d", parameter, val)
		return val
	} else if mode == 2 {
		// relative mode
		val := p.getMem(p.base + parameter)
		dbg(2, "   (relmode) -> mem[%d+%d] = %d", p.base, parameter, val)
		return val
	}

	// immediate mode
	dbg(2, "   (immmode) -> = %d", parameter)
	return parameter
}

// Directions.
const (
	Empty = iota
	Wall
	Block
	Paddle
	Ball
)

type Arcade struct {
	cpu       *program
	input     chan int
	output    chan int
	screen    map[Point]int
	ballPos   Point
	paddlePos Point
	score     int
}

type Point struct {
	x, y int
}

var P0 = Point{0, 0}

func init() {
	flag.IntVar(&debug, "debug", 0, "debug level")
	flag.Parse()
}

func newArcade(code string) *Arcade {
	return &Arcade{
		cpu:       newProgram(code),
		input:     make(chan int),
		output:    make(chan int),
		screen:    make(map[Point]int),
		ballPos:   P0,
		paddlePos: P0,
		score:     0,
	}
}

func (a *Arcade) Run() {
	halt := make(chan bool)
	wantinput := make(chan bool)
	go func() {
		a.cpu.setMem(0, 2) // hack for free games!
		a.cpu.run(a.input, a.output, wantinput)
		halt <- true
	}()

	//go func() {
	//	for {
	//		t := a.ballPos.TiltWith(a.paddlePos)
	//		dbg(1, "Tilt: %d (ball: %v, paddle %v)", t, a.ballPos, a.paddlePos)
	//	}
	//}()

	for {
		select {
		case x := <-a.output:
			y := <-a.output
			id := <-a.output
			dbg(1, "-> got coords {%d, %d} -> %d", x, y, id)

			// Gather score
			if x == -1 && y == 0 {
				a.score = id
				//a.input <- a.ballPos.TiltWith(a.paddlePos)
				//dbg(1, "Sent Tilt!")
				//wantinput <- true
				break
			}

			a.screen[Point{x, y}] = id
			switch id {
			case Ball:
				a.ballPos = Point{x, y}
				dbg(1, "Got ball %v", a.ballPos)
			case Paddle:
				a.paddlePos = Point{x, y}
				dbg(1, "Got paddle %v", a.paddlePos)
			}

			a.Paint()
		case <-wantinput:
			a.input <- a.ballPos.TiltWith(a.paddlePos)
			dbg(1, "Sent Tilt!")

		case <-halt:
			return
		}

	}
}

// TiltWidth checks whether p1 is left, right to or aligned with p2. It returns:
// 	-1 if p1 is left of p2
//	 0 if p1 is aligned with p2
//  +1 if p1 is right of p2
func (p1 Point) TiltWith(p2 Point) int {
	switch {
	case p1.x < p2.x:
		return -1
	case p1.x > p2.x:
		return 1
	}

	return 0

}

func (a *Arcade) Paint() {
	if debug < 1 {
		return
	}
	var min, max Point

	for p := range a.screen {
		min = min.Min(p)
		max = max.Max(p)
	}

	for y := min.y; y <= max.y; y++ {
		for x := min.x; x <= max.x; x++ {
			switch a.screen[Point{x, y}] {
			case Empty:
				fmt.Print(" ")
			case Wall:
				fmt.Print("#")
			case Block:
				fmt.Print("%")
			case Paddle:
				fmt.Print("=")
			case Ball:
				fmt.Print("o")
			}
		}
		fmt.Println()
	}
	fmt.Printf("== Score: %3d ==\n", a.score)
}

func (a *Arcade) CountTiles(id int) int {
	count := 0
	for _, t := range a.screen {
		if t == id {
			count++
		}
	}

	return count
}

func mod(a, b int) int {
	return (a%b + b) % b
}

func main() {

	var in string
	fmt.Scan(&in)

	machine := newArcade(in)
	machine.Run()
	machine.Paint()
	blocks := machine.CountTiles(Block)
	fmt.Printf("Part one: %d\n", blocks)
	fmt.Printf("Part two: %d\n", machine.score)

}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func (p Point) Min(p2 Point) Point {
	return Point{
		x: min(p.x, p2.x),
		y: min(p.y, p2.y),
	}
}

func (p Point) Max(p2 Point) Point {
	return Point{
		x: max(p.x, p2.x),
		y: max(p.y, p2.y),
	}
}
