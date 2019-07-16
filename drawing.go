package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/godbus/dbus"
	log "github.com/sirupsen/logrus"
)

var (
	ErrSkipEmpty    = errors.New("skipped seemingly empty drawing")
	ErrSkipExisting = errors.New("skipped existing drawing")
)

// Point represents a single point in a drawing
type Point struct {
	TOffset  int64   `json:"toffset"`
	Position []int64 `json:"position"`
	Pressure int     `json:"pressure"`
}

// Stroke is a path or collection of Points
type Stroke struct {
	Points []Point `json:"points"`
}

// Drawing contains the retrieved JSON data from tuhi
type Drawing struct {
	Version    int      `json:"version"`
	DeviceName string   `json:"devicename"`
	Dimensions []int    `json:"dimensions"`
	Timestamp  int64    `json:"timestamp"`
	Strokes    []Stroke `json:"strokes"`
}

func (d Drawing) countPoints() uint64 {
	var i uint64
	for _, s := range d.Strokes {
		i += uint64(len(s.Points))
	}

	return i
}

// syncDrawings retrieves one or more drawings and merges them into one SVG
func syncDrawings(dev dbus.ObjectPath, drawings []uint64) (string, error) {
	filename := generateFilename(drawings)
	if _, err := os.Stat(filename); err == nil {
		return "", ErrSkipExisting
	}

	dd := []Drawing{}
	for _, drawing := range drawings {
		data, err := fetchDrawing(dev, drawing)
		if err != nil {
			return "", err
		}

		var d Drawing
		err = json.Unmarshal(data, &d)
		if err != nil {
			return "", err
		}

		if cs := d.countPoints(); cs < 24 {
			continue
		}

		dd = append(dd, d)
	}

	if len(dd) == 0 {
		return "", ErrSkipEmpty
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	renderDrawing(f, dd)
	f.Close()

	return filename, nil
}

// syncAllDrawings syncs all drawings on a device
func syncAllDrawings(dev dbus.ObjectPath) error {
	drawings, err := findDrawings(dev)
	if err != nil {
		return err
	}

	// store drawings individually
	for _, d := range drawings {
		t := time.Unix(int64(d), 0)
		log.Println("Found drawing", d, "-", t.Format("2006-01-02 15:04:05"))

		filename, err := syncDrawings(dev, []uint64{d})
		if err != nil {
			if err == ErrSkipEmpty || err == ErrSkipExisting {
				log.Warn(err)
				continue
			}
			return err
		}
		log.Println("Saved drawing as:", filename)
		err = renderSVGPNG(filename, filename+".png", image.Point{1480, 2160})
		if err != nil {
			return err
		}
	}

	// create a merged render of all drawings
	filename, err := syncDrawings(dev, drawings)
	if err != nil {
		if err == ErrSkipEmpty || err == ErrSkipExisting {
			log.Warn(err)
			return nil
		}
		return err
	}
	log.Println("Saved merged render as:", filename)
	err = renderSVGPNG(filename, filename+".png", image.Point{1480, 2160})
	if err != nil {
		return err
	}

	return nil
}

func generateFilename(drawings []uint64) string {
	t1 := time.Unix(int64(drawings[0]), 0)
	t2 := time.Unix(int64(drawings[len(drawings)-1]), 0)

	if !t1.Equal(t2) {
		return fmt.Sprintf("folio_merged_%d_(%s_until_%s).svg",
			len(drawings),
			t1.Format("2006-01-02"),
			t2.Format("2006-01-02"))
	}
	return fmt.Sprintf("folio_%s.svg", t1.Format("2006-01-02_15.04.05"))
}
