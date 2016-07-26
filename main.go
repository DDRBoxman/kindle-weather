package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	forecast "github.com/mlbright/forecast/v2"
)

func main() {
	renderWeather(fetchWeather())
}

func fetchWeather() *forecast.Forecast {
	keybytes, err := ioutil.ReadFile("api_key.txt")
	if err != nil {
		log.Fatal(err)
	}
	key := string(keybytes)
	key = strings.TrimSpace(key)

	f, err := forecast.Get(key, "30.2672", "-97.7431", "now", forecast.CA)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func renderWeather(forecast *forecast.Forecast) {
	draw2d.SetFontFolder("./fonts")

	dest := image.NewRGBA(image.Rect(0, 0, 600, 800))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})

	gc.SetFontSize(288)
	gc.SetFontData(draw2d.FontData{
		Name: "kindleweather",
	})
	gc.FillStringAt(getDayFontLetter(forecast.Daily.Data[0].Icon), 10, 400)

	gc.SetFontSize(32)
	gc.SetFontData(draw2d.FontData{
		Name: "Roboto",
	})
	gc.FillStringAt("Today:", 10, 35)

	gc.FillStringAt("High:", 430, 121)
	gc.FillStringAt("Low:", 430, 275)

	gc.SetFontSize(58)
	hw := gc.FillStringAt(strconv.FormatFloat(forecast.Daily.Data[0].TemperatureMin, 'f', 0, 64), 430, 194)
	lw := gc.FillStringAt(strconv.FormatFloat(forecast.Daily.Data[0].TemperatureMax, 'f', 0, 64), 430, 343)

	gc.SetFontSize(37)
	gc.FillStringAt("°F", 430+hw, 173)
	gc.FillStringAt("°F", 430+lw, 322)

	renderDay(gc, 0, forecast.Daily.Data[1], time.Now().Add(time.Hour*24*2).Weekday().String())
	renderDay(gc, 200, forecast.Daily.Data[2], time.Now().Add(time.Hour*24*2).Weekday().String())
	renderDay(gc, 400, forecast.Daily.Data[3], time.Now().Add(time.Hour*24*3).Weekday().String())

	gc.SetLineWidth(5)
	gc.MoveTo(200, 400)
	gc.LineTo(200, 770)
	gc.Close()

	gc.MoveTo(400, 400)
	gc.LineTo(400, 770)
	gc.Close()

	gc.FillStroke()

	// Save to file
	draw2dimg.SaveToPngFile("/tmp/weather.png", dest)

	clearDisplay()

	convertPNGImage()

	cmd := exec.Command("./pngcrush", "-bit_depth", "4", "-c", "0", "/tmp/reducedweather.png", "/tmp/reducedweather-crush.png")
	cmd.Run()

	showImage("/tmp/reducedweather-crush.png")
}

func renderDay(gc *draw2dimg.GraphicContext, offset float64, dataPoint forecast.DataPoint, weekday string) {
	gc.SetFontData(draw2d.FontData{
		Name: "Roboto",
	})
	gc.SetFontSize(28)
	gc.FillStringAt(weekday+":", offset+10, 440)

	gc.FillStringAt("High:", 20+offset, 650)
	gc.FillStringAt("Low:", 20+offset, 740)

	gc.SetFontSize(40)
	hw := gc.FillStringAt(strconv.FormatFloat(dataPoint.TemperatureMax, 'f', 0, 64), 20+offset, 700)
	lw := gc.FillStringAt(strconv.FormatFloat(dataPoint.TemperatureMin, 'f', 0, 64), 20+offset, 790)

	gc.SetFontSize(37)
	gc.FillStringAt("°F", 20+offset+hw, 700)
	gc.FillStringAt("°F", 20+offset+lw, 790)

	gc.SetFontSize(128)
	gc.SetFontData(draw2d.FontData{
		Name: "kindleweather",
	})
	gc.FillStringAt(getDayFontLetter(dataPoint.Icon), 20+offset, 600)
}

func convertPNGImage() {
	pngImgFile, _ := os.Open("/tmp/weather.png")
	defer pngImgFile.Close()

	imgSrc, _ := png.Decode(pngImgFile)

	newImg := image.NewRGBA(imgSrc.Bounds())

	// we will use white background to replace PNG's transparent background
	// you can change it to whichever color you want with
	// a new color.RGBA{} and use image.NewUniform(color.RGBA{<fill in color>}) function
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// paste PNG image OVER to newImage
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)

	newPNGFile, _ := os.Create("/tmp/reducedweather.png")
	defer newPNGFile.Close()

	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	encoder.Encode(newPNGFile, newImg)

}

func clearDisplay() {
	cmd := exec.Command("eips", "-c")
	cmd.Run()

	cmd = exec.Command("eips", "-c")
	cmd.Run()
}

func showImage(imagePath string) {
	cmd := exec.Command("eips", "-f", "-g", imagePath)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func getDayFontLetter(icon string) string {
	return weatherFontMapping[icon]
}
