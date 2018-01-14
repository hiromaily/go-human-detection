# go-human-detection
human detection by Open CV3.

## About this project
This depends on [GoCV](https://gocv.io/), [Github](https://github.com/hybridgroup/gocv)  
After face detected, Google Home would speak something.

## Install
Read [GoCV Github](https://github.com/hybridgroup/gocv).

If you'd like to work with Google Home, the below Environment Variable can be used.
```
export GOOGLE_HOME=https://xxx.ngrok.io/google-home-notifier
```

## Features
For now, face detection can send message to Google Home device to speak.

### Face Detection
```
$ go-cv -mode 1 -gh 'https://xxx.ngrok.io/google-home-notifier'
```

### Motion Detection
```
$ go-cv -mode 2
```

### Streaming from web camera on the web. 
```
$ go-cv -mode 3 -port 8080
# It would be helpful to cooperate with ngrok
$ ngrok http 8080 
  ==> http://xxx.ngrok.io
```
