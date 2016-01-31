package main

import (
	"fmt"
	sdl "github.com/veandco/go-sdl2/sdl"
	ttf "github.com/veandco/go-sdl2/sdl_ttf"
	"os"
	"strconv"
)

type Cell struct {
	x    int32
	y    int32
	size int32
	val  int
	mark bool
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

	winSize := int((cellSize + cellPadding) * 10)

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
				handleKey(&ctx, t.Keysym.Sym)
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

	for y := int32(0); y < 9; y++ {
		sx := startX

		for x := int32(0); x < 9; x++ {
			nx := (x * size) + sx
			ny := (y * size) + startY

			board[i] = Cell{nx, ny, size, 0, false}

			sx += gap
			i += 1
		}

		startY += gap
	}

	return board
}

func drawBoard(ctx *Ctx) {
	ctx.renderer.SetDrawColor(0, 0, 0, 0)
	ctx.renderer.Clear()

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
			ctx.renderer.FillRect(&rect)
		}

		text := " "
		if cell.val != 0 {
			text = strconv.Itoa(cell.val)
		}

		color := sdl.Color{255, 255, 255, 255}
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

func handleKey(ctx *Ctx, key sdl.Keycode) {
	i := 0
	for i < 81 {
		if ctx.board[i].mark {
			if key >= 49 && key <= 57 {
				ctx.board[i].val = int(key) - 48
			} else if key == 8 {
				ctx.board[i].val = 0
			}

			ctx.board[i].mark = false
		}
		i += 1
	}

	drawBoard(ctx)
}

func close(ctx *Ctx) {
	ctx.window.Destroy()
	ctx.renderer.Destroy()
	ctx.font.Close()
}
