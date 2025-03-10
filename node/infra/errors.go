package infra

import "errors"

var ErrUnsupportedNetwork = errors.New("unsupported network")
var ErrInvalidAddress = errors.New("invalid address")
var ErrConnectionRefused = errors.New("connection refused")
var ErrDialUnsupported = errors.New("dial unsupported")
