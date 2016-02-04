package main

import (
	"fmt"
	sdl "github.com/veandco/go-sdl2/sdl"
	ttf "github.com/veandco/go-sdl2/sdl_ttf"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Cell struct {
	x     int32
	y     int32
	size  int32
	val   int
	mark  bool
	fixed bool
}

type Board [81]Cell

type Ctx struct {
	window     *sdl.Window
	renderer   *sdl.Renderer
	board      *Board
	font       *ttf.Font
	charWidth  int
	charHeight int
}

func run() int {
	font, charWidth, charHeight := initFont()

	cellSize := int32(30)
	cellPadding := int32(6)
	startX := int32(20)
	startY := int32(20)

	winSize := int(
		((cellSize + cellPadding) * 10) +
			(cellPadding * 2), // extra padding separating squares
	)

	window := createWindow("Sugoku", winSize, winSize)
	renderer := createRenderer(window)
	board := initBoard(cellSize, startX, startY, cellPadding)
	ctx := Ctx{window, renderer, &board, font, charWidth, charHeight}

	drawBoard(&ctx)

	running := true
	for running {
		if event := sdl.WaitEvent(); event != nil {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				if t.State == 1 {
					clickButton(&ctx, t.X, t.Y)
				}
			case *sdl.KeyDownEvent:
				handleKey(&ctx, t.Keysym)
			}
		}
	}

	close(&ctx)

	return 0
}

func main() {
	os.Exit(run())
}

func initFont() (*ttf.Font, int, int) {
	if err := ttf.Init(); err != nil {
		panic(err)
	}

	font, err := ttf.OpenFont("/Library/Fonts/Courier New.ttf", 16)
	if err != nil {
		panic(err)
	}

	fW, fH, err := font.SizeUTF8("0")
	if err != nil {
		panic(err)
	}

	return font, fW, fH
}

func createWindow(title string, width int, height int) *sdl.Window {
	window, err := sdl.CreateWindow(
		title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		width, height, sdl.WINDOW_SHOWN)

	if err != nil {
		panic(fmt.Sprintf("Failed to create window: %s\n", err))
	}

	return window
}

func createRenderer(window *sdl.Window) *sdl.Renderer {
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		panic(fmt.Sprintf("Failed to create renderer: %s\n", err))
	}

	return renderer
}

func initBoard(size int32, startX int32, startY int32, gap int32) Board {
	var board Board

	i := 0
	initNums := make(map[int]bool, 20)

	// seed rand
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	for _, v := range rnd.Perm(81)[:20] {
		initNums[v] = true
	}

	for y := int32(0); y < 9; y++ {
		sx := startX

		for x := int32(0); x < 9; x++ {
			nx := (x * size) + sx
			ny := (y * size) + startY

			if _, ok := initNums[i]; ok {
				val := randomValueForLine(&board, i)
				board[i] = Cell{nx, ny, size, val, false, true}
			} else {
				board[i] = Cell{nx, ny, size, 0, false, false}
			}

			sx += gap

			if (x+1)%3 == 0 {
				sx += gap
			}

			i += 1
		}

		startY += gap

		if (y+1)%3 == 0 {
			startY += gap
		}
	}

	return board
}

func drawBoard(ctx *Ctx) {
	ctx.renderer.SetDrawColor(0, 0, 0, 0)
	ctx.renderer.Clear()

	highlightValue := getValueToHighlight(ctx.board)

	i := 0
	for i < 81 {
		cell := ctx.board[i]
		rect := sdl.Rect{cell.x, cell.y, cell.size, cell.size}
		innerRect := sdl.Rect{
			cell.x + ((cell.size / 2) - int32(ctx.charWidth/2)),
			cell.y + ((cell.size / 2) - int32(ctx.charHeight/2)),
			int32(ctx.charWidth),
			int32(ctx.charHeight),
		}

		ctx.renderer.SetDrawColor(80, 80, 80, 0)
		ctx.renderer.DrawRect(&rect)
		ctx.renderer.SetDrawColor(0, 0, 0, 0)
		ctx.renderer.DrawRect(&innerRect)

		if cell.mark {
			ctx.renderer.SetDrawColor(80, 80, 80, 0)
		} else if cell.fixed {
			ctx.renderer.SetDrawColor(40, 40, 40, 0)
		}

		if cell.mark || cell.fixed {
			ctx.renderer.FillRect(&rect)
		}

		text := " "
		color := sdl.Color{255, 255, 255, 0}

		if cell.val != 0 {
			text = strconv.Itoa(cell.val)
			isConflict := isConflictingNumber(ctx.board, i, cell.val)

			if !cell.fixed {
				if isConflict {
					color = sdl.Color{255, 0, 0, 0}
				}
			}

			if !isConflict && (cell.val == highlightValue) {
				color = sdl.Color{255, 255, 0, 0}
			}
		}

		textSurface, err := ctx.font.RenderUTF8_Solid(text, color)
		if err != nil {
			panic(err)
		}

		texture, err := ctx.renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			panic(err)
		}

		textSurface.Free()

		ctx.renderer.Copy(texture, nil, &innerRect)

		i += 1
	}

	ctx.renderer.Present()
}

