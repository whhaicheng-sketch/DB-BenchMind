package usecase

import (
	"database/sql"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
)

func TestOracleSwingbenchSchemaCountValueTreatsNullAsZero(t *testing.T) {
	if got := oracleSwingbenchSchemaCountValue(sql.NullInt64{}); got != 0 {
		t.Fatalf("oracleSwingbenchSchemaCountValue(NULL) = %d, want 0", got)
	}
	if got := oracleSwingbenchSchemaCountValue(sql.NullInt64{Int64: 4, Valid: true}); got != 4 {
		t.Fatalf("oracleSwingbenchSchemaCountValue(4) = %d, want 4", got)
	}
}

func TestOracleSwingbenchSchemaReadyRequiresFullRunPrerequisites(t *testing.T) {
	if oracleSwingbenchSchemaReady(4, 2, 3) != true {
		t.Fatal("oracleSwingbenchSchemaReady() should accept full object set")
	}
	if oracleSwingbenchSchemaReady(4, 1, 3) != false {
		t.Fatal("oracleSwingbenchSchemaReady() should reject missing ORDERENTRY package body")
	}
	if oracleSwingbenchSchemaReady(2, 2, 3) != false {
		t.Fatal("oracleSwingbenchSchemaReady() should reject partial SOE tables")
	}
	if oracleSwingbenchSchemaReady(4, 2, 0) != false {
		t.Fatal("oracleSwingbenchSchemaReady() should reject missing supporting sequences")
	}
}

func TestOracleSwingbenchZeroThroughputFailure_RequiresPrepareWhenRunNeverStarts(t *testing.T) {
	stdout := `Time     Users       TPM      TPS     Errors   NCR   UCD   BP    OP    PO    BO
10:58:35 [0/10]      0        0       0        0     0     0     0     0     0
10:58:36 [0/10]      0        0       0        0     0     0     0     0     0
10:58:37 [0/10]      0        0       0        0     0     0     0     0     0`

	err := oracleSwingbenchZeroThroughputFailure(stdout, &adapter.FinalResult{
		TotalTransactions:  0,
		TransactionsPerSec: 0,
		AvgTPS:             0,
		AvgTPM:             0,
	})
	if err == nil {
		t.Fatal("oracleSwingbenchZeroThroughputFailure() expected error")
	}
	if got := err.Error(); !containsBenchmarkUseCaseSubs(got, []string{"Run Prepare first", "Cleanup removes workload objects", "[0/"}) {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestOracleSwingbenchZeroThroughputFailure_AllowsRunsAfterRealThroughputAppears(t *testing.T) {
	stdout := `Time     Users       TPM      TPS     Errors   NCR   UCD   BP    OP    PO    BO
10:58:35 [10/10]     120      2       0        3     1     0     0     0     0`

	err := oracleSwingbenchZeroThroughputFailure(stdout, &adapter.FinalResult{
		TotalTransactions:  120,
		TransactionsPerSec: 2,
		AvgTPS:             2,
		AvgTPM:             120,
	})
	if err != nil {
		t.Fatalf("oracleSwingbenchZeroThroughputFailure() unexpected error: %v", err)
	}
}

func TestOracleSwingbenchRunPreflightError_CleanupInvalidatedRequiresPrepare(t *testing.T) {
	err := oracleSwingbenchRunPreflightError(oracleSwingbenchRunPreflightCleanupInvalidated, nil)
	if err == nil {
		t.Fatal("oracleSwingbenchRunPreflightError() expected error")
	}
	want := "Oracle Swingbench run failed: Cleanup removed required SOE objects. Please run Prepare first."
	if got := err.Error(); got != want {
		t.Fatalf("cleanup invalidated error = %q, want %q", got, want)
	}
}

func TestOracleSwingbenchEnvironmentHistoryGuard_FailsDirectRunAfterCleanup(t *testing.T) {
	runs := []*execution.Run{
		{
			ID:    "run-cleanup",
			State: execution.StateCompleted,
			Result: &execution.BenchmarkResult{
				ConnectionName: "Oracle Shared",
				DatabaseType:   "oracle",
				TemplateName:   "Oracle Swingbench Cleanup",
			},
			Message: "phase=cleanup",
		},
	}

	blocked, err := oracleSwingbenchHistoryInvalidatesDirectRun(
		runs,
		"Oracle Shared",
		"oracle",
		"swingbench",
		"order-entry",
	)
	if err != nil {
		t.Fatalf("oracleSwingbenchHistoryInvalidatesDirectRun() unexpected error: %v", err)
	}
	if !blocked {
		t.Fatal("oracleSwingbenchHistoryInvalidatesDirectRun() should block direct run after cleanup")
	}
}

func TestOracleSwingbenchEnvironmentHistoryGuard_ClearsAfterPrepare(t *testing.T) {
	runs := []*execution.Run{
		{
			ID:    "run-cleanup",
			State: execution.StateCompleted,
			Result: &execution.BenchmarkResult{
				ConnectionName: "Oracle Shared",
				DatabaseType:   "oracle",
				TemplateName:   "Oracle Swingbench Cleanup",
			},
			Message: "phase=cleanup",
		},
		{
			ID:    "run-prepare",
			State: execution.StatePrepared,
			Result: &execution.BenchmarkResult{
				ConnectionName: "Oracle Shared",
				DatabaseType:   "oracle",
				TemplateName:   "Oracle Swingbench Prepare",
			},
			Message: "phase=prepare",
		},
	}

	blocked, err := oracleSwingbenchHistoryInvalidatesDirectRun(
		runs,
		"Oracle Shared",
		"oracle",
		"swingbench",
		"order-entry",
	)
	if err != nil {
		t.Fatalf("oracleSwingbenchHistoryInvalidatesDirectRun() unexpected error: %v", err)
	}
	if blocked {
		t.Fatal("oracleSwingbenchHistoryInvalidatesDirectRun() should allow run after later prepare")
	}
}
