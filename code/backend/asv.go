package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"atn/code/backend/internal/signal"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/ivfreader"
	"github.com/spf13/pflag"
)

const (
	osLinux       = "linux"
	osMac         = "darwin"
	osWindows     = "windows"
	ivfFileHandle = ".output.ivf"
	saveRawVideo  = "q"
	doNotTimeOut  = -1
)

type userArguments struct {
	sessionDescription string
	inputVideoPath     string
	inputResolution    string
	runCleanup         bool
	videoIsLive        bool
	operatingSys       string
	ivfHandle          string
	serveOn            string
	stunServers        string
}

func parseArgs() userArguments {
	var inputVideo = pflag.String("video-device", "", "path to the video device or it's name (probably \"FHD Capture\" or /dev/video2)")
	var inputResolutionFlag = pflag.String("input-resolution", "1920x1080", "resolution of camera/input device")
	var runCleanupFlag = pflag.Bool("run-cleanup", true, "clean up leftover output files")
	var stunServer = pflag.String("stun-server", "stun:stun.l.google.com:19302", "stun server to use")
	var serveOnStr = pflag.String("serve-on", ":3000", "port to serve on")

	pflag.Parse()

	return userArguments{
		sessionDescription: "",
		inputVideoPath:     *inputVideo,
		inputResolution:    *inputResolutionFlag,
		runCleanup:         *runCleanupFlag,
		videoIsLive:        true,
		operatingSys:       runtime.GOOS,
		ivfHandle:          ivfFileHandle,
		stunServers:        *stunServer,
		serveOn:            *serveOnStr,
	}
}

func videoDriver(platform string) (string, error) {
	switch platform {
	case osLinux:
		return "video4linux2", nil
	case osMac:
		return "avfoundation", nil
	case osWindows:
		return "dshow", nil

	default:
		return "", fmt.Errorf("%s not supported", platform)
	}
}

func composeStreamCommand(inputDevice, platform, resolution string) (*exec.Cmd, error) {
	driver, _ := videoDriver(platform) // error handled by default case

	switch platform {
	case osWindows:
		fallthrough
	case osLinux:
		return exec.Command("ffmpeg", "-threads", "4", "-y", "-f",
			driver, "-s", resolution, "-i", inputDevice, "-g",
			"30", "-deadline", "realtime", ivfFileHandle), nil
	case osMac:
		return exec.Command("ffmpeg", "-threads", "6", "-probesize",
			"100000000", "-f", driver, "-s", resolution, "-r", "30",
			"-i", inputDevice, "-pix_fmt", "yuv420p", "-g", "1",
			"-deadline", "realtime", "-speed", "16", "-b", "3000k",
			"-an", ivfFileHandle), nil

	default:
		return nil, fmt.Errorf("%s not supported", platform)
	}
}

func cleanUpIvfFile() {
	os.Remove(ivfFileHandle)
}

func retrieveIvfFile(handle string, maxTries int) (*os.File, error) {
	file, ivfErr := os.Open(handle)

	count := 0
	for ivfErr != nil {
		count++
		if count > maxTries && maxTries != doNotTimeOut {
			return nil, fmt.Errorf("timed out trying to open %s", handle)
		}

		if maxTries == doNotTimeOut { // if timeout turned on, they don't need to see attempts
			fmt.Printf("trying to open %s...\n", handle)
		}

		time.Sleep(time.Millisecond * 500)
		file, ivfErr = os.Open(handle)
	}

	return file, nil
}

func retrieveIvfReader(file *os.File, maxTries int) (*ivfreader.IVFReader, *ivfreader.IVFFileHeader, error) {
	ivf, header, ivfErr := ivfreader.NewWith(file)
	count := 0
	for ivfErr != nil {
		count++
		if count > maxTries && maxTries != doNotTimeOut {
			return nil, nil, fmt.Errorf("timed out trying to open ivfreader: %s", ivfErr)
		}

		if maxTries == doNotTimeOut { // track failure if running forever
			fmt.Println("trying to start ivfreader...")
		}

		time.Sleep(time.Millisecond * 500)
		ivf, header, ivfErr = ivfreader.NewWith(file)
	}

	return ivf, header, nil
}

