// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"example/internal"
	"example/internal/app"
	"example/models/root"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// flags & viper config
	reconf      bool
	configFile  string
	viperConfig *viper.Viper

	// base logger
	log = app.DefaultLogger(!releaseBuild)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "This is a bubble tea application template",
	Long: `'app' is a base program template that can be use as a start to write large
or complex bubble tea applications that fits the use of a model tree to
modularize multiple components.

For more information and a complete usage description, see app(1)
manual page.`,
	Version: internal.ProgramVersion,

	PreRun: func(cmd *cobra.Command, args []string) {
		internal.ProgramName = cmd.Name()
	},

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log.Info("Starting", "program", internal.ProgramName, "version", internal.ProgramVersion)
		fmt.Printf("Starting %v version %v, log file: %v, config file: %v\n",
			internal.ProgramName, internal.ProgramVersion, app.GetLoggerOutputName(), viperConfig.ConfigFileUsed())

		// Initialize App's base (root) model
		m := root.New(
			root.WithLogger(log),
			root.WithViper(viperConfig),
			root.WithSpewConfigState(&spew.ConfigState{MaxDepth: 1}),
		)
		// Run the bubble tea program with the new base model.
		if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
			return err
		}

		return nil
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		_ = app.CloseLoggerOutput()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&reconf, "reconf", false, "force running through the configuration")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.config/app/app.toml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viperConfig = viper.New()

	// use command line flag set config file, or load from default location
	if configFile != "" {
		// When asked to reconfigure, truncate the current config file if it exists.
		if reconf {
			_ = os.Truncate(configFile, 0)
		}
		viperConfig.SetConfigFile(configFile)
	} else {
		viperConfig.SetConfigName("app")
		viperConfig.SetConfigType("toml")
		viperConfig.AddConfigPath("$HOME/.config/app")

		basedir := filepath.Join(os.Getenv("HOME"), ".config", "app")
		configFile = filepath.Join(basedir, "app.toml")
		if _, err := os.Stat(configFile); err != nil {
			err := os.MkdirAll(basedir, 0755)
			if err != nil {
				log.Error("Could not create config path", "path", basedir, "error", err)
				fmt.Printf("Could not create config path, path=%v, error=%v\n", basedir, err)
				os.Exit(2)
			}
			f, err := os.Create(configFile)
			if err != nil {
				log.Error("Could not create empty config file", "file", configFile, "error", err)
				fmt.Printf("Could not create empty config file, file=%v, error=%v\n", configFile, err)
				os.Exit(2)
			}
			f.Close()
		} else if reconf {
			_ = os.Truncate(configFile, 0)
		}
	}

	// read in environment variables that match
	viperConfig.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viperConfig.ReadInConfig(); err != nil {
		log.Error("Problem reading/decoding config file", "viper error", err)
		fmt.Printf("Problem reading/decoding config file, viper error=%v\n", err)
		os.Exit(2)
	} else {
		log.Info("Using config", "file", viperConfig.ConfigFileUsed())
		log.Debug("Config", "all settings", spew.Sdump(viperConfig.AllSettings()))
	}
}
