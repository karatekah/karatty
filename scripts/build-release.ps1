param(
    [ValidatePattern('^v[0-9]+\.[0-9]+\.[0-9]+-alpha\.[0-9]+$')]
    [string]$Version = 'v0.1.0-alpha.1',
    [string]$GoExe = 'go'
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$outputDir = Join-Path $repoRoot "dist/$Version"
if (Test-Path -LiteralPath $outputDir) {
    throw "Release output already exists: $outputDir"
}
New-Item -ItemType Directory -Path $outputDir -Force | Out-Null

$previousGOOS = $env:GOOS
$previousGOARCH = $env:GOARCH
$previousCGO = $env:CGO_ENABLED
$targets = @(
    @{ OS = 'windows'; Arch = 'amd64'; Ext = '.exe' },
    @{ OS = 'linux'; Arch = 'amd64'; Ext = '' },
    @{ OS = 'linux'; Arch = 'arm64'; Ext = '' },
    @{ OS = 'darwin'; Arch = 'amd64'; Ext = '' },
    @{ OS = 'darwin'; Arch = 'arm64'; Ext = '' }
)
$hashLines = [System.Collections.Generic.List[string]]::new()
try {
    $env:CGO_ENABLED = '0'
    Push-Location $repoRoot
    try {
        foreach ($target in $targets) {
            $env:GOOS = $target.OS
            $env:GOARCH = $target.Arch
            $name = "karatty-$Version-$($target.OS)-$($target.Arch)$($target.Ext)"
            $path = Join-Path $outputDir $name
            & $GoExe build -buildvcs=false -trimpath -ldflags "-s -w -X main.version=$Version" -o $path ./cmd/termlink
            if ($LASTEXITCODE -ne 0) { throw "Go build failed for $name" }
            $hash = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant()
            $hashLines.Add("$hash  $name")
        }
    } finally {
        Pop-Location
    }
} finally {
    if ($null -eq $previousGOOS) { Remove-Item Env:GOOS -ErrorAction SilentlyContinue } else { $env:GOOS = $previousGOOS }
    if ($null -eq $previousGOARCH) { Remove-Item Env:GOARCH -ErrorAction SilentlyContinue } else { $env:GOARCH = $previousGOARCH }
    if ($null -eq $previousCGO) { Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue } else { $env:CGO_ENABLED = $previousCGO }
}

$hashPath = Join-Path $outputDir 'SHA256SUMS'
[System.IO.File]::WriteAllText($hashPath, ($hashLines -join "`n") + "`n", [System.Text.UTF8Encoding]::new($false))
Write-Output "Built $($targets.Count) Karatty clients in $outputDir"
