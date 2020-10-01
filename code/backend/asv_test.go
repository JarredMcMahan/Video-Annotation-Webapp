package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/ivfreader"
)

const (
	testIvfFile = "test_data.ivf"
	brokenData  = "broken_data.ivf"
)

func TestCleanUpIvfFile(t *testing.T) {
	os.Create(ivfFileHandle)

	cleanUpIvfFile()

	if os.Remove(ivfFileHandle) == nil {
		t.Error("os.Remove(ivfFileHandle) didn't return an error; cleanup failed")
	}
}

func TestRetrieveIvfFile(t *testing.T) {
	_, err := retrieveIvfFile("not_a_real_file", 2)

	if err == nil {
		t.Error("retrieveIvfFile should have failed but did not")
	}

	_, err = retrieveIvfFile(testIvfFile, -1)

	if err != nil {
		t.Errorf("retrieveIvfFile should have worked with %s but failed", testIvfFile)
	}
}

func TestRetrieveIvfReader(t *testing.T) {
	f, err := retrieveIvfFile(testIvfFile, 1)
	if err != nil {
		t.Errorf("error in setup: %s", err)
	}

	_, _, err = retrieveIvfReader(f, -1)

	if err != nil {
		t.Errorf("error reading %s: %s", testIvfFile, err)
	}

	f.Truncate(10) // not entirely sure what "10" represents, but this corrupts the data
	_, _, err = retrieveIvfReader(f, 1)
	if err == nil {
		t.Error("retrieveIvfFile did not error when it should have")
	}
}

func checkEq(one, two []string) error {
	if len(one) != len(two) {
		return fmt.Errorf("len(one) = %d != len(two) = %d", len(one), len(two))
	}

	for i := 0; i < len(one); i++ {
		if one[i] != two[i] {
			return fmt.Errorf("one: %s, two: %s", one[i], two[i])
		}
	}

	return nil
}

func TestComposeStreamCommandMac(t *testing.T) {
	mac, err := composeStreamCommand("device", osMac, "1920x1080")
	if err != nil {
		t.Errorf("%s", err)
	}

	if !strings.Contains(mac.Path, "ffmpeg") {
		t.Errorf("path for %s is %s should be ffmpeg", osMac, mac.Path)
	}

	driver, err := videoDriver(osMac)

	expectedCommand := []string{"ffmpeg", "-threads", "6", "-probesize",
		"100000000", "-f", driver, "-s",
		"1920x1080", "-r", "30", "-i", "device",
		"-pix_fmt", "yuv420p", "-g", "1", "-deadline",
		"realtime", "-speed", "16", "-b", "3000k",
		"-an", ivfFileHandle}

	if err = checkEq(mac.Args, expectedCommand); err != nil {
		t.Errorf("command for %s is wrong; %s", osMac, err)
	}

}

func TestComposeStreamCommandWindows(t *testing.T) {
	windows, err := composeStreamCommand("device", osWindows, "1920x1080")
	if err != nil {
		t.Errorf("%s", err)
	}

	if !strings.Contains(windows.Path, "ffmpeg") {
		t.Errorf("path is %s should be ffmpeg", windows.Path)
	}

	driver, err := videoDriver(osWindows)

	expectedCommand := []string{"ffmpeg", "-threads", "4", "-y",
		"-f", driver, "-s", "1920x1080", "-i",
		"device", "-g", "30", "-deadline",
		"realtime", ivfFileHandle}

	if err = checkEq(windows.Args, expectedCommand); err != nil {
		t.Errorf("command for %s is wrong; %s", osWindows, err)
	}

}

func TestComposeStreamCommandLinux(t *testing.T) {
	linux, err := composeStreamCommand("device", osLinux, "1920x1080")
	if err != nil {
		t.Errorf("%s", err)
	}

	if !strings.Contains(linux.Path, "ffmpeg") {
		t.Errorf("path is %s should be ffmpeg", linux.Path)
	}

	driver, err := videoDriver(osLinux)

	expectedCommand := []string{"ffmpeg", "-threads", "4", "-y",
		"-f", driver, "-s", "1920x1080", "-i",
		"device", "-g", "30", "-deadline",
		"realtime", ivfFileHandle}

	if err = checkEq(linux.Args, expectedCommand); err != nil {
		t.Errorf("command for %s is wrong; %s", osLinux, err)
	}
}

