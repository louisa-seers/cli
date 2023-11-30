/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */

package cmd

import (
	"github.com/spf13/cobra"
)

type PropertiesCmdValues struct {
	ecosystemBootstrap string
	namespace          string
	propertyName       string
}

type PropertiesCommand struct {
	values       *PropertiesCmdValues
	cobraCommand *cobra.Command
}

// ------------------------------------------------------------------------------------------------
// Constructors methods
// ------------------------------------------------------------------------------------------------
func NewPropertiesCommand(factory Factory, rootCommand GalasaCommand) (GalasaCommand, error) {

	cmd := new(PropertiesCommand)
	err := cmd.init(factory, rootCommand)
	return cmd, err
}

// ------------------------------------------------------------------------------------------------
// Public methods
// ------------------------------------------------------------------------------------------------
func (cmd *PropertiesCommand) Name() string {
	return COMMAND_NAME_PROPERTIES
}

func (cmd *PropertiesCommand) CobraCommand() *cobra.Command {
	return cmd.cobraCommand
}

func (cmd *PropertiesCommand) Values() interface{} {
	return cmd.values
}

// ------------------------------------------------------------------------------------------------
// Private methods
// ------------------------------------------------------------------------------------------------
func (cmd *PropertiesCommand) init(factory Factory, rootCommand GalasaCommand) error {

	var err error = nil

	propertiesCmdValues := &PropertiesCmdValues{}

	propertiesCobraCmd := &cobra.Command{
		Use:   "properties",
		Short: "Manages properties in an ecosystem",
		Long:  "Allows interaction with the CPS to create, query and maintain properties in Galasa Ecosystem",
	}

	addBootstrapFlag(propertiesCobraCmd, &propertiesCmdValues.ecosystemBootstrap)

	rootCommand.CobraCommand().AddCommand(propertiesCobraCmd)

	cmd.values = propertiesCmdValues
	cmd.cobraCommand = propertiesCobraCmd

	return err
}

func addNamespaceProperty(cmd *cobra.Command, isMandatory bool, propertiesCmdValues *PropertiesCmdValues) {

	flagName := "namespace"
	var description string
	if isMandatory {
		description = "A mandatory flag that describes the container for a collection of properties."
	} else {
		description = "An optional flag that describes the container for a collection of properties."
	}

	cmd.PersistentFlags().StringVarP(&propertiesCmdValues.namespace, flagName, "s", "", description)

	if isMandatory {
		cmd.MarkPersistentFlagRequired(flagName)
	}

}

// Some sub-commands need a name field to be mandatory, some don't.
func addNameProperty(cmd *cobra.Command, isMandatory bool, propertiesCmdValues *PropertiesCmdValues) {
	flagName := "name"
	var description string
	if isMandatory {
		description = "A mandatory field indicating the name of a property in the namespace."
	} else {
		description = "An optional field indicating the name of a property in the namespace."
	}

	cmd.PersistentFlags().StringVarP(&propertiesCmdValues.propertyName, flagName, "n", "", description)

	if isMandatory {
		cmd.MarkPersistentFlagRequired(flagName)
	}
}
