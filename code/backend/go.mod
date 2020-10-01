module atn/code/backend/asv

go 1.13

replace atn/code/backend/internal/signal => ./internal/signal

require (
	atn/code/backend/internal/signal v0.0.0-00010101000000-000000000000
	github.com/pion/webrtc/v2 v2.2.0
	github.com/spf13/pflag v1.0.5
)