func TestComposeStreamCommandFailure(t *testing.T) {
	badCommand, err := composeStreamCommand("device", "not a platform", "1920x1080")
	if err == nil {
		t.Errorf("invalid platform allowed in command: %s", badCommand.Args)
	}
}

type mockVideoTrackReturningNil struct{}

func (mv mockVideoTrackReturningNil) WriteSample(s media.Sample) error {
	return nil
}

type mockVideoTrackReturningError struct{}

func (mv mockVideoTrackReturningError) WriteSample(s media.Sample) error {
	return fmt.Errorf("error")
}

func TestVideoControl(t *testing.T) {
	madeUpArgs := userArguments{}

	madeUpArgs.inputVideoPath = testIvfFile
	madeUpArgs.operatingSys = runtime.GOOS
	madeUpArgs.inputResolution = "1920x1080"
	madeUpArgs.ivfHandle = testIvfFile
	madeUpArgs.videoIsLive = false

	err := videoControl(mockVideoTrackReturningNil{}, testIvfFile, madeUpArgs)
	if err != nil {
		t.Errorf("videoControl failed with %s", err)
	}

	err = videoControl(mockVideoTrackReturningError{}, testIvfFile, madeUpArgs)
	if err != nil {
		t.Errorf("videoControl failed with %s", err)
	}

	madeUpArgs.operatingSys = "not a real device"
	err = videoControl(mockVideoTrackReturningError{}, testIvfFile, madeUpArgs)
	if err == nil {
		t.Errorf("videoControl passed with fake device named: %s", madeUpArgs.operatingSys)
	}
}

type mockIvfRead struct{}

func (mIvfR mockIvfRead) ParseNextFrame() ([]byte, *ivfreader.IVFFrameHeader, error) {
	return nil, nil, fmt.Errorf("error")
}

func (mIvfR mockIvfRead) ResetReader(reset func(bytesRead int64) io.Reader) {
	_ = reset(0)
	panic("gotta get out of the endless function")
}

func TestStreamVideo(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("got %s, but err should be nil", err)
		}
	}()

	f, _ := os.Open(testIvfFile)
	madeUpArgs := userArguments{}
	madeUpArgs.videoIsLive = true
	streamVideo(mockIvfRead{}, mockVideoTrackReturningNil{}, 10.0, 10.0, f, madeUpArgs)
}

func TestParseArgs(t *testing.T) {
	os.Args = []string{"asv", "--video-device=4321", "--input-resolution=720p",
		"--run-cleanup=false"}

	args := parseArgs()

	if args.inputVideoPath != "4321" {
		t.Errorf("inputVideoPath failed with %s, should be %s", args.inputVideoPath, "4321")
	}

	if args.inputResolution != "720p" {
		t.Errorf("inputResolution failed with %s, should be %s", args.inputResolution, "720p")
	}

	if args.runCleanup != false { // this is written like this because I want to be clear about the assertion
		t.Errorf("runCleanup failed with %t, should be %t", args.runCleanup, false)
	}

	if args.videoIsLive != true {
		t.Errorf("videoIsLive failed with %t, should be %t", args.videoIsLive, true)
	}
}

