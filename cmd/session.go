package cmd

import (
	"path/filepath"

	"github.com/nigiwen/gen-handler/internal/parser"
	"github.com/nigiwen/gen-handler/internal/scanner"
	"github.com/nigiwen/gen-handler/internal/tui"
)

var (
	scanEntities  = scanner.ScanEntities
	runSession    = tui.RunSession
	globGrpcFiles = filepath.Glob
	parseGrpcFile = parser.ParseGrpcFile
)
