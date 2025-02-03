package sample

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/pokala15/pcbook/pb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return pb.Keyboard_AZERTY
	case 2:
		return pb.Keyboard_QWERTY
	default:
		return pb.Keyboard_QWERTZ
	}
}

func randomBool() bool {
	return rand.Intn(2) == 1
}

func randomCpuBrand() string {
	return randomStringFromSet("intel", "apple m1")
}

func randomGpuBrand() string {
	return randomStringFromSet("NVIDIA", "AMD")
}

func randomGpuName(brand string) string {
	switch brand {
	case "NVIDIA":
		return randomStringFromSet("RTX 2060",
			"RTX 2070",
			"GTX 1660-Ti",
			"GTX 1070",
		)
	default:
		return randomStringFromSet(
			"RX 590",
			"RX 580",
			"RX 5700-XT",
			"RX Vega-56",
		)
	}
}

func randomStringFromSet(a ...string) string {
	n := len(a)
	if n == 0 {
		return ""
	} else {
		return a[rand.Intn(n)]
	}
}

func randomCpuName(brand string) string {
	switch brand {
	case "intel":
		return randomStringFromSet(
			"Xeon E-2286M",
			"Core i9-9980HK",
			"Core i7-9750H",
			"Core i5-9400F",
			"Core i3-1005G1",
		)
	default:
		return randomStringFromSet(
			"Ryzen 7 PRO 2700U",
			"Ryzen 5 PRO 3500U",
			"Ryzen 3 PRO 3200GE",
		)
	}
}

func randomFloat64(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randomFloat32(min float32, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func randomInt(min int, max int) int {
	return min + rand.Intn(max-min+1)
}

func randomScreenResolution() *pb.Screen_Resolution {
	height := randomInt(1080, 4320)
	width := height * 16 / 9

	resolution := &pb.Screen_Resolution{
		Width:  uint32(width),
		Height: uint32(height),
	}
	return resolution
}

func randomScreenPanel() pb.Screen_Panel {
	if rand.Intn(2) == 1 {
		return pb.Screen_IPS
	}
	return pb.Screen_OLED
}

func randomId() string {
	return uuid.New().String()
}

func randomLaptopBrand() string {
	return randomStringFromSet("Apple", "Dell", "Lenovo")
}

func randLaptopName(brand string) string {
	switch brand {
	case "Apple":
		return randomStringFromSet("Macbook Air", "Macbook Pro")
	case "Dell":
		return randomStringFromSet("Latitude", "Vostro", "XPS", "Alienware")
	default:
		return randomStringFromSet("Thinkpad X1", "Thinkpad P1", "Thinkpad P53")
	}
}