func TestRun(t *testing.T) {
	validSessionDescription := "eyJ0eXBlIjoib2ZmZXIiLCJzZHAiOiJ2PTBcclxubz1tb3ppbGxhLi4uVEhJU19JU19TRFBBUlRBLTcyLjAuMSA0MzU3NzU1ODA1NzczMjE3MTAgMCBJTiBJUDQgMC4wLjAuMFxyXG5zPS1cclxudD0wIDBcclxuYT1zZW5kcmVjdlxyXG5hPWZpbmdlcnByaW50OnNoYS0yNTYgMTI6NUE6RkI6Qjc6N0U6ODY6QkM6RjA6RTI6OTU6QjQ6Q0Y6QTA6QUM6RDY6QzQ6QkM6REY6MUE6NDQ6NEE6OEY6RkY6RDE6NDA6MEY6RTI6RDA6Mjg6NjU6Rjk6QTlcclxuYT1ncm91cDpCVU5ETEUgMFxyXG5hPWljZS1vcHRpb25zOnRyaWNrbGVcclxuYT1tc2lkLXNlbWFudGljOldNUyAqXHJcbm09dmlkZW8gNTQ0NDEgVURQL1RMUy9SVFAvU0FWUEYgMTIwIDEyMVxyXG5jPUlOIElQNCA5OC4yMDAuMjQzLjE1MVxyXG5hPWNhbmRpZGF0ZTowIDEgVURQIDIxMjIyNTI1NDMgMTkyLjE2OC4wLjExMiA1NDQ0MSB0eXAgaG9zdFxyXG5hPWNhbmRpZGF0ZToyIDEgVENQIDIxMDU1MjQ0NzkgMTkyLjE2OC4wLjExMiA5IHR5cCBob3N0IHRjcHR5cGUgYWN0aXZlXHJcbmE9Y2FuZGlkYXRlOjAgMiBVRFAgMjEyMjI1MjU0MiAxOTIuMTY4LjAuMTEyIDQ4MjU4IHR5cCBob3N0XHJcbmE9Y2FuZGlkYXRlOjIgMiBUQ1AgMjEwNTUyNDQ3OCAxOTIuMTY4LjAuMTEyIDkgdHlwIGhvc3QgdGNwdHlwZSBhY3RpdmVcclxuYT1jYW5kaWRhdGU6MSAxIFVEUCAxNjg2MDUyODYzIDk4LjIwMC4yNDMuMTUxIDU0NDQxIHR5cCBzcmZseCByYWRkciAxOTIuMTY4LjAuMTEyIHJwb3J0IDU0NDQxXHJcbmE9Y2FuZGlkYXRlOjEgMiBVRFAgMTY4NjA1Mjg2MiA5OC4yMDAuMjQzLjE1MSA0ODI1OCB0eXAgc3JmbHggcmFkZHIgMTkyLjE2OC4wLjExMiBycG9ydCA0ODI1OFxyXG5hPXNlbmRyZWN2XHJcbmE9ZW5kLW9mLWNhbmRpZGF0ZXNcclxuYT1leHRtYXA6MyB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDpzZGVzOm1pZFxyXG5hPWV4dG1hcDo0IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L2Ficy1zZW5kLXRpbWVcclxuYT1leHRtYXA6NSB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDp0b2Zmc2V0XHJcbmE9ZXh0bWFwOjYvcmVjdm9ubHkgaHR0cDovL3d3dy53ZWJydGMub3JnL2V4cGVyaW1lbnRzL3J0cC1oZHJleHQvcGxheW91dC1kZWxheVxyXG5hPWZtdHA6MTIwIG1heC1mcz0xMjI4ODttYXgtZnI9NjBcclxuYT1mbXRwOjEyMSBtYXgtZnM9MTIyODg7bWF4LWZyPTYwXHJcbmE9aWNlLXB3ZDozMmE5NDcyZWVlMTllMmRiM2QwNmI0ODA5NWFmYTk1ZFxyXG5hPWljZS11ZnJhZzo3NGRjMDdiNVxyXG5hPW1pZDowXHJcbmE9bXNpZDotIHtkZTMwMmU5Yi1mYTE2LTQ2NzYtOGNjMy1hMDg2ZjljOWExYWJ9XHJcbmE9cnRjcDo0ODI1OCBJTiBJUDQgOTguMjAwLjI0My4xNTFcclxuYT1ydGNwLWZiOjEyMCBuYWNrXHJcbmE9cnRjcC1mYjoxMjAgbmFjayBwbGlcclxuYT1ydGNwLWZiOjEyMCBjY20gZmlyXHJcbmE9cnRjcC1mYjoxMjAgZ29vZy1yZW1iXHJcbmE9cnRjcC1mYjoxMjEgbmFja1xyXG5hPXJ0Y3AtZmI6MTIxIG5hY2sgcGxpXHJcbmE9cnRjcC1mYjoxMjEgY2NtIGZpclxyXG5hPXJ0Y3AtZmI6MTIxIGdvb2ctcmVtYlxyXG5hPXJ0Y3AtbXV4XHJcbmE9cnRwbWFwOjEyMCBWUDgvOTAwMDBcclxuYT1ydHBtYXA6MTIxIFZQOS85MDAwMFxyXG5hPXNldHVwOmFjdHBhc3NcclxuYT1zc3JjOjQ1NjIyNTg3NiBjbmFtZTp7YjdlNjVmZDYtYjA1OC00NDcwLTg5ZDItYjU4ODBlODUwMGE4fVxyXG4ifQ=="

	invalidSD := validSessionDescription[:len(validSessionDescription)-6]

	mockArgs := userArguments{
		sessionDescription: validSessionDescription,
		inputVideoPath:     "",
		inputResolution:    "1920x1080",
		runCleanup:         true,
		videoIsLive:        true,
		operatingSys:       osLinux,
		ivfHandle:          testIvfFile,
		stunServers:        "stun:stun.l.google.com:19302",
	}

	_, err := run(mockArgs)
	if err != nil {
		t.Errorf("run failed with %s", err)
	}

	mockArgs.sessionDescription = invalidSD
	_, err = run(mockArgs)
	if err == nil {
		t.Errorf("run passed but should have failed, SDP")
	}

	mockArgs.sessionDescription = validSessionDescription
	mockArgs.stunServers = ""
	_, err = run(mockArgs)
	if err == nil {
		t.Errorf("run passed but should have failed, stunServer")
	}
}

