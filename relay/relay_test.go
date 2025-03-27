package relay

import (
	"context"
	"testing"

	pb "github.com/boozec/rahanna/relay/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterName(t *testing.T) {
	server := &Server{}
	ctx := context.Background()

	t.Run("Valid IP Registration", func(t *testing.T) {
		response, err := server.RegisterName(ctx, &pb.RelayRequest{Ip: "192.168.1.1"})
		require.NoError(t, err)
		assert.NotEmpty(t, response.Name)
		assert.Equal(t, "192.168.1.1", response.Ip)
	})

	t.Run("Empty IP Registration", func(t *testing.T) {
		response, err := server.RegisterName(ctx, &pb.RelayRequest{Ip: ""})
		assert.Error(t, err)
		assert.Nil(t, response)
	})

}

func TestLookup(t *testing.T) {
	server := &Server{}
	ctx := context.Background()

	registerResponse, err := server.RegisterName(ctx, &pb.RelayRequest{Ip: "192.168.1.1"})
	require.NoError(t, err)

	t.Run("Successful Lookup", func(t *testing.T) {
		response, err := server.Lookup(ctx, &pb.LookupRequest{Name: registerResponse.Name})
		require.NoError(t, err)
		assert.Equal(t, registerResponse.Name, response.Name)
		assert.Equal(t, "192.168.1.1", response.Ip)
	})

	t.Run("Lookup with Empty Name", func(t *testing.T) {
		response, err := server.Lookup(ctx, &pb.LookupRequest{Name: ""})
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("Lookup Non-Existent Name", func(t *testing.T) {
		response, err := server.Lookup(ctx, &pb.LookupRequest{Name: "nonexistent"})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestCloseName(t *testing.T) {
	server := &Server{}
	ctx := context.Background()

	// Prepare a registered name
	registerResponse, err := server.RegisterName(ctx, &pb.RelayRequest{Ip: "192.168.1.1"})
	require.NoError(t, err)

	t.Run("Successful Close", func(t *testing.T) {
		response, err := server.CloseName(ctx, &pb.LookupRequest{Name: registerResponse.Name})
		require.NoError(t, err)
		assert.True(t, response.Status)

		// Verify the name is no longer in the table
		_, err = server.Lookup(ctx, &pb.LookupRequest{Name: registerResponse.Name})
		assert.Error(t, err)
	})

	t.Run("Close with Empty Name", func(t *testing.T) {
		_, err := server.CloseName(ctx, &pb.LookupRequest{Name: ""})
		assert.Error(t, err)
	})

	t.Run("Close Non-Existent Name", func(t *testing.T) {
		response, err := server.CloseName(ctx, &pb.LookupRequest{Name: "nonexistent"})
		assert.Error(t, err)
		assert.False(t, response.Status)
	})
}
