package main

import (
	"flag"
	"fmt"
	gh "github.com/hiromaily/golibs/googlehome"
	lg "github.com/hiromaily/golibs/log"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"math/rand"
	"os"
	"time"
)

const (
	MinimumSize = 150
)

var (
	urlGoogleHome = flag.String("gh", "", "endpoint URL of Google Home device")
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

func init() {
	//command-line
	flag.Parse()

	//log
	lg.InitializeLog(lg.DebugStatus, lg.LogOff, 99,
		"[HumanDetection]", "/tmp/go/human_detection.log")
}

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
		time.Sleep(5 * time.Second)
		isRunning = true
	}()
}
