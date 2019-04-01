// Vautour - A distributed & extensible web hunter
// Copyright (C) 2019 Quentin Machu & Vautour contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"github.com/quentin-m/vautour/src/pkg/formatter"
	_ "github.com/quentin-m/vautour/src/modules/pastebin"
	"github.com/quentin-m/vautour/src/pkg/vautour"
	"github.com/quentin-m/vautour/src/pkg/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"
	_ "github.com/quentin-m/vautour/src/modules/redis"
	_ "github.com/quentin-m/vautour/src/modules/elasticsearch"
	_ "github.com/quentin-m/vautour/src/modules/yara"
)

func main() {
	// Parse command-line arguments.
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagConfigPath := flag.String("config", "./config/vautour.yaml", "Load configuration from the specified file.")
	flagLogLevel := flag.String("log-level", "info", "Define the logging level.")
	flagCPUProfilePath := flag.String("cpu-profile", "", "Write a CPU profile to the specified file before exiting.")
	flagVersion := flag.Bool("version", false, "Display the version of Vautour.")
	flag.Parse()

	if *flagVersion {
		fmt.Printf("Vautour (%s)\n", version.Version)
		return
	}

	// Configure Vautour.
	configureLogger(flagLogLevel)

	// Load configuration.
	config, err := loadConfig(*flagConfigPath)
	if err != nil {
		log.WithError(err).Fatal("failed to load configuration")
	}

	// Enable CPU Profiling if specified
	if *flagCPUProfilePath != "" {
		defer stopCPUProfiling(startCPUProfiling(*flagCPUProfilePath))
	}

	// Start Vautour.
	vautour.Boot(config)
}

// Initialize logging system
func configureLogger(flagLogLevel *string) {
	logLevel, err := log.ParseLevel(strings.ToUpper(*flagLogLevel))
	if err != nil {
		log.WithError(err).Error("failed to set logger parser level")
	}

	log.SetLevel(logLevel)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&formatter.JSONExtendedFormatter{ShowLn: true})
}

func startCPUProfiling(path string) *os.File {
	f, err := os.Create(path)
	if err != nil {
		log.WithError(err).Fatal("failed to create profile file")
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.WithError(err).Fatal("failed to start CPU profiling")
	}

	log.Info("started CPU profiling")

	return f
}

func stopCPUProfiling(f *os.File) {
	pprof.StopCPUProfile()
	f.Close()
	log.Info("stopped CPU profiling")
}

type namespacedConfig struct {
	Vautour vautour.Config
}

func loadConfig(cfgPath string) (vautour.Config, error) {
	cfg := vautour.Config{}

	// Load configuration file if specified.
	if cfgPath == "" {
		return cfg, nil
	}
	yamlConfig, err := ioutil.ReadFile(os.ExpandEnv(cfgPath))
	if err != nil {
		log.WithError(err).Fatal("could not load configuration file")
	}

	// Parse the configuration.
	var nCfg namespacedConfig
	if err := yaml.Unmarshal([]byte(yamlConfig), &nCfg); err != nil {
		return cfg, err
	}

	return nCfg.Vautour, nil
}