func streamVideo(ivf ivfReader, videoTrack videoMediaTrack, timebaseNum, timebaseDenom float32, ivfFile *os.File, uArgs userArguments) {
	// send video spaced out -- this makes sending less lossy
	defer fmt.Println("")
	sleepTime := time.Millisecond * time.Duration((timebaseNum/timebaseDenom)*1000)
	for {
		frame, _, ivfErr := ivf.ParseNextFrame()
		if ivfErr != nil && uArgs.videoIsLive {
			ivf.ResetReader(func(bytesRead int64) io.Reader {
				ivfFile.Seek(bytesRead, io.SeekStart)
				return ivfFile
			})

			continue
		} else if ivfErr != nil && !uArgs.videoIsLive {
			fmt.Println("")
			return
		}

		time.Sleep(sleepTime)
		if ivfErr = videoTrack.WriteSample(media.Sample{Data: frame, Samples: 90000}); ivfErr != nil {
			continue
		}

		fmt.Print("=")
	}
}

func grabIvfUtilsWithDelay(uArgs userArguments, maxTries int, sleepFor time.Duration) (ivfReader, *ivfreader.IVFFileHeader, *os.File, error) {
	if uArgs.videoIsLive {
		time.Sleep(time.Second * sleepFor) // file might not exist yet
	}

	file, err := retrieveIvfFile(uArgs.ivfHandle, maxTries)

	if err != nil {
		return nil, nil, nil, err
	}

	if uArgs.videoIsLive {
		time.Sleep(time.Second * sleepFor) // it still might not have a complete header written
	}

	ivf, header, err := retrieveIvfReader(file, maxTries)

	if err != nil {
		return nil, nil, nil, err
	}

	return ivf, header, file, nil
}

func videoControl(videoTrack videoMediaTrack, ivfFilePath string, uArgs userArguments) error {
	// Open a IVF file and start reading using our IVFReader
	execStream, err := composeStreamCommand(uArgs.inputVideoPath, uArgs.operatingSys, uArgs.inputResolution)
	if err != nil {
		return err
	}

	execStream.Start()
	defer func() {
		if execStream.Process != nil {
			execStream.Process.Kill()
		}
	}()

	defer cleanUpIvfFile()

	const sleepDuration = 2
	ivf, header, file, _ := grabIvfUtilsWithDelay(uArgs, doNotTimeOut, sleepDuration)

	// Send our video file one frame at a time
	streamVideo(ivf, videoTrack, float32(header.TimebaseNumerator), float32(header.TimebaseDenominator), file, uArgs)

	return nil
}

func run(args userArguments) (string, error) {

	if args.runCleanup {
		cleanUpIvfFile()
	}

	offer := webrtc.SessionDescription{}
	err := signal.Decode(args.sessionDescription, &offer)

	if err != nil {
		return "", fmt.Errorf("decode error: %s", err)
	}

	// We make our own mediaEngine so we can place the sender's codecs in it.  This because we must use the
	// dynamic media type from the sender in our answer. This is not required if we are the offerer
	mediaEngine := webrtc.MediaEngine{}
	err = mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		return "", fmt.Errorf("start media engine error: %s", err)
	}

	// Search for VP8 Payload type. If the offer doesn't support VP8 exit since
	// since they won't be able to decode anything we send them
	var payloadType uint8
	for _, videoCodec := range mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		if videoCodec.Name == "VP8" {
			payloadType = videoCodec.PayloadType
			break
		}
	}

	// Create a new RTCPeerConnection
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{args.stunServers},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// Create a video track
	videoTrack, err := peerConnection.NewTrack(payloadType, rand.Uint32(), "video", "pion")
	if err != nil {
		return "", err
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		return "", err
	}

	go videoControl(videoTrack, ivfFileHandle, args)

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("\nConnection State has changed %s \n", connectionState.String())
	})

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}

	// Output the answer in base64 so we can paste it in browser
	localSdp := signal.Encode(answer)

	return localSdp, nil
}

type bsdp struct {
	BrowserSdp string
}

type ssdp struct {
	ServerSdp string
}

func getBrowserSdp(w http.ResponseWriter, r *http.Request, args userArguments) error {
	var s bsdp
	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &s)
	if err != nil {
		return fmt.Errorf("unmarhsal POST error: %s\n", err)
	}

	args.sessionDescription = s.BrowserSdp
	localSdp, err := run(args)
	if err != nil {
		return fmt.Errorf("run setup error: %s\n", err)
	}

	json.NewEncoder(w).Encode(&ssdp{ServerSdp: localSdp})

	return nil
}

func registerFrontEndHandlers(uArgs userArguments) {
	http.Handle("/", http.FileServer(http.Dir("../frontend/build")))
	http.HandleFunc("/browsersdp", func(w http.ResponseWriter, r *http.Request) {
		err := getBrowserSdp(w, r, uArgs)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
	})

}

func main() {

	args := parseArgs()

	registerFrontEndHandlers(args)

	http.ListenAndServe(args.serveOn, nil)
}
