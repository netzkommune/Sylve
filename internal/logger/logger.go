// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var L zerolog.Logger

func InitLogger(dataDir string, level int8) {
	zerolog.TimeFieldFormat = "2006/01/02 15:04:05"

	switch level {
	case 0:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 4:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case 5:
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006/01/02 15:04:05",
		NoColor:    false,
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(dataDir, "logs.json"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     28,
		Compress:   true,
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

	L = zerolog.New(multiWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = L

	fmt.Println("")
	L.Info().Msg("Logger initialized")
}
