# kindle-weather

Kindle weather display app that doesn't require a server.

Based off of Matthew Petroff's weather display
https://mpetroff.net/2012/09/kindle-weather-display/

## Photos

![screenshot](https://ddrboxman.github.io/kindle-weather/screenshot.png)
![picture](https://ddrboxman.github.io/kindle-weather/IMG_0608.jpg)

## Requirements

Requires an API key from https://developer.forecast.io/

Only tested on [kindle touch (k5)](http://wiki.mobileread.com/wiki/K5_Index) other kindles may work, let me know.

## How it works

  1. Fetches data from forecast.io
  2. Draws font image with [draw2d](https://github.com/llgcode/draw2d) a vector golang drawing lib
    * We take advantage of a custom made icon font to render the icons smoothly.
  3. Removes the transparent background from that image
  4. Use pngcrush to set the bit depth to 4
  5. Wipe and display the image with eips
