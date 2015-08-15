package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	gcolor "github.com/bthomson/go-color"
	"github.com/golang/freetype"
)

const (
	width  = 160
	height = 96
	size   = 24
	dpi    = 72
)

func main() {
	fontBytes, err := ioutil.ReadFile("luxisr.ttf")
	if err != nil {
		log.Fatal(err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	fg, bg := image.White, image.Black
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(font)
	c.SetFontSize(size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(freetype.FullHinting)

	count := 0
	l := NewLife(width, height)
	for {
		if count%10 == 0 {
			count = 0
			pt := freetype.Pt(rand.Intn(width-width/4), height/4+rand.Intn(height-height/4))
			c.DrawString("NOPE", pt)
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					c := rgba.RGBAAt(x, y)
					l.a.Set(x, y, c.R > 0 || c.B > 0 || c.G > 0)
				}
			}
		}
		count++

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				var c color.Color = color.Black
				if l.a.Alive(x, y) {
					c = rainbow(y)
				}
				rgba.Set(x, y, c)
			}
		}

		//fmt.Println(l)
		sendImage(rgba)
		l.Step()

		time.Sleep(time.Second / 8)
	}
}

func rainbow(y int) color.Color {
	h := float64(y) / float64(height)
	rgb := gcolor.HSL{H: h, S: 1, L: 0.5}.ToRGB()
	return color.RGBA{
		R: byte(rgb.R * 255),
		G: byte(rgb.G * 255),
		B: byte(rgb.B * 255),
	}
}

func writeImage(m *image.RGBA) {
	f, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	err = png.Encode(b, m)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}

func sendImage(m *image.RGBA) {
	c, err := net.Dial("udp", "151.217.8.152:6073")
	if err != nil {
		log.Fatal(err)
	}
	uc := c.(*net.UDPConn)
	uc.SetWriteBuffer(width*height*3 + 1)
	var buf bytes.Buffer
	buf.WriteByte(0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := m.RGBAAt(x, y)
			buf.WriteByte(c.R)
			buf.WriteByte(c.G)
			buf.WriteByte(c.B)
		}
	}
	c.Write(buf.Bytes())
}

// An implementation of Conway's Game of Life.

// Field represents a two-dimensional field of cells.
type Field struct {
	s    [][]bool
	w, h int
}

// NewField returns an empty field of the specified width and height.
func NewField(w, h int) *Field {
	s := make([][]bool, h)
	for i := range s {
		s[i] = make([]bool, w)
	}
	return &Field{s: s, w: w, h: h}
}

// Set sets the state of the specified cell to the given value.
func (f *Field) Set(x, y int, b bool) {
	f.s[y][x] = b
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
func (f *Field) Alive(x, y int) bool {
	x += f.w
	x %= f.w
	y += f.h
	y %= f.h
	return f.s[y][x]
}

// Next returns the state of the specified cell at the next time step.
func (f *Field) Next(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && f.Alive(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && f.Alive(x, y)
}

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b *Field
	w, h int
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(w, h int) *Life {
	a := NewField(w, h)
	for i := 0; i < (w * h / 4); i++ {
		a.Set(rand.Intn(w), rand.Intn(h), true)
	}
	return &Life{
		a: a, b: NewField(w, h),
		w: w, h: h,
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.b.Set(x, y, l.a.Next(x, y))
		}
	}
	// Swap fields a and b.
	l.a, l.b = l.b, l.a
}

// String returns the game board as a string.
func (l *Life) String() string {
	var buf bytes.Buffer
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			b := byte(' ')
			if l.a.Alive(x, y) {
				b = '*'
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}
