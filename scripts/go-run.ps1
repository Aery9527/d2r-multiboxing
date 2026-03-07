param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GoRunArgs)

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptRoot

$resolvedArgs = if ($GoRunArgs -and $GoRunArgs.Length -gt 0) { $GoRunArgs } else { @() }
$goArgs = @("run", ".\cmd\d2r-hyper-launcher") + $resolvedArgs

Push-Location $repoRoot
try {
    & go @goArgs
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
