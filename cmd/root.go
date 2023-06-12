/*
Copyright © 2023 Denis Halturin <dhalturin@hotmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/shared"
)

var (
	configFile          = ""
	configPath          = ""
	logFormat           = ""
	logFormatJsonPretty = false
	logLevel            = ""
	rootCmd             = &cobra.Command{
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Short:             "Utility for refreshing Slack configuration token",
		Use:               shared.PkgName,
		Version:           shared.Version,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", shared.PathConf(), "Set the configuration file path")
	rootCmd.PersistentFlags().StringVar(&configFile, "config-file", "config.yaml", "Set the configuration file name")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Set the log format: text, json")
	rootCmd.PersistentFlags().BoolVar(&logFormatJsonPretty, "log-pretty", false, "Json logs will be indented")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "Set the log level: debug, info, warn, error, fatal")
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetConfigFile(fmt.Sprintf("%s/%s", configPath, configFile))
	viper.SetDefault("storage", "fs")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("rotator")

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("err", err).Fatal("can't parse log level")
	}

	log.SetOutput(io.Discard)
	log.SetLevel(level)
	log.SetFormatter(
		&log.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		},
	)

	log.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})

	log.AddHook(&writer.Hook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})

	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{
			PrettyPrint: logFormatJsonPretty,
		})
	}

	if err := viper.ReadInConfig(); err == nil {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Error(err)
	}
}
