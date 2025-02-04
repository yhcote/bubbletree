// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package app

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

var (
	ErrViperToLocalConfig = errors.New("could not translate viper to local app settings")
)

const (
	// Config keys used by both 'huh forms' and viper to identify one specific
	// configuration item.
	EmployeeName       = "employee.name"
	EmployeeDOB        = "employee.dob"
	EmployeeRole       = "employee.role"
	EmployeeSalary     = "employee.salary"
	EmployeeHiringDate = "employee.hiringdate"
	EmployeeAcctPasswd = "employee.password"

	// Set some reasonable defaults when possible
	DefaultRole   = "Software Engineer"
	DefaultPasswd = "letmein"
)

// AllConfigKeys is the complete list of config keys. This is looped over by
// the configurator model to process config.
var AllConfigKeys = []string{
	EmployeeName,
	EmployeeDOB,
	EmployeeRole,
	EmployeeSalary,
	EmployeeHiringDate,
	EmployeeAcctPasswd,
}

// ViperToLocalConfig translates configurations loaded into Viper into a
// locally structured config instance.
func ViperToLocalConfig(viper *viper.Viper) (AppConfig, error) {
	name := viper.GetString(EmployeeName)
	if name == "" {
		return AppConfig{}, fmt.Errorf("`name` is empty, this should not happen: %w", ErrViperToLocalConfig)
	}
	dob := viper.GetString(EmployeeDOB)
	if dob == "" {
		return AppConfig{}, fmt.Errorf("`dob` is empty, this should not happen: %w", ErrViperToLocalConfig)
	}
	role := viper.GetString(EmployeeRole)
	if role == "" {
		return AppConfig{}, fmt.Errorf("`role` is empty, this should not happen: %w", ErrViperToLocalConfig)
	}
	salary := viper.GetString(EmployeeSalary)
	if salary == "" {
		return AppConfig{}, fmt.Errorf("`salary` is empty, this should not happen: %w", ErrViperToLocalConfig)
	}
	hiringDate := viper.GetString(EmployeeHiringDate)
	if hiringDate == "" {
		return AppConfig{}, fmt.Errorf("`hiringDate` is empty, this should not happen: %w", ErrViperToLocalConfig)
	}
	passwd := viper.GetString(EmployeeAcctPasswd)
	if passwd == "" {
		return AppConfig{}, fmt.Errorf("`passwd` is empty, this should not happen: %w", ErrViperToLocalConfig)
	}
	return AppConfig{
		Employees: []Employee{
			{
				Name:       name,
				Dob:        dob,
				Role:       role,
				Salary:     salary,
				HiringDate: hiringDate,
				AcctPasswd: passwd,
			},
		},
	}, nil
}

// AppConfig is the local App config
type AppConfig struct {
	Employees []Employee
}

// Employee config element
type Employee struct {
	Name       string
	Dob        string
	Role       string
	Salary     string
	HiringDate string
	AcctPasswd string
}
