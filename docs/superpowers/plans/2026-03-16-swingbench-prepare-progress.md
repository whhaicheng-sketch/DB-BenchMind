# Swingbench Prepare Progress Note

Last updated: 2026-03-16

## Current symptom

- Oracle objects and data can already exist (`SOE.CUSTOMERS` populated), but the GUI remains in `Prepare`.
- Oracle session inspection shows no active benchmark workload sessions; only the operator `sqlplus` session remains active.
- This indicates the database-side work is done, while DB-BenchMind is still waiting on local process completion/state finalization.

## Confirmed findings

- `cleanup/bootstrap/verify` now use `SYS AS SYSDBA`.
- Cleanup now fails fast with a concrete Oracle reason if `SOE` still exists after retries.
- Swingbench prepare no longer uses inactivity timeouts to kill a step.

## Current hypothesis

- `Build Indexes` is generated as an `sbutil` shell command.
- The command is currently detected as a Swingbench realtime command because the shell string contains `swingbench`.
- That routes `sbutil` through `executeCommandSwingbench()`, which is designed for `charbench/oewizard` output and completion behavior.
- The likely fix is to classify `sbutil` as a normal synchronous command so completion is based on the shell process exiting, not Swingbench realtime parsing semantics.

## Applied fix

- `isSwingbenchCommandLine()` has been narrowed so `sbutil` shell commands are no longer treated as Swingbench realtime commands.
- As a result, `Build Indexes` now runs through the normal synchronous execution path and finishes when the local shell process actually exits.
- This preserves the rule that completion is determined by the tool process result, not by fixed silence windows.

## Next implementation steps

1. Re-run Oracle `Prepare` manually and confirm the final `Build Indexes` step transitions out of `Prepare`.
2. If it still hangs, inspect the local process tree and run logs for the exact final shell command and exit state.
3. If needed, add explicit step completion logging around `Build Indexes` command start/exit.

## Resume commands

Run these from `/opt/project/DB-BenchMind`:

```bash
go test ./internal/app/usecase -run 'TestIsSwingbenchCommandLine|TestExecuteCommandSequence_CreateFailureLogsDiagnosticOutput|TestExecuteCommandSync_IncludesStdoutForSqlplusStyleFailures|TestSwingbenchNoOutputTimeoutForStep' -count=1
go test ./internal/infra/adapter -run 'TestSwingbenchAdapter_BuildPrepareCommand|TestSwingbenchAdapter_BuildCleanupCommand_DropsUserAndTablespaces|TestSwingbenchAdapter_BuildCleanupCommand_IsIdempotentForMissingOrPartialObjects|TestResolveOracleAdminCredentials_MapsSystemConnectionToSysdbaSemantics|TestResolveOracleWizardCredentials' -count=1
```

Manual verification target: re-run Oracle `Prepare` and watch whether the last step exits instead of hanging in `Prepare`.