func clickButton(ctx *Ctx, x int32, y int32) {
	i := 0

	for i < 81 {
		cell := ctx.board[i]
		ctx.board[i].mark = false
		if x > cell.x && x < (cell.x+cell.size) {
			if y > cell.y && y < (cell.y+cell.size) {
				ctx.board[i].mark = true
			}
		}
		i += 1
	}

	drawBoard(ctx)
}

func handleKey(ctx *Ctx, key sdl.Keysym) {
	if key.Sym == 8 || (key.Sym >= 49 && key.Sym <= 57) {
		i := 0

		for i < 81 {
			if ctx.board[i].mark {
				if !ctx.board[i].fixed {
					if key.Sym == 8 {
						ctx.board[i].val = 0
					} else {
						ctx.board[i].val = int(key.Sym) - 48
					}
				}
				break
			}

			i += 1
		}

	} else if key.Scancode == sdl.SCANCODE_DOWN ||
		key.Scancode == sdl.SCANCODE_UP ||
		key.Scancode == sdl.SCANCODE_LEFT ||
		key.Scancode == sdl.SCANCODE_RIGHT {

		i := 0
		found := -1

		for i < 81 {
			if ctx.board[i].mark {
				found = i
				break
			}
			i += 1
		}

		if found > -1 {
			mark := found

			if key.Scancode == sdl.SCANCODE_DOWN {
				mark += 9

				if mark > 80 {
					mark = found % 9
				}

			} else if key.Scancode == sdl.SCANCODE_RIGHT {
				mark += 1

				if mark%9 == 0 {
					mark -= 9
				}

			} else if key.Scancode == sdl.SCANCODE_UP {
				mark -= 9

				if mark < 0 {
					mark = (80 - 9) + ((found % 9) + 1)
				}
			} else if key.Scancode == sdl.SCANCODE_LEFT {
				mark -= 1

				if mark < 0 || mark%9 == 8 {
					mark += 9
				}
			}

			ctx.board[i].mark = false
			ctx.board[mark].mark = true

		} else {
			ctx.board[0].mark = true
		}
	}

	drawBoard(ctx)
}

func getConflictingNumbers(board *Board, pos int) [9]bool {
	var numbers [9]bool

	// get numbers from x and y
	x := pos - (pos % 9)
	y := pos % 9

	i := 0
	for i < 9 {
		xp := x + i
		yp := y + (i * 9)

		xv := board[xp].val
		yv := board[yp].val

		if xv > 0 {
			xv -= 1
			numbers[xv] = numbers[xv] || xp != pos
		}

		if yv > 0 {
			yv -= 1
			numbers[yv] = numbers[yv] || yp != pos
		}

		i += 1
	}

	// get numbers from the square
	sx := (pos / 9) / 3
	sy := (pos % 9) / 3

	a := 0
	for a < 3 {
		b := 0
		for b < 3 {
			p := (((sx * 3) + a) * 9) + ((sy * 3) + b)
			bp := board[p]
			if bp.val > 0 {
				pv := bp.val - 1
				numbers[pv] = numbers[pv] || p != pos
			}
			b += 1
		}
		a += 1
	}

	return numbers
}

func isConflictingNumber(board *Board, pos int, number int) bool {
	numbers := getConflictingNumbers(board, pos)
	return numbers[number-1]
}

func randomValueForLine(board *Board, pos int) int {
	numbers := getConflictingNumbers(board, pos)

	// pick one unused random number
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	number := -1

	i := 0
	for i < 9 {
		j := rnd.Intn(9)
		if !numbers[j] {
			number = j + 1
		}
		i += 1
	}

	return number
}

func getValueToHighlight(board *Board) int {
	for _, v := range board {
		if v.mark {
			return v.val
		}
	}

	return -1
}

func close(ctx *Ctx) {
	ctx.window.Destroy()
	ctx.renderer.Destroy()
	ctx.font.Close()
}
