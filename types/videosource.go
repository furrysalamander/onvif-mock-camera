package types

import "image"

type VideoSource interface {
	Start() (<-chan *image.RGBA, error)
	Stop() error
	FrameSize() (width, height int)

	ContinuousMove(pan, tilt, zoom float64) error
	RelativeMove(pan, tilt, zoom float64) error
	AbsoluteMove(pan, tilt, zoom float64) error
	StopPTZ(panTilt, zoom bool) error
}
