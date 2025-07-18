//go:build !nocert
// +build !nocert

package main

import "embed"

//go:embed cert/*
var certFS embed.FS