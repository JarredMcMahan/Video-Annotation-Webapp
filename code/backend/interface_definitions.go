package main

import (
	"io"

	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/ivfreader"
)

type ivfReader interface {
	ParseNextFrame() ([]byte, *ivfreader.IVFFrameHeader, error)
	ResetReader(reset func(bytesRead int64) io.Reader)
}

type videoMediaTrack interface {
	WriteSample(s media.Sample) error
}