func TestGrabIvfUtilsWithDelay(t *testing.T) {
	mockArgs := userArguments{
		videoIsLive: true,
		ivfHandle:   testIvfFile,
	}

	_, _, _, err := grabIvfUtilsWithDelay(mockArgs, 1, 0)
	if err != nil {
		t.Errorf("err of %s", err)
	}

	mockArgs.ivfHandle = "not a real file"
	_, _, _, err = grabIvfUtilsWithDelay(mockArgs, 1, 0)
	if err == nil {
		t.Errorf("no error while ivfHandle was not real")
	}

	mockArgs.ivfHandle = brokenData
	_, _, _, err = grabIvfUtilsWithDelay(mockArgs, 1, 0)
	if err == nil {
		t.Errorf("no error while ivfHandle was truncated data")
	}
}

func TestRegisterFrontEndHandlers(t *testing.T) {
	mockArgs := userArguments{
		sessionDescription: "",
		inputVideoPath:     "",
		inputResolution:    "1920x1080",
		runCleanup:         true,
		videoIsLive:        true,
		operatingSys:       osLinux,
		ivfHandle:          testIvfFile,
		stunServers:        "stun:stun.l.google.com:19302",
	}

	registerFrontEndHandlers(mockArgs)
}

func TestGetBrowserSdp(t *testing.T) {
	mockArgs := userArguments{
		sessionDescription: "",
		inputVideoPath:     "",
		inputResolution:    "1920x1080",
		runCleanup:         true,
		videoIsLive:        true,
		operatingSys:       osLinux,
		ivfHandle:          testIvfFile,
		stunServers:        "stun:stun.l.google.com:19302",
	}

	validSessionDescription := "{\"BrowserSdp\": \"eyJ0eXBlIjoib2ZmZXIiLCJzZHAiOiJ2PTBcclxubz1tb3ppbGxhLi4uVEhJU19JU19TRFBBUlRBLTcyLjAuMSA0MzU3NzU1ODA1NzczMjE3MTAgMCBJTiBJUDQgMC4wLjAuMFxyXG5zPS1cclxudD0wIDBcclxuYT1zZW5kcmVjdlxyXG5hPWZpbmdlcnByaW50OnNoYS0yNTYgMTI6NUE6RkI6Qjc6N0U6ODY6QkM6RjA6RTI6OTU6QjQ6Q0Y6QTA6QUM6RDY6QzQ6QkM6REY6MUE6NDQ6NEE6OEY6RkY6RDE6NDA6MEY6RTI6RDA6Mjg6NjU6Rjk6QTlcclxuYT1ncm91cDpCVU5ETEUgMFxyXG5hPWljZS1vcHRpb25zOnRyaWNrbGVcclxuYT1tc2lkLXNlbWFudGljOldNUyAqXHJcbm09dmlkZW8gNTQ0NDEgVURQL1RMUy9SVFAvU0FWUEYgMTIwIDEyMVxyXG5jPUlOIElQNCA5OC4yMDAuMjQzLjE1MVxyXG5hPWNhbmRpZGF0ZTowIDEgVURQIDIxMjIyNTI1NDMgMTkyLjE2OC4wLjExMiA1NDQ0MSB0eXAgaG9zdFxyXG5hPWNhbmRpZGF0ZToyIDEgVENQIDIxMDU1MjQ0NzkgMTkyLjE2OC4wLjExMiA5IHR5cCBob3N0IHRjcHR5cGUgYWN0aXZlXHJcbmE9Y2FuZGlkYXRlOjAgMiBVRFAgMjEyMjI1MjU0MiAxOTIuMTY4LjAuMTEyIDQ4MjU4IHR5cCBob3N0XHJcbmE9Y2FuZGlkYXRlOjIgMiBUQ1AgMjEwNTUyNDQ3OCAxOTIuMTY4LjAuMTEyIDkgdHlwIGhvc3QgdGNwdHlwZSBhY3RpdmVcclxuYT1jYW5kaWRhdGU6MSAxIFVEUCAxNjg2MDUyODYzIDk4LjIwMC4yNDMuMTUxIDU0NDQxIHR5cCBzcmZseCByYWRkciAxOTIuMTY4LjAuMTEyIHJwb3J0IDU0NDQxXHJcbmE9Y2FuZGlkYXRlOjEgMiBVRFAgMTY4NjA1Mjg2MiA5OC4yMDAuMjQzLjE1MSA0ODI1OCB0eXAgc3JmbHggcmFkZHIgMTkyLjE2OC4wLjExMiBycG9ydCA0ODI1OFxyXG5hPXNlbmRyZWN2XHJcbmE9ZW5kLW9mLWNhbmRpZGF0ZXNcclxuYT1leHRtYXA6MyB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDpzZGVzOm1pZFxyXG5hPWV4dG1hcDo0IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L2Ficy1zZW5kLXRpbWVcclxuYT1leHRtYXA6NSB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDp0b2Zmc2V0XHJcbmE9ZXh0bWFwOjYvcmVjdm9ubHkgaHR0cDovL3d3dy53ZWJydGMub3JnL2V4cGVyaW1lbnRzL3J0cC1oZHJleHQvcGxheW91dC1kZWxheVxyXG5hPWZtdHA6MTIwIG1heC1mcz0xMjI4ODttYXgtZnI9NjBcclxuYT1mbXRwOjEyMSBtYXgtZnM9MTIyODg7bWF4LWZyPTYwXHJcbmE9aWNlLXB3ZDozMmE5NDcyZWVlMTllMmRiM2QwNmI0ODA5NWFmYTk1ZFxyXG5hPWljZS11ZnJhZzo3NGRjMDdiNVxyXG5hPW1pZDowXHJcbmE9bXNpZDotIHtkZTMwMmU5Yi1mYTE2LTQ2NzYtOGNjMy1hMDg2ZjljOWExYWJ9XHJcbmE9cnRjcDo0ODI1OCBJTiBJUDQgOTguMjAwLjI0My4xNTFcclxuYT1ydGNwLWZiOjEyMCBuYWNrXHJcbmE9cnRjcC1mYjoxMjAgbmFjayBwbGlcclxuYT1ydGNwLWZiOjEyMCBjY20gZmlyXHJcbmE9cnRjcC1mYjoxMjAgZ29vZy1yZW1iXHJcbmE9cnRjcC1mYjoxMjEgbmFja1xyXG5hPXJ0Y3AtZmI6MTIxIG5hY2sgcGxpXHJcbmE9cnRjcC1mYjoxMjEgY2NtIGZpclxyXG5hPXJ0Y3AtZmI6MTIxIGdvb2ctcmVtYlxyXG5hPXJ0Y3AtbXV4XHJcbmE9cnRwbWFwOjEyMCBWUDgvOTAwMDBcclxuYT1ydHBtYXA6MTIxIFZQOS85MDAwMFxyXG5hPXNldHVwOmFjdHBhc3NcclxuYT1zc3JjOjQ1NjIyNTg3NiBjbmFtZTp7YjdlNjVmZDYtYjA1OC00NDcwLTg5ZDItYjU4ODBlODUwMGE4fVxyXG4ifQ==\"}"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := http.NewRequest("POST", "/browsersdp", strings.NewReader(validSessionDescription))
		if err != nil {
			t.Errorf("NewRequest error: %s", err)
		}

		err = getBrowserSdp(w, r, mockArgs)
		if err != nil {
			t.Errorf("error in server response: %s", err)
		}
	}))

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("%s", err)
	}
}

