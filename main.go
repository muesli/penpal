package main

import (
	"flag"
	"os"
	"time"

	"github.com/godbus/dbus"
	log "github.com/sirupsen/logrus"
)

var (
	conn *dbus.Conn

	animation = flag.Uint64("animation", 0, "animates drawing with ID")
)

func main() {
	flag.Parse()

	var err error
	conn, err = dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	devs, err := findDevices()
	if err != nil {
		log.Warn("Can't retrieve list of devices. Is tuhi daemon running?")
		log.Fatal(err)
	}
	if len(devs) == 0 {
		log.Error("No paired devices found")
	}

	if *animation > 0 {
		if len(os.Args) < 4 {
			log.Fatal("Usage: -animate ID output.gif")
		}

		drawings, err := findDrawings(devs[0])
		if err != nil {
			panic(err)
		}

		var d uint64
		for _, v := range drawings {
			if v == *animation {
				d = v
			}
		}
		if d == 0 {
			for _, v := range drawings {
				t := time.Unix(int64(v), 0)
				log.Println("Available drawing", v, "-", t.Format("2006-01-02 15:04:05"))

				if v == *animation {
					d = v
				}
			}

			log.Fatal("Could not find drawing with ID ", *animation)
		}

		f, err := os.Create(os.Args[3])
		if err != nil {
			panic(err)
		}
		err = renderAnimation(f, devs[0], d)
		if err != nil {
			panic(err)
		}
		log.Println("Saved animation as:", os.Args[3])
		f.Close()
		return
	}

	for {
		for _, d := range devs {
			log.Println("Found device:", d)

			err = startListening(d)
			if err != nil {
				panic(err)
			}

			err = syncAllDrawings(d)
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(time.Minute)
	}
}
