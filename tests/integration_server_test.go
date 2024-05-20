//go:build integration
// +build integration

// Package tests. Integration tests.
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tccompose "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/client"
)

// integration tests for the server part.

type IntegrationServerTestSuite struct {
	suite.Suite
}

func Test_IntegrationServerTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationServerTestSuite))
}

func (s *IntegrationServerTestSuite) Test_Basic_Integration() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	t := s.T()

	// ARRANGE
	port := "7100"
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	compose, err := tccompose.NewDockerComposeWith(tccompose.WithStackFiles("../build/dev/docker-compose.yaml"))
	require.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tccompose.RemoveOrphans(true), tccompose.RemoveImagesLocal), "compose.Down()")
	})

	serverPort, err := nat.NewPort("tcp", port)
	require.NoError(t, err, "nat.NewPort()")

	// we are not running clients from docker-compose, we want to request the server locally
	err = compose.
		WaitForService("server", wait.ForListeningPort(serverPort)).
		Up(
			context.Background(),
			tccompose.RunServices("redis", "server"),
			tccompose.Wait(true),
		)
	require.NoError(t, err, "compose.Up()")
	time.Sleep(5 * time.Second) // giving time to start properly

	cl, err := client.New("127.0.0.1:" + port)
	require.NoError(t, err, "client.New()")

	// ACT+ASSERT 1: auth req
	respAuth, err := cl.SendRequest(
		context.Background(),
		protocol.Request{
			Type:    protocol.MessageTypeClientAuthReq,
			Payload: nil,
		},
	)
	require.NoError(t, err, "cl.SendRequest(auth)")
	require.Equal(t, protocol.MessageTypeSrvAuthResp, respAuth.Type, "sending auth request")
	require.Contains(t, respAuth.Payload[0], "v1;", "auth response payload")

	// ACT+ASSERT 2: invalid data request
	respData1, err := cl.SendRequest(
		context.Background(),
		protocol.Request{
			Type:    protocol.MessageTypeClientDataReq,
			Payload: protocol.GeneratePayloadFromTokenAndSeed("wrong token", respAuth.Payload[0]),
		},
	)
	require.NoError(t, err, "cl.SendRequest(auth)")
	require.Equal(t, protocol.MessageTypeError, respData1.Type, "sending invalid data request")

	// continue here with the client journey
}
