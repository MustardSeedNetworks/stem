//go:build !cgo || !linux

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package dataplane provides CGO bindings to the C reflector dataplane.
//
// This file contains stub implementations for non-CGO or non-Linux builds.
// The actual dataplane requires CGO and Linux for AF_PACKET/AF_XDP support.
package dataplane

import (
	"fmt"

	"github.com/krisarmstrong/stem/internal/reflector/config"
)

type Stats struct {
	PacketsReceived  uint64
	PacketsReflected uint64
	BytesReceived    uint64
	BytesReflected   uint64
	TxErrors         uint64
	RxInvalid        uint64
	SigProbeOT       uint64
	SigDataOT        uint64
	SigLatency       uint64
	SigRFC2544       uint64
	SigY1564         uint64
	SigMSN           uint64
	LatencyMin       float64
	LatencyAvg       float64
	LatencyMax       float64
	LatencyCount     uint64
}

type Dataplane struct {
	cfg     *config.Config
	running bool //nolint:unused // placeholder for CGO implementation
}

var ErrNotSupported = fmt.Errorf("CGO dataplane not available on this platform")

func New(cfg *config.Config) (*Dataplane, error) {
	return nil, ErrNotSupported
}

func (dp *Dataplane) Start() error {
	return ErrNotSupported
}

func (dp *Dataplane) Stop() {}

func (dp *Dataplane) Close() {}

func (dp *Dataplane) GetStats() Stats {
	return Stats{}
}

func (dp *Dataplane) IsRunning() bool {
	return false
}

func (dp *Dataplane) Interface() string {
	if dp != nil && dp.cfg != nil {
		return dp.cfg.Interface
	}
	return ""
}

func (dp *Dataplane) Config() *config.Config {
	if dp != nil {
		return dp.cfg
	}
	return nil
}

func (dp *Dataplane) UpdateConfig(updates map[string]interface{}) error {
	return ErrNotSupported
}

func (dp *Dataplane) ResetStats() {}
