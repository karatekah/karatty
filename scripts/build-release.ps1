param(
    [ValidatePattern('^v[0-9]+\.[0-9]+\.[0-9]+-alpha\.[0-9]+$')]
    [string]$Version = 'v0.1.0-alpha.1',
    [string]$GoExe = 'go',
    [string]$CargoExe = 'cargo',
    [string]$RustcExe = 'rustc'
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$goVersion = (& $GoExe version).Trim()
$rustVersion = (& $RustcExe --version).Trim()
if ($goVersion -notmatch '^go version go1\.27\.3 ' -or $rustVersion -notmatch '^rustc 1\.99\.1 ') {
    throw 'Release builds require Go 1.27.3 and Rust 1.99.1'
}
$outputDir = Join-Path $repoRoot "dist/$Version"
if (Test-Path -LiteralPath $outputDir) {
    throw "Release output already exists: $outputDir"
}
$targetOS = (& $GoExe env GOOS).Trim()
$targetArch = (& $GoExe env GOARCH).Trim()
if ($LASTEXITCODE -ne 0 -or !$targetOS -or !$targetArch) { throw 'Cannot determine native Go target' }
$extension = if ($targetOS -eq 'windows') { '.exe' } else { '' }
$bundleName = "karatty-$Version-$targetOS-$targetArch"
$bundleDir = Join-Path $outputDir $bundleName
New-Item -ItemType Directory -Path $bundleDir -Force | Out-Null

$previousCGO = $env:CGO_ENABLED
try {
    $env:CGO_ENABLED = '0'
    Push-Location $repoRoot
    try {
        $clientPath = Join-Path $bundleDir "karatty$extension"
        & $GoExe build -buildvcs=false -trimpath -ldflags "-s -w -X main.version=$Version" -o $clientPath ./cmd/termlink
        if ($LASTEXITCODE -ne 0) { throw 'Go client build failed' }
        & $CargoExe build --locked --release --manifest-path crypto/Cargo.toml
        if ($LASTEXITCODE -ne 0) { throw 'Rust crypto build failed' }
        $cryptoSource = Join-Path $repoRoot "crypto/target/release/karatty-crypto$extension"
        if (!(Test-Path -LiteralPath $cryptoSource)) { throw 'Rust crypto executable missing' }
        Copy-Item -LiteralPath $cryptoSource -Destination (Join-Path $bundleDir "karatty-crypto$extension")
        & (Join-Path $PSScriptRoot 'collect-licenses.ps1') -OutputDir $bundleDir -GoExe $GoExe -CargoExe $CargoExe -RustcExe $RustcExe
        $commit = (& git rev-parse HEAD).Trim()
        if ($LASTEXITCODE -ne 0) { throw 'Cannot identify source commit' }
        [System.IO.File]::WriteAllText((Join-Path $bundleDir 'BUILD_INFO'), "version=$Version`ncommit=$commit`nplatform=$targetOS-$targetArch`n", [System.Text.UTF8Encoding]::new($false))
    } finally {
        Pop-Location
    }
} finally {
    if ($null -eq $previousCGO) { Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue } else { $env:CGO_ENABLED = $previousCGO }
}

$hashLines = @(
    foreach ($name in @("karatty$extension", "karatty-crypto$extension")) {
        $hash = (Get-FileHash -LiteralPath (Join-Path $bundleDir $name) -Algorithm SHA256).Hash.ToLowerInvariant()
        "$hash  $name"
    }
)
[System.IO.File]::WriteAllText((Join-Path $bundleDir 'SHA256SUMS'), ($hashLines -join "`n") + "`n", [System.Text.UTF8Encoding]::new($false))
if ($targetOS -eq 'windows') {
    $archive = Join-Path $outputDir "$bundleName.zip"
    Compress-Archive -LiteralPath $bundleDir -DestinationPath $archive
} else {
    $archive = Join-Path $outputDir "$bundleName.tar.gz"
    & tar -czf $archive -C $outputDir $bundleName
    if ($LASTEXITCODE -ne 0) { throw 'Archive creation failed' }
}
$archiveHash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
[System.IO.File]::WriteAllText((Join-Path $outputDir 'SHA256SUMS'), "$archiveHash  $(Split-Path -Leaf $archive)`n", [System.Text.UTF8Encoding]::new($false))
Write-Output "Built $bundleName and its archive in $outputDir"
