package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"math"
	"os"

	svg "github.com/ajstarks/svgo"
	"github.com/godbus/dbus"
	svgmisc "github.com/rogpeppe/misc/svg"
	log "github.com/sirupsen/logrus"
)

func renderDrawing(w io.Writer, d []*Drawing) {
	s := svg.New(w)
	s.Start(d[0].Dimensions[1]/10, d[0].Dimensions[0]/10)
	s.Rect(0, 0, d[0].Dimensions[1]/10, d[0].Dimensions[0]/10, `fill="white"`)
	s.Gtransform(fmt.Sprintf("translate(%d,0) scale(-1,1)", d[0].Dimensions[1]/10))

	for _, dr := range d {
		for _, stroke := range dr.Strokes {
			path := "M"
			var opacity string
			for i, p := range stroke.Points {
				path += fmt.Sprintf("%.2f,%.2f ", float64(p.Position[1])/10, float64(p.Position[0])/10)
				if i > 0 {
					s.Path(path, `stroke="black" stroke-opacity="`+opacity+`" stroke-width="2" style="fill:none"`)
					path = fmt.Sprintf("M%.2f,%.2f ", float64(p.Position[1])/10, float64(p.Position[0])/10)
				}

				opacity = fmt.Sprintf("%.2f", float64(p.Pressure)/2048)
			}
		}
	}

	s.Gend()
	s.End()
}

func renderDrawingMaxPoints(w io.Writer, d []*Drawing, max uint64) {
	s := svg.New(w)
	s.Start(d[0].Dimensions[1]/10, d[0].Dimensions[0]/10)
	s.Rect(0, 0, d[0].Dimensions[1]/10, d[0].Dimensions[0]/10, `fill="white"`)
	s.Gtransform(fmt.Sprintf("translate(%d,0) scale(-1,1)", d[0].Dimensions[1]/10))

	var pc uint64
	for _, dr := range d {
		for _, stroke := range dr.Strokes {
			path := "M"
			var opacity string
			for i, p := range stroke.Points {
				if pc >= max {
					break
				}
				pc++

				path += fmt.Sprintf("%.2f,%.2f ", float64(p.Position[1])/10, float64(p.Position[0])/10)
				if i > 0 {
					s.Path(path, `stroke="black" stroke-opacity="`+opacity+`" stroke-width="2" style="fill:none"`)
					path = fmt.Sprintf("M%.2f,%.2f ", float64(p.Position[1])/10, float64(p.Position[0])/10)
				}

				opacity = fmt.Sprintf("%.2f", float64(p.Pressure)/2048)
			}
		}
	}

	s.Gend()
	s.End()
}

// renderAnimation retrieves one drawing and renders it as an animated GIF
func renderAnimation(w io.Writer, dev dbus.ObjectPath, drawing uint64) error {
	data, err := fetchDrawing(dev, drawing)
	if err != nil {
		return err
	}

	var d Drawing
	err = json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	var images []*image.Paletted
	var delays []int
	var b []byte
	buf := bytes.NewBuffer(b)

	cp := d.countPoints()
	ss := uint64(math.Max(float64(cp)/100, 1)) // (10 seconds)
	log.Println("Total points:", cp)
	log.Println("Frame points:", ss)

	for steps := uint64(0); steps < cp+ss; steps += ss {
		renderDrawingMaxPoints(buf, []*Drawing{&d}, steps)

		img, err := svgmisc.Render(buf, image.Point{d.Dimensions[1] / 40, d.Dimensions[0] / 40})
		if err != nil {
			return err
		}
		log.Println("Rendered points:", steps)

		palettedImage := image.NewPaletted(img.Bounds(), palette.WebSafe)
		draw.Draw(palettedImage, palettedImage.Rect, img, img.Bounds().Min, draw.Over)

		images = append(images, palettedImage)
		delays = append(delays, 0) //int(p.TOffset/10))

		buf.Reset()
	}

	return gif.EncodeAll(w, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}

// renderSVGPNG renders an SVG as a PNG image
func renderSVGPNG(in string, out string, size image.Point) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := svgmisc.Render(f, size)
	if err != nil {
		return err
	}

	fpng, err := os.Create(out)
	if err != nil {
		return err
	}

	if err := png.Encode(fpng, img); err != nil {
		fpng.Close()
		return err
	}

	if err := fpng.Close(); err != nil {
		return err
	}

	return nil
}
