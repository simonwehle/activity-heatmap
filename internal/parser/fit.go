package parser

import (
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
)

const semicirclesToDegrees = 180.0 / (1 << 31)

type Point struct {
	Lat float64
	Lon float64
}

type Segment struct {
	Points []Point
}

func ParseFITFile(path string) ([]Segment, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := decoder.New(f)
	fit, err := dec.Decode()
	if err != nil {
		return nil, err
	}

	activity := filedef.NewActivity(fit.Messages...)
	var points []Point
	for _, rec := range activity.Records {
		if int32(rec.PositionLat) == 0x7FFFFFFF || int32(rec.PositionLong) == 0x7FFFFFFF {
			continue
		}
		lat := float64(rec.PositionLat) * semicirclesToDegrees
		lon := float64(rec.PositionLong) * semicirclesToDegrees
		if lat == 0 && lon == 0 {
			continue
		}
		points = append(points, Point{Lat: lat, Lon: lon})
	}

	if len(points) < 2 {
		return nil, nil
	}
	return []Segment{{Points: points}}, nil
}