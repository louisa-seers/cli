/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandListContainsProjectCommand(t *testing.T) {
	/// Given...
	factory := NewMockFactory()
	commands, _ := NewCommandCollection(factory)

	// When...
	projectCommand, err := commands.GetCommand(COMMAND_NAME_PROJECT)
	assert.Nil(t, err)

	// Then...
	assert.NotNil(t, projectCommand)
	assert.Equal(t, COMMAND_NAME_PROJECT, projectCommand.Name())
	assert.Nil(t, projectCommand.Values())
}