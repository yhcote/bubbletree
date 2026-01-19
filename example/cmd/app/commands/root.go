// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"example/internal/app"
	"example/models/coreapp"
	"example/ui/themes"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
	"github.com/yhcote/bubbletree/logger"
)

const (
	progname = "app"
)

var (
	// program general config flags & viper related
	configForce bool
	configFile  string
	configViper *viper.Viper
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   progname,
	Short: "This is a bubble tea application template",
	Long: "'" + progname + "'" + ` is a base program template that can be use as a start to write large
or complex bubble tea applications that fits the use of a model tree to
modularize multiple components.

For more information and a complete usage description, see ` + progname + `(1)
manual page.`,
	Version: app.ProgramVersion,

	PreRun: func(cmd *cobra.Command, args []string) {
		app.ProgramName = cmd.Name()
	},

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		logger.Log().Info("Starting", "program", app.ProgramName, "version", app.ProgramVersion)
		fmt.Printf("Starting %v version %v\n - log file:\t\t%v\n - config file:\t\t%v\n\n",
			app.ProgramName, app.ProgramVersion, logger.GetLoggerOutputName(), configViper.ConfigFileUsed())

		// Initialize App's base (coreapp) model
		options := []bubbletree.AppOption{
			bubbletree.WithProgname(app.ProgramName),
			bubbletree.WithProgver(app.ProgramVersion),
			bubbletree.WithConfigViper(configViper),
			bubbletree.WithReconfigure(configForce),
			bubbletree.WithSpewConfigState(&spew.ConfigState{MaxDepth: 1}),
			bubbletree.WithTheme(themes.Default()),
		}

		// Run the bubble tree program via the initialized coreapp.
		err = coreapp.Run(options...)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}
		return nil
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		_ = logger.CloseLoggerOutput()
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
	logger.SetLoggerLevel(!releaseBuild)
	cobra.OnInitialize(initConfigViper)

	rootCmd.PersistentFlags().BoolVar(&configForce, "reconf", false, "force running through the configuration")
	rootCmd.PersistentFlags().StringVar(&configFile, "config",
		filepath.Join(filepath.Join(os.Getenv("HOME"), ".config", progname), progname+".json"),
		"config file (default is $HOME/.config/"+progname+"/"+progname+".json)")
}

// initConfigViper reads in config file and ENV variables if set.
func initConfigViper() {
	configViper = viper.New()
	configViper.SetConfigFile(configFile)

	// Open an existing config or create a new one.
	openCreateJsonFile(configFile, configForce)

	// Read in environment variables that match.
	configViper.AutomaticEnv()

	// If a config file is found, read it in.
	readConfigFile(configViper)
	logger.Log().Info("Using config file", "file", configViper.ConfigFileUsed())
	logger.Log().Debug("Config file", "data", spew.Sdump(configViper.AllSettings()))
}

func readConfigFile(vpr *viper.Viper) {
	if err := vpr.ReadInConfig(); err != nil {
		logger.Log().Error("Problem reading/decoding viper data file", "file", vpr.ConfigFileUsed(), "error", err)
		fmt.Fprintf(os.Stderr, "Problem reading/decoding viper data file, file=%v, error=%v\n", vpr.ConfigFileUsed(), err)
		os.Exit(2)
	}
}

func openCreateJsonFile(filename string, force bool) {
	if _, err := os.Stat(filename); err != nil || force {
		err := os.MkdirAll(filepath.Dir(filename), 0755)
		if err != nil {
			logger.Log().Error("Could not create file path", "path", filepath.Dir(filename), "error", err)
			fmt.Fprintf(os.Stderr, "Could not create file path, path=%v, error=%v\n", filepath.Dir(filename), err)
			os.Exit(2)
		}
		f, err := os.Create(filename)
		if err != nil {
			logger.Log().Error("Could not create empty file", "file", filename, "error", err)
			fmt.Fprintf(os.Stderr, "Could not create empty file, file=%v, error=%v\n", filename, err)
			os.Exit(2)
		}
		_, err = f.WriteString("{}")
		if err != nil {
			logger.Log().Error("Could not write empty JSON to new file", "file", filename, "error", err)
			fmt.Fprintf(os.Stderr, "Could not write empty JSON to new file, file=%v, error=%v\n", filename, err)
			os.Exit(2)
		}
		_ = f.Close()
	}
}