func TestGetBrowserSdpInvalidInputs(t *testing.T) {
	mockArgs := userArguments{
		sessionDescription: "",
		inputVideoPath:     "",
		inputResolution:    "1920x1080",
		runCleanup:         true,
		videoIsLive:        true,
		operatingSys:       osLinux,
		ivfHandle:          testIvfFile,
		stunServers:        "stun:stun.l.google.com:19302",
	}

	validSessionDescription := "{\"BrowserSdp\": \"eyJ0eXBlIjoib2ZmZXIiLCJzZHAiOiJ2PTBcclxubz1tb3ppbGxhLi4uVEhJU19JU19TRFBBUlRBLTcyLjAuMSA0MzU3NzU1ODA1NzczMjE3MTAgMCBJTiBJUDQgMC4wLjAuMFxyXG5zPS1cclxudD0wIDBcclxuYT1zZW5kcmVjdlxyXG5hPWZpbmdlcnByaW50OnNoYS0yNTYgMTI6NUE6RkI6Qjc6N0U6ODY6QkM6RjA6RTI6OTU6QjQ6Q0Y6QTA6QUM6RDY6QzQ6QkM6REY6MUE6NDQ6NEE6OEY6RkY6RDE6NDA6MEY6RTI6RDA6Mjg6NjU6Rjk6QTlcclxuYT1ncm91cDpCVU5ETEUgMFxyXG5hPWljZS1vcHRpb25zOnRyaWNrbGVcclxuYT1tc2lkLXNlbWFudGljOldNUyAqXHJcbm09dmlkZW8gNTQ0NDEgVURQL1RMUy9SVFAvU0FWUEYgMTIwIDEyMVxyXG5jPUlOIElQNCA5OC4yMDAuMjQzLjE1MVxyXG5hPWNhbmRpZGF0ZTowIDEgVURQIDIxMjIyNTI1NDMgMTkyLjE2OC4wLjExMiA1NDQ0MSB0eXAgaG9zdFxyXG5hPWNhbmRpZGF0ZToyIDEgVENQIDIxMDU1MjQ0NzkgMTkyLjE2OC4wLjExMiA5IHR5cCBob3N0IHRjcHR5cGUgYWN0aXZlXHJcbmE9Y2FuZGlkYXRlOjAgMiBVRFAgMjEyMjI1MjU0MiAxOTIuMTY4LjAuMTEyIDQ4MjU4IHR5cCBob3N0XHJcbmE9Y2FuZGlkYXRlOjIgMiBUQ1AgMjEwNTUyNDQ3OCAxOTIuMTY4LjAuMTEyIDkgdHlwIGhvc3QgdGNwdHlwZSBhY3RpdmVcclxuYT1jYW5kaWRhdGU6MSAxIFVEUCAxNjg2MDUyODYzIDk4LjIwMC4yNDMuMTUxIDU0NDQxIHR5cCBzcmZseCByYWRkciAxOTIuMTY4LjAuMTEyIHJwb3J0IDU0NDQxXHJcbmE9Y2FuZGlkYXRlOjEgMiBVRFAgMTY4NjA1Mjg2MiA5OC4yMDAuMjQzLjE1MSA0ODI1OCB0eXAgc3JmbHggcmFkZHIgMTkyLjE2OC4wLjExMiBycG9ydCA0ODI1OFxyXG5hPXNlbmRyZWN2XHJcbmE9ZW5kLW9mLWNhbmRpZGF0ZXNcclxuYT1leHRtYXA6MyB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDpzZGVzOm1pZFxyXG5hPWV4dG1hcDo0IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L2Ficy1zZW5kLXRpbWVcclxuYT1leHRtYXA6NSB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDp0b2Zmc2V0XHJcbmE9ZXh0bWFwOjYvcmVjdm9ubHkgaHR0cDovL3d3dy53ZWJydGMub3JnL2V4cGVyaW1lbnRzL3J0cC1oZHJleHQvcGxheW91dC1kZWxheVxyXG5hPWZtdHA6MTIwIG1heC1mcz0xMjI4ODttYXgtZnI9NjBcclxuYT1mbXRwOjEyMSBtYXgtZnM9MTIyODg7bWF4LWZyPTYwXHJcbmE9aWNlLXB3ZDozMmE5NDcyZWVlMTllMmRiM2QwNmI0ODA5NWFmYTk1ZFxyXG5hPWljZS11ZnJhZzo3NGRjMDdiNVxyXG5hPW1pZDowXHJcbmE9bXNpZDotIHtkZTMwMmU5Yi1mYTE2LTQ2NzYtOGNjMy1hMDg2ZjljOWExYWJ9XHJcbmE9cnRjcDo0ODI1OCBJTiBJUDQgOTguMjAwLjI0My4xNTFcclxuYT1ydGNwLWZiOjEyMCBuYWNrXHJcbmE9cnRjcC1mYjoxMjAgbmFjayBwbGlcclxuYT1ydGNwLWZiOjEyMCBjY20gZmlyXHJcbmE9cnRjcC1mYjoxMjAgZ29vZy1yZW1iXHJcbmE9cnRjcC1mYjoxMjEgbmFja1xyXG5hPXJ0Y3AtZmI6MTIxIG5hY2sgcGxpXHJcbmE9cnRjcC1mYjoxMjEgY2NtIGZpclxyXG5hPXJ0Y3AtZmI6MTIxIGdvb2ctcmVtYlxyXG5hPXJ0Y3AtbXV4XHJcbmE9cnRwbWFwOjEyMCBWUDgvOTAwMDBcclxuYT1ydHBtYXA6MTIxIFZQOS85MDAwMFxyXG5hPXNldHVwOmFjdHBhc3NcclxuYT1zc3JjOjQ1NjIyNTg3NiBjbmFtZTp7YjdlNjVmZDYtYjA1OC00NDcwLTg5ZDItYjU4ODBlODUwMGE4fVxyXG4ifQ=="

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := http.NewRequest("POST", "/browsersdp", strings.NewReader(validSessionDescription))
		if err != nil {
			t.Errorf("NewRequest error: %s", err)
		}

		err = getBrowserSdp(w, r, mockArgs)
		if err == nil {
			t.Errorf("no error in server response")
		}
	}))

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("%s", err)
	}
}

