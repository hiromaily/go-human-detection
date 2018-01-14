package main

import (
	"flag"
	"fmt"
	col "github.com/hiromaily/golibs/color"
	gh "github.com/hiromaily/golibs/googlehome"
	lg "github.com/hiromaily/golibs/log"
	"github.com/saljam/mjpeg"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	// minimum of detecting face size
	MinimumSize = 150
	// minimum of detection area size
	MinimumArea = 3000
	// after calling google home, how long google home sleep (second)
	GoogleHomeInterval = 10
)

var (
	mode          = flag.Int("mode", 0, "mode")
	urlGoogleHome = flag.String("gh", "", "endpoint URL of Google Home device")
	port          = flag.Int("port", 8080, "port of web server")
	isRunning     bool
	messages      = []string{
		"Hi, How are you?",
		"Hi, Nice to meet you!",
		"Hi, Long time no see!",
		"Talk to you later.",
		"You look good.",
		"Have a nice day at work!",
	}
)

var usage = `Usage: %s [options...]
Options:
  -mode  Set specific functionality.
     0: Display simple camera
     1: Face Detection
     2: Motion Detection
     3: Web Streamer
  -gh    endpoint URL of Google Home device
  -port  port number of web server when mode:3 Web Streamer
e.g.:
  go-cv -mode 3 -port 8080
`

func init() {
	//command-line
	flag.Usage = func() {
		//fmt.Fprint(os.Stderr, fmt.Sprintf(usage, os.Args[0]))
		fmt.Fprint(os.Stderr, col.Addf(col.Green, usage, os.Args[0]))
	}

	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()

		os.Exit(1)
		return
	}

	//log
	lg.InitializeLog(lg.DebugStatus, lg.LogOff, 99,
		"[HumanDetection]", "/tmp/go/human_detection.log")
}

func main() {
	switch *mode {
	case 0:
		lg.Infof("mode:%d, basic", *mode)
		basic()
	case 1:
		lg.Infof("mode:%d, Face Detection", *mode)
		//face detection
		faceDetection()
	case 2:
		lg.Infof("mode:%d, Motion Detection", *mode)
		//motion detection
		motionDetection()
	case 3:
		lg.Infof("mode:%d, Web Streamer", *mode)
		//web streamer
		webStreamer()
	default:
		lg.Errorf("mode:%d, mode was out of range", *mode)
	}
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

//face detection
//https://github.com/hybridgroup/gocv/tree/master/cmd/facedetect
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

		// draw a rectangle around each face on the original image
		biggest := image.Rectangle{}
		for _, r := range rects {
			//when detected face is too small, it should be skipped
			fmt.Printf("x:%d, y:%d, width:%d, height:%d \n", r.Min.X, r.Min.Y, r.Size().X, r.Size().Y)
			if r.Size().X < MinimumSize {
				continue
			}

			//only the biggest face should be detected.
			if biggest.Size().X < r.Size().X {
				biggest = r
			}
			//gocv.Rectangle(img, r, blue, 3)
		}
		//
		if len(rects) != 0 && biggest.Size().X > 0 {
			// only the biggest face should be detected.
			gocv.Rectangle(img, biggest, blue, 3)

			// After detecting face, say Hello by Google Home
			if *urlGoogleHome != "" && !isRunning {
				callGoogleAPI(*urlGoogleHome)
			}
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)
	}
}

// motion detection
// https://github.com/hybridgroup/gocv/blob/master/cmd/motion-detect/main.go
func motionDetection() {
	deviceID := 0

	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("Motion Window")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	status := "Ready"

	fmt.Printf("Start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(img); !ok {
			fmt.Printf("Error cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		status = "Ready"
		statusColor := color.RGBA{0, 255, 0, 0}

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, imgDelta)

		// remaining cleanup of the image to use for finding contours
		gocv.Threshold(imgDelta, imgThresh, 25, 255, gocv.ThresholdBinary)
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.Dilate(imgThresh, imgThresh, kernel)

		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		for _, c := range contours {
			area := gocv.ContourArea(c)
			if area < MinimumArea {
				continue
			}

			status = "Motion detected"
			statusColor = color.RGBA{255, 0, 0, 0}
			rect := gocv.BoundingRect(c)
			gocv.Rectangle(img, rect, color.RGBA{255, 0, 0, 0}, 2)
		}

		gocv.PutText(img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		window.IMShow(img)
		window.WaitKey(1)
	}
}

// web streamer
// it can be checked on the web
// https://github.com/hybridgroup/gocv/blob/master/cmd/mjpeg-streamer/main.go
func webStreamer() {
	deviceID := 0

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// create the mjpeg stream
	stream := mjpeg.NewStream()

	// start capturing
	go func() {
		for {
			if ok := webcam.Read(img); !ok {
				fmt.Printf("cannot read device %d\n", deviceID)
				return
			}
			if img.Empty() {
				continue
			}

			buf, _ := gocv.IMEncode(".jpg", img)
			//FIXME: WARNING: DATA RACE
			stream.UpdateJPEG(buf)
		}
	}()

	// start http server
	http.Handle("/", stream)
	lg.Infof("Server start with port %d ...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

//It works with Google Home
func callGoogleAPI(ghURL string) {
	//ghURL is expected like "https://xxxxx.ngrok.io/google-home-notifier"

	//choose text from messages
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(messages)) //0 to (len-1)

	//send message
	code, err := gh.SendMessage(ghURL, messages[idx])
	if err != nil {
		lg.Errorf("gh.SendMessage() return error| code:%d, err:%s", code, err)
	}

	go func() {
		time.Sleep(GoogleHomeInterval * time.Second)
		isRunning = true
	}()
}
