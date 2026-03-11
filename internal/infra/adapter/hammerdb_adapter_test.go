// Package adapter provides HammerDB benchmark tool adapter tests.
package adapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// TestHammerDBAdapter_ParseFinalResults tests the ParseFinalResults function.
func TestHammerDBAdapter_ParseFinalResults(t *testing.T) {
	ctx := context.Background()
	adapter := NewHammerDBAdapter()

	t.Run("parse NOPM result", func(t *testing.T) {
		// HammerDB output with NOPM result
		stdout := `Virtual Users 1
TEST RESULT : System achieved 12345 NOPM from 1 Virtual Users
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 12345.0, result.TPM, "TPM should match NOPM value")
		assert.Equal(t, 12345.0/60, result.TransactionsPerSec, "TPS should be TPM/60")
		assert.Equal(t, 12345.0/60, result.AvgTPS, "AvgTPS should match TPS")
	})

	t.Run("parse TPM result", func(t *testing.T) {
		// HammerDB output with TPM result
		stdout := `Virtual Users 10
TEST RESULT : System achieved 50000 TPM from 10 Virtual Users
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 50000.0, result.TPM, "TPM should match")
		assert.Equal(t, 50000.0/60, result.TransactionsPerSec, "TPS should be TPM/60")
	})

	t.Run("parse with response times", func(t *testing.T) {
		// HammerDB output with response time statistics
		stdout := `Virtual Users 5
TEST RESULT : System achieved 25000 TPM from 5 Virtual Users
Average response time: 150.50ms
Min response time: 50.25ms
Max response time: 450.75ms
95th percentile response: 300.00ms
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 25000.0, result.TPM, "TPM should match")
		assert.Equal(t, 150.50, result.LatencyAvg, "Average latency should match")
		assert.Equal(t, 50.25, result.LatencyMin, "Min latency should match")
		assert.Equal(t, 450.75, result.LatencyMax, "Max latency should match")
		assert.Equal(t, 300.00, result.LatencyP95, "95th percentile should match")
	})

	t.Run("parse with errors", func(t *testing.T) {
		// HammerDB output with errors
		stdout := `Virtual Users 1
TEST RESULT : System achieved 10000 TPM from 1 Virtual Users
Errors: 5
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 10000.0, result.TPM, "TPM should match")
		assert.Equal(t, int64(5), result.IgnoredErrors, "Error count should match")
	})

	t.Run("parse with total time", func(t *testing.T) {
		// HammerDB output with total time
		stdout := `Virtual Users 10
TEST RESULT : System achieved 30000 TPM from 10 Virtual Users
Total time: 120.0s
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 30000.0, result.TPM, "TPM should match")
		assert.Equal(t, 120.0, result.TotalTime, "Total time should match")
	})

	t.Run("parse empty output", func(t *testing.T) {
		// Empty HammerDB output
		stdout := ``

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error for empty output")
		assert.Equal(t, 0.0, result.TPM, "TPM should be 0 for empty output")
		assert.Equal(t, 0.0, result.TransactionsPerSec, "TPS should be 0 for empty output")
	})

	t.Run("calculate transaction count from TPM and time", func(t *testing.T) {
		// HammerDB output with TPM and time (should calculate transaction count)
		stdout := `Virtual Users 5
TEST RESULT : System achieved 60000 TPM from 5 Virtual Users
Total time: 60.0s
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 60000.0, result.TPM, "TPM should match")
		assert.Equal(t, 60.0, result.TotalTime, "Total time should match")
		// Expected: 60000 * (60/60) = 60000 transactions
		assert.Equal(t, int64(60000), result.TotalTransactions, "Transaction count should be calculated")
	})

	t.Run("parse decimal TPM values", func(t *testing.T) {
		// HammerDB output with decimal TPM
		stdout := `Virtual Users 3
TEST RESULT : System achieved 12345.67 TPM from 3 Virtual Users
Logging of metrics stopped at...`

		result, err := adapter.ParseFinalResults(ctx, stdout)

		require.NoError(t, err, "ParseFinalResults should not return error")
		assert.Equal(t, 12345.67, result.TPM, "TPM should match decimal value")
	})
}

// TestHammerDBAdapter_Type tests the Type method.
func TestHammerDBAdapter_Type(t *testing.T) {
	adapter := NewHammerDBAdapter()
	assert.Equal(t, AdapterTypeHammerDB, adapter.Type(), "Adapter type should be HammerDB")
}

// TestHammerDBAdapter_SupportsDatabase tests the SupportsDatabase method.
func TestHammerDBAdapter_SupportsDatabase(t *testing.T) {
	adapter := NewHammerDBAdapter()

	t.Run("MySQL", func(t *testing.T) {
		assert.True(t, adapter.SupportsDatabase(connection.DatabaseTypeMySQL), "Should support MySQL")
	})

	t.Run("Oracle", func(t *testing.T) {
		assert.True(t, adapter.SupportsDatabase(connection.DatabaseTypeOracle), "Should support Oracle")
	})

	t.Run("SQLServer", func(t *testing.T) {
		assert.True(t, adapter.SupportsDatabase(connection.DatabaseTypeSQLServer), "Should support SQLServer")
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		assert.True(t, adapter.SupportsDatabase(connection.DatabaseTypePostgreSQL), "Should support PostgreSQL")
	})
}

// TestHammerDBAdapter_ValidateConfig tests the ValidateConfig method.
func TestHammerDBAdapter_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	adapter := NewHammerDBAdapter()

	t.Run("nil config", func(t *testing.T) {
		err := adapter.ValidateConfig(ctx, nil)
		assert.Error(t, err, "Should return error for nil config")
	})

	t.Run("config with nil connection", func(t *testing.T) {
		config := &Config{
			Connection: nil,
		}
		err := adapter.ValidateConfig(ctx, config)
		assert.Error(t, err, "Should return error for nil connection")
	})
}
