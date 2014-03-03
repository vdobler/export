package export

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"testing"
)

type Obs struct {
	Age     int
	Origin  string
	Weight  float64
	Height  float64
	Special []byte
}

func (o Obs) BMI() float64 {
	return o.Weight / (o.Height * o.Height)
}

func (o Obs) Group() int {
	return 10*(o.Age/10) + 5
}

func (o Obs) Fancy() (int, error) {
	if o.Height < 1.65 {
		return 0, fmt.Errorf("too small (was %.2f)", o.Height)
	}
	return int(100 * math.Sqrt(o.Height-1.65)), nil
}

func (o Obs) Country() string {
	o2c := map[string]string{
		"ch": "Schweiz",
		"de": "Deutschland",
		"uk": "England",
	}
	return o2c[o.Origin]
}

func (o Obs) Other() bool {
	return true
}

func (o Obs) Other2(a int) int {
	return 0
}

var measurement = []Obs{
	Obs{Age: 20, Origin: "de", Weight: 80, Height: 1.88},
	Obs{Age: 22, Origin: "de", Weight: 85, Height: 1.85},
	Obs{Age: 20, Origin: "de", Weight: 90, Height: 1.95},
	Obs{Age: 25, Origin: "de", Weight: 90, Height: 1.72},

	Obs{Age: 20, Origin: "ch", Weight: 77, Height: 1.78},
	Obs{Age: 20, Origin: "ch", Weight: 82, Height: 1.75},
	Obs{Age: 28, Origin: "ch", Weight: 85, Height: 1.80},
	Obs{Age: 20, Origin: "ch", Weight: 84, Height: 1.62},

	Obs{Age: 31, Origin: "de", Weight: 85, Height: 1.88},
	Obs{Age: 30, Origin: "de", Weight: 90, Height: 1.85},
	Obs{Age: 30, Origin: "de", Weight: 99, Height: 1.95},
	Obs{Age: 42, Origin: "de", Weight: 95, Height: 1.72},

	Obs{Age: 30, Origin: "ch", Weight: 80, Height: 1.78},
	Obs{Age: 30, Origin: "ch", Weight: 85, Height: 1.75},
	Obs{Age: 37, Origin: "ch", Weight: 87, Height: 1.80},
	Obs{Age: 47, Origin: "ch", Weight: 90, Height: 1.62},

	Obs{Age: 42, Origin: "uk", Weight: 60, Height: 1.68},
	Obs{Age: 42, Origin: "uk", Weight: 65, Height: 1.65},
	Obs{Age: 44, Origin: "uk", Weight: 55, Height: 1.52},
	Obs{Age: 44, Origin: "uk", Weight: 70, Height: 1.72},
}

func TestExtractor(t *testing.T) {
	extractor, err := NewExtractor(measurement, "Age", "Origin", "Weight", "BMI", "Fancy")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	fmt.Printf("%v\n", extractor)

	d := CSVDumper{
		Writer:       csv.NewWriter(os.Stdout),
		ShowHeader:   true,
		MissingValue: "NA",
	}

	d.Dump(extractor)
}