func TestGetBrowserSdpInvalidSdp(t *testing.T) {
	mockArgs := userArguments{
		sessionDescription: "",
		inputVideoPath:     "",
		inputResolution:    "1920x1080",
		runCleanup:         true,
		videoIsLive:        true,
		operatingSys:       osLinux,
		ivfHandle:          testIvfFile,
		stunServers:        "stun:stun.l.google.com:19302",
	}

	validSessionDescription := "{\"BrowserSdp\": \"eyJ0eXBlIjoib2ZmZXIiLCJzZHAiOiJ2PTBcclxubz1tb3ppbGxhLi4uVEhJU19JU19TRFBBUlRBLTcyLjAuMSA0MzU3NzU1ODA1NzczMjE3MTAgMCBJTiBJUDQgMC4wLjAuMFxyXG5zPS1cclxudD0wIDBcclxuYT1zZW5kcmVjdlxyXG5hPWZpbmdlcnByaW50OnNoYS0yNTYgMTI6NUE6RkI6Qjc6N0U6ODY6QkM6RjA6RTI6OTU6QjQ6Q0Y6QTA6QUM6RDY6QzQ6QkM6REY6MUE6NDQ6NEE6OEY6RkY6RDE6NDA6MEY6RTI6RDA6Mjg6NjU6Rjk6QTlcclxuYT1ncm91cDpCVU5ETEUgMFxyXG5hPWljZS1vcHRpb25zOnRyaWNrbGVcclxuYT1tc2lkLXNlbWFudGljOldNUyAqXHJcbm09dmlkZW8gNTQ0NDEgVURQL1RMUy9SVFAvU0FWUEYgMTIwIDEyMVxyXG5jPUlOIElQNCA5OC4yMDAuMjQzLjE1MVxyXG5hPWNhbmRpZGF0ZTowIDEgVURQIDIxMjIyNTI1NDMgMTkyLjE2OC4wLjExMiA1NDQ0MSB0eXAgaG9zdFxyXG5hPWNhbmRpZGF0ZToyIDEgVENQIDIxMDU1MjQ0NzkgMTkyLjE2OC4wLjExMiA5IHR5cCBob3N0IHRjcHR5cGUgYWN0aXZlXHJcbmE9Y2FuZGlkYXRlOjAgMiBVRFAgMjEyMjI1MjU0MiAxOTIuMTY4LjAuMTEyIDQ4MjU4IHR5cCBob3N0XHJcbmE9Y2FuZGlkYXRlOjIgMiBUQ1AgMjEwNTUyNDQ3OCAxOTIuMTY4LjAuMTEyIDkgdHlwIGhvc3QgdGNwdHlwZSBhY3RpdmVcclxuYT1jYW5kaWRhdGU6MSAxIFVEUCAxNjg2MDUyODYzIDk4LjIwMC4yNDMuMTUxIDU0NDQxIHR5cCBzcmZseCByYWRkciAxOTIuMTY4LjAuMTEyIHJwb3J0IDU0NDQxXHJcbmE9Y2FuZGlkYXRlOjEgMiBVRFAgMTY4NjA1Mjg2MiA5OC4yMDAuMjQzLjE1MSA0ODI1OCB0eXAgc3JmbHggcmFkZHIgMTkyLjE2OC4wLjExMiBycG9ydCA0ODI1OFxyXG5hPXNlbmRyZWN2XHJcbmE9ZW5kLW9mLWNhbmRpZGF0ZXNcclxuYT1leHRtYXA6MyB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDpzZGVzOm1pZFxyXG5hPWV4dG1hcDo0IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L2Ficy1zZW5kLXRpbWVcclxuYT1leHRtYXA6NSB1cm46aWV0ZjpwYXJhbXM6cnRwLWhkcmV4dDp0b2Zmc2V0XHJcbmE9ZXh0bWFwOjYvcmVjdm9ubHkgaHR0cDovL3d3dy53ZWJydGMub3JnL2V4cGVyaW1lbnRzL3J0cC1oZHJleHQvcGxheW91dC1kZWxheVxyXG5hPWZtdHA6MTIwIG1heC1mcz0xMjI4ODttYXgtZnI9NjBcclxuYT1mbXRwOjEyMSBtYXgtZnM9MTIyODg7bWF4LWZyPTYwXHJcbmE9aWNlLXB3ZDozMmE5NDcyZWVlMTllMmRiM2QwNmI0ODA5NWFmYTk1ZFxyXG5hPWljZS11ZnJhZzo3NGRjMDdiNVxyXG5hPW1pZDowXHJcbmE9bXNpZDotIHtkZTMwMmU5Yi1mYTE2LTQ2NzYtOGNjMy1hMDg2ZjljOWExYWJ9XHJcbmE9cnRjcDo0ODI1OCBJTiBJUDQgOTguMjAwLjI0My4xNTFcclxuYT1ydGNwLWZiOjEyMCBuYWNrXHJcbmE9cnRjcC1mYjoxMjAgbmFjayBwbGlcclxuYT1ydGNwLWZiOjEyMCBjY20gZmlyXHJcbmE9cnRjcC1mYjoxMjAgZ29vZy1yZW1iXHJcbmE9cnRjcC1mYjoxMjEgbmFja1xyXG5hPXJ0Y3AtZmI6MTIxIG5hY2sgcGxpXHJcbmE9cnRjcC1mYjoxMjEgY2NtIGZpclxyXG5hPXJ0Y3AtZmI6MTIxIGdvb2ctcmVtYlxyXG5hPXJ0Y3AtbXV4XHJcbmE9cnRwbWFwOjEyMCBWUDgvOTAwMDBcclxuYT1ydHBtYXA6MTIxIFZQOS85MDAwMFxyXG5hPXNldHVwOmFjdHBhc3NcclxuYT1zc3JjOjQ1NjIyNTg3NiBjbmFtZTp7YjdlNjVmZDYtYjA1OC00NDcwLTg5ZDItYjU4ODBlODUwMGE4fVxyXG4ifQ\"}"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := http.NewRequest("POST", "/browsersdp", strings.NewReader(validSessionDescription))
		if err != nil {
			t.Errorf("NewRequest error: %s", err)
		}

		err = getBrowserSdp(w, r, mockArgs)
		if err == nil {
			t.Errorf("no error in server response")
		}
	}))

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("%s", err)
	}
}
