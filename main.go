package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"os"
	//"time"
)

const MinimumSize = 150

func main() {
	faceDetection()
}

func basic() {
	webcam, _ := gocv.VideoCaptureDevice(0)
	window := gocv.NewWindow("Hello")
	img := gocv.NewMat()

	for {
		webcam.Read(img)
		window.IMShow(img)
		window.WaitKey(1)
	}
}

func faceDetection() {
	deviceID := 0

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	filePath := fmt.Sprintf("%s/src/gocv.io/x/gocv/data/haarcascade_frontalface_default.xml", os.Getenv("GOPATH"))
	//classifier.Load("data/haarcascade_frontalface_default.xml")
	classifier.Load(filePath)

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		//sleep is not good...
		//time.Sleep(1 * time.Second)

		if ok := webcam.Read(img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		//TODO:After detecting face, say Hello by Google Home

		// draw a rectangle around each face on the original image
		for _, r := range rects {
			//when detected face is too small, it should be skipped
			fmt.Printf("x:%d, y:%d, width:%d, height:%d \n", r.Min.X, r.Min.Y, r.Size().X, r.Size().Y)
			if r.Size().X < MinimumSize {
				continue
			}

			//TODO:only the biggest face should be detected.
			gocv.Rectangle(img, r, blue, 3)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)
	}
}
