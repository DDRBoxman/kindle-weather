package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jasonwinn/noaa"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

func main() {
	point := noaa.Point{Latitude: 30.2672, Longitude: -97.7431}
	forecast := point.Forecast(5)

	forecastDay := forecast.ForecastDays[0]

	draw2d.SetFontFolder("./fonts")

	dest := image.NewRGBA(image.Rect(0, 0, 600, 800))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})

	gc.SetFontSize(288)
	gc.SetFontData(draw2d.FontData{
		Name: "kindleweather",
	})
	gc.FillStringAt(getDayFontLetter(forecastDay), 10, 400)

	gc.SetFontSize(32)
	gc.SetFontData(draw2d.FontData{
		Name: "Roboto",
	})
	gc.FillStringAt("Today:", 10, 35)

	gc.FillStringAt("High:", 430, 121)
	gc.FillStringAt("Low:", 430, 275)

	gc.SetFontSize(58)
	hw := gc.FillStringAt(strconv.FormatFloat(forecastDay.MaxTemperature, 'f', -1, 64), 430, 194)
	lw := gc.FillStringAt(strconv.FormatFloat(forecastDay.MinTemperature, 'f', -1, 64), 430, 343)

	gc.SetFontSize(37)
	gc.FillStringAt("°F", 430+hw, 173)
	gc.FillStringAt("°F", 430+lw, 322)

	renderDay(gc, 0, forecast.ForecastDays[1], time.Now().Add(time.Hour*24*2).Weekday().String())
	renderDay(gc, 200, forecast.ForecastDays[2], time.Now().Add(time.Hour*24*2).Weekday().String())
	renderDay(gc, 400, forecast.ForecastDays[3], time.Now().Add(time.Hour*24*3).Weekday().String())

	gc.SetLineWidth(5)
	gc.MoveTo(200, 400)
	gc.LineTo(200, 770)
	gc.Close()

	gc.MoveTo(400, 400)
	gc.LineTo(400, 770)
	gc.Close()

	gc.FillStroke()

	// Save to file
	draw2dimg.SaveToPngFile("weather.png", dest)

	clearDisplay()

	convertPNGImage()

	showImage("weather.jpg")
}

func renderDay(gc *draw2dimg.GraphicContext, offset float64, forecastDay noaa.ForecastDay, weekday string) {
	gc.SetFontData(draw2d.FontData{
		Name: "Roboto",
	})
	gc.SetFontSize(28)
	gc.FillStringAt(weekday+":", offset+10, 440)

	gc.FillStringAt("High:", 20+offset, 650)
	gc.FillStringAt("Low:", 20+offset, 740)

	gc.SetFontSize(40)
	hw := gc.FillStringAt(strconv.FormatFloat(forecastDay.MaxTemperature, 'f', -1, 64), 20+offset, 700)
	lw := gc.FillStringAt(strconv.FormatFloat(forecastDay.MinTemperature, 'f', -1, 64), 20+offset, 790)

	gc.SetFontSize(37)
	gc.FillStringAt("°F", 20+offset+hw, 700)
	gc.FillStringAt("°F", 20+offset+lw, 790)

	gc.SetFontSize(128)
	gc.SetFontData(draw2d.FontData{
		Name: "kindleweather",
	})
	gc.FillStringAt(getDayFontLetter(forecastDay), 20+offset, 600)
}

func convertPNGImage() {
	pngImgFile, _ := os.Open("./weather.png")
	defer pngImgFile.Close()

	imgSrc, _ := png.Decode(pngImgFile)

	newImg := image.NewRGBA(imgSrc.Bounds())

	// we will use white background to replace PNG's transparent background
	// you can change it to whichever color you want with
	// a new color.RGBA{} and use image.NewUniform(color.RGBA{<fill in color>}) function
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// paste PNG image OVER to newImage
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)

	// create new out JPEG file
	jpgImgFile, _ := os.Create("./weather.jpg")

	defer jpgImgFile.Close()

	var opt jpeg.Options
	opt.Quality = 80

	// convert newImage to JPEG encoded byte and save to jpgImgFile
	// with quality = 80
	jpeg.Encode(jpgImgFile, newImg, &opt)
}

func clearDisplay() {
	cmd := exec.Command("eips", "-c")
	cmd.Run()

	cmd = exec.Command("eips", "-c")
	cmd.Run()
}

func showImage(imagePath string) {
	cmd := exec.Command("eips", "-g", imagePath)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func getDayFontLetter(f noaa.ForecastDay) string {
	return weatherFontMapping[getDayIconName(f)]
}

func getDayIconName(f noaa.ForecastDay) string {
	split := strings.Split(f.SummaryDay["icon"], "/")
	split = strings.Split(split[len(split)-1], ".")
	return split[0]
}
