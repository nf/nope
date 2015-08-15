package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/golang/freetype"
)

const (
	width  = 160
	height = 96
	size   = 50
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

	pt := freetype.Pt(10, 10+int(c.PointToFix32(size)>>8))
	c.DrawString("NOPE", pt)

	sendImage(rgba)
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
