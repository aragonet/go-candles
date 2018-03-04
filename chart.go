package gocandles

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"

	"os"
)

type Options struct {
	LinesChartColor      color.RGBA
	BackgroundChartColor color.RGBA
	PositiveCandleColor  color.RGBA
	NegativeCandleColor  color.RGBA
	PikeCandleColor      color.RGBA
	YLabelText           string
	YLabelColor          color.RGBA
	YLabelSize           int
	Width                int
	Height               int
	CandleWidth          int
	Columns              int
	Rows                 int
	OutputFileName       string
}

var (
	opts                                                     Options
	labelHeight                                              = 16
	higherValue, lowerValue                                  float64
	chartHmin, chartHmax, startTimePosition, endTimePosition int
	startTime                                                time.Time
	timeDiff                                                 time.Duration
)

type Candle struct {
	Date   int64
	High   float64
	Low    float64
	Open   float64
	Close  float64
	Volume float64
}

func (c Candle) getColor() color.RGBA {
	if c.Open > c.Close {
		return opts.NegativeCandleColor
	}
	return opts.PositiveCandleColor
}

func CreateChart(data []Candle, config Options) {
	opts = config
	img := image.NewNRGBA(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.BackgroundChartColor}, image.ZP, draw.Src)

	createAxes(img, data)

	createLabelY(opts.YLabelText, img)

	f, _ := os.OpenFile(opts.OutputFileName, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, img)
}

func createLabelY(label string, img *image.NRGBA) {
	ttf, _ := truetype.Parse(goregular.TTF)
	d := &font.Drawer{
		Face: truetype.NewFace(ttf, &truetype.Options{
			Size:    18,
			DPI:     72,
			Hinting: font.HintingNone,
		}),
	}

	labelYWidth := d.MeasureString(label).Ceil()
	img2 := image.NewNRGBA(image.Rect(0, 0, labelYWidth, labelHeight))
	addLabel(img2, 0, 15, label, 18, opts.YLabelColor)

	e := image.Rect(5, (opts.Height/2)-(labelYWidth/2), 5+labelHeight, labelYWidth*3)
	img2 = imaging.Rotate90(img2)

	draw.Draw(img, e, img2, image.ZP, draw.Over)
}

