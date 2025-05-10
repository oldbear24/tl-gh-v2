package main

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPool is a mock implementation of pgxpool.Pool.
type MockPool struct {
	mock.Mock
}

func (m *MockPool) Acquire(ctx context.Context) (*MockConn, error) {
	args := m.Called(ctx)
	return args.Get(0).(*MockConn), args.Error(1)
}

// MockConn is a mock implementation of pgxpool.Conn.
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgx.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.CommandTag), args.Error(1)
}

func (m *MockConn) Release() {
	m.Called()
}

func TestSetMemberInactive(t *testing.T) {
	mockPool := new(MockPool)
	mockConn := new(MockConn)

	// Mock the pool to return the mock connection.
	mockPool.On("Acquire", mock.Anything).Return(mockConn, nil)

	// Mock the Exec call to simulate a successful update.
	mockConn.On("Exec", mock.Anything, "UPDATE players SET active = false WHERE discord_id = $1 AND guild = $2", mock.Anything, mock.Anything).Return(nil, nil)

	// Mock the Release call.
	mockConn.On("Release").Return()

	// Replace the global pool with the mock pool.
	pool = mockPool

	// Create a test member and guild ID.
	member := &discordgo.Member{
		User: &discordgo.User{
			ID: "12345",
		},
	}
	guildId := "67890"

	// Call the function.
	SetMemberInactive(member, guildId)

	// Assert that the mock methods were called as expected.
	mockPool.AssertCalled(t, "Acquire", mock.Anything)
	mockConn.AssertCalled(t, "Exec", mock.Anything, "UPDATE players SET active = false WHERE discord_id = $1 AND guild = $2", member.User.ID, guildId)
	mockConn.AssertCalled(t, "Release")
}

func TestSetMemberInactive_AcquireError(t *testing.T) {
	mockPool := new(MockPool)

	// Mock the pool to return an error when acquiring a connection.
	mockPool.On("Acquire", mock.Anything).Return(nil, assert.AnError)

	// Replace the global pool with the mock pool.
	pool = mockPool

	// Create a test member and guild ID.
	member := &discordgo.Member{
		User: &discordgo.User{
			ID: "12345",
		},
	}
	guildId := "67890"

	// Call the function.
	SetMemberInactive(member, guildId)

	// Assert that Acquire was called and no further calls were made.
	mockPool.AssertCalled(t, "Acquire", mock.Anything)
}

func TestSetMemberInactive_ExecError(t *testing.T) {
	mockPool := new(MockPool)
	mockConn := new(MockConn)

	// Mock the pool to return the mock connection.
	mockPool.On("Acquire", mock.Anything).Return(mockConn, nil)

	// Mock the Exec call to return an error.
	mockConn.On("Exec", mock.Anything, "UPDATE players SET active = false WHERE discord_id = $1 AND guild = $2", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Mock the Release call.
	mockConn.On("Release").Return()

	// Replace the global pool with the mock pool.
	pool = mockPool

	// Create a test member and guild ID.
	member := &discordgo.Member{
		User: &discordgo.User{
			ID: "12345",
		},
	}
	guildId := "67890"

	// Call the function.
	SetMemberInactive(member, guildId)

	// Assert that the mock methods were called as expected.
	mockPool.AssertCalled(t, "Acquire", mock.Anything)
	mockConn.AssertCalled(t, "Exec", mock.Anything, "UPDATE players SET active = false WHERE discord_id = $1 AND guild = $2", member.User.ID, guildId)
	mockConn.AssertCalled(t, "Release")
}
