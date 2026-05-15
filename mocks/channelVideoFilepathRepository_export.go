package mocks

import "github.com/stretchr/testify/mock"

// ChannelVideoFilepathRepository exposes the generated mock for cross-package tests.
// It wraps the unexported generated mock type.
type ChannelVideoFilepathRepository = channelVideoFilepathRepository

// NewChannelVideoFilepathRepository creates a new mock instance bound to testing.T.
func NewChannelVideoFilepathRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ChannelVideoFilepathRepository {
	return newChannelVideoFilepathRepository(t)
}