func createAxes(img *image.NRGBA, data []Candle) {
	chartHmin = opts.Height - 40
	chartHmax = 20
	chartWmax := opts.Width - 30
	separationWithLabel := 5 + 13 + 15

	var max float64
	var min float64
	var maxLength int
	for i, d := range data {
		if i == 0 {
			max = d.High
			min = d.Low
		} else {
			if d.High > max {
				max = d.High
			}
			if d.Low < min {
				min = d.Low
			}
		}
		if len(strconv.Itoa(int(d.High))) > maxLength {
			maxLength = len(strconv.Itoa(int(d.High)))
		}
		if len(strconv.Itoa(int(d.Low))) > maxLength {
			maxLength = len(strconv.Itoa(int(d.Low)))
		}
	}

	chartSeparation := separationWithLabel + maxLength*8 + 8*8
	chartLineWidth := 5

	diff := ((max - min) * 5) / 100
	higherValue = max + diff
	lowerValue = min - diff

	ySeparation := (chartHmin - chartHmax) / opts.Rows
	for i := 0; i < opts.Rows+1; i++ {
		rowLabelValue := higherValue - ((higherValue - lowerValue) / float64(opts.Rows) * float64(i))

		yPosition := chartHmax + int(ySeparation)*i
		line(chartSeparation, chartWmax+chartLineWidth, yPosition, yPosition, opts.LinesChartColor, img)
		addLabel(img, separationWithLabel, yPosition+chartLineWidth, fmt.Sprintf("%.8f", rowLabelValue), 12, opts.LinesChartColor)
	}

	startTime = time.Unix(data[0].Date, 0)
	endTime := time.Unix(data[len(data)-1].Date, 0)

	timeDiff = endTime.Sub(startTime)

	separation := (chartWmax - chartSeparation) / opts.Columns
	for i := 0; i < opts.Columns+1; i++ {
		d := startTime.Add((timeDiff / time.Duration(opts.Columns)) * time.Duration(i))
		dateFormatted := d.Format("_2 Jan")
		timeFormatted := d.Format("15:04")

		xPosition := chartSeparation + chartLineWidth + separation*i
		if i == 0 {
			startTimePosition = xPosition
		} else if i == opts.Columns {
			endTimePosition = xPosition
		}

		line(xPosition, xPosition, chartHmax, chartHmin+3, opts.LinesChartColor, img)
		beautifySpace := 5
		fontSize := 12
		addLabel(img, xPosition-((len(dateFormatted)*7)/2)+beautifySpace, chartHmin+fontSize+beautifySpace, dateFormatted, float64(fontSize), opts.LinesChartColor)
		addLabel(img, xPosition-((len(dateFormatted)*7)/2)+beautifySpace+1, chartHmin+fontSize*2+beautifySpace+2, timeFormatted, float64(fontSize), opts.LinesChartColor)
	}

	for _, d := range data {
		t := time.Unix(d.Date, 0)
		newXPosition := getXPointInChart(t)
		candleColor := d.getColor()
		candleHighYPosition := getYPointInChart(d.High)
		candleOpenYpoint := getYPointInChart(d.Open)
		candleCloseYpoint := getYPointInChart(d.Close)
		candleLowYpoint := getYPointInChart(d.Low)

		if candleColor == opts.PositiveCandleColor {
			aux := candleCloseYpoint
			candleCloseYpoint = candleOpenYpoint
			candleOpenYpoint = aux

		}

		line(newXPosition, newXPosition+1, candleHighYPosition, candleOpenYpoint, opts.PikeCandleColor, img)
		halfCandleWidth := opts.CandleWidth / 2
		line(newXPosition-halfCandleWidth, newXPosition+halfCandleWidth, candleOpenYpoint, candleCloseYpoint, candleColor, img)
		line(newXPosition, newXPosition+1, candleCloseYpoint, candleLowYpoint, opts.PikeCandleColor, img)
	}
}

func getYPointInChart(value float64) int {
	ypercent := ((value - lowerValue) * 100) / (higherValue - lowerValue)
	auxYPoint := float64(chartHmin-chartHmax) * (float64(ypercent) / 100)
	newYPoint := chartHmin - int(auxYPoint)
	return int(newYPoint)
}

func getXPointInChart(value time.Time) int {
	ypercent := ((value.Sub(startTime)) * 100) / (timeDiff)
	auxYPoint := float64(endTimePosition-startTimePosition) * (float64(ypercent) / 100)
	newYPoint := startTimePosition + int(auxYPoint)
	return int(newYPoint)
}

func line(x1, x2, y1, y2 int, col color.RGBA, img *image.NRGBA) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			img.Set(x, y, col)
		}
	}
}

func square(x1, x2, y1, y2, thickness int, col color.RGBA, img *image.NRGBA) {
	line(x1, x2, y1, y1+thickness, col, img)
	line(x1, x2, y2-thickness, y2, col, img)
	line(x1, x1+thickness, y1, y2, col, img)
	line(x2-thickness, x2, y1, y2, col, img)
}

func hlLine(x1, y, x2 int, col color.RGBA, img *image.NRGBA) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

func vLine(x, y1, y2 int, col color.RGBA, img *image.NRGBA) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

func addLabel(img *image.NRGBA, x, y int, label string, size float64, col color.RGBA) {
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
	ttf, _ := truetype.Parse(goregular.TTF)
	d := &font.Drawer{
		Dst: img,
		Src: image.NewUniform(col),
		Face: truetype.NewFace(ttf, &truetype.Options{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingNone,
		}),
		Dot: point,
	}
	d.DrawString(label)
}
