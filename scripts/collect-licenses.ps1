param(
    [Parameter(Mandatory = $true)][string]$OutputDir,
    [string]$GoExe = 'go',
    [string]$CargoExe = 'cargo',
    [string]$RustcExe = 'rustc'
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$licenseDir = Join-Path $OutputDir 'licenses'
New-Item -ItemType Directory -Path $licenseDir -Force | Out-Null
$notices = [System.Collections.Generic.List[string]]::new()
$notices.Add('# Third party notices')
$notices.Add('')
$notices.Add('Karatty is Apache-2.0. Dependencies retain their own licenses. The licenses/ directory preserves their supplied license and attribution texts. Rust entries include build and target-specific dependencies from Cargo.lock. For alternatives in SPDX expressions, this distribution uses MIT or Apache-2.0 where offered; additional required notices remain included.')
$notices.Add('')

function Copy-LicenseFiles([string]$Id, [string]$Source, [string[]]$Names = @()) {
    $files = if ($Names.Count) {
        @($Names | ForEach-Object { Get-Item -LiteralPath (Join-Path $Source $_) })
    } else {
        @(Get-ChildItem -LiteralPath $Source -File | Where-Object Name -Match '^(LICENSE|LICENCE|NOTICE|COPYING|COPYRIGHT)')
    }
    if (!$files.Count) { throw "Missing license text for $Id" }
    $safeId = $Id -replace '[^a-zA-Z0-9._-]', '_'
    $destination = Join-Path $licenseDir $safeId
    New-Item -ItemType Directory -Path $destination -Force | Out-Null
    foreach ($file in $files) { Copy-Item -LiteralPath $file.FullName -Destination (Join-Path $destination $file.Name) }
    return $safeId
}

Push-Location $repoRoot
try {
    Copy-Item -LiteralPath (Join-Path $repoRoot 'LICENSE') -Destination (Join-Path $OutputDir 'LICENSE')
    $goRoot = (& $GoExe env GOROOT).Trim()
    if ($LASTEXITCODE -ne 0) { throw 'Cannot locate Go license' }
    $id = Copy-LicenseFiles 'go-standard-library' $goRoot @('LICENSE','PATENTS')
    $notices.Add("- Go standard library: licenses/$id/")
    $modules = @(& $GoExe list -deps -f '{{if .Module}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}' ./cmd/termlink)
    if ($LASTEXITCODE -ne 0) { throw 'Cannot list Go client dependencies' }
    foreach ($module in ($modules | Where-Object { $_ -and $_ -notmatch '^github.com/karatekah/termlink\|' } | Sort-Object -Unique)) {
        $parts = $module -split '\|'
        $id = Copy-LicenseFiles "go-$($parts[0])-$($parts[1])" $parts[2]
        $notices.Add("- $($parts[0]) $($parts[1]): licenses/$id/")
    }
    $metadataJson = & $CargoExe metadata --locked --format-version 1 --manifest-path crypto/Cargo.toml
    if ($LASTEXITCODE -ne 0) { throw 'Cannot list Rust dependencies' }
    $metadata = $metadataJson | ConvertFrom-Json
    foreach ($package in ($metadata.packages | Where-Object source | Sort-Object name,version)) {
        $source = Split-Path -Parent $package.manifest_path
        $names = @()
        if ($package.name -in @('matrix-pickle','matrix-pickle-derive')) {
            $vcs = Get-Content -LiteralPath (Join-Path $source '.cargo_vcs_info.json') -Raw | ConvertFrom-Json
            if ($package.version -ne '0.2.3' -or $vcs.git.sha1 -ne '3eb007eae1ca87e3a3683c99a58fd19289abae52') { throw 'Review updated matrix-pickle license' }
            $source = Join-Path $repoRoot 'third_party/matrix-pickle-0.2.3'
        } elseif ($package.name -eq 'r-efi') {
            # r-efi keeps its MIT permission and copyright notices in AUTHORS.
            $names = @('AUTHORS')
        }
        $id = Copy-LicenseFiles "rust-$($package.name)-$($package.version)" $source $names
        $notices.Add("- $($package.name) $($package.version) [$($package.license)]: licenses/$id/")
    }
    $rustRoot = (& $RustcExe --print sysroot).Trim()
    if ($LASTEXITCODE -ne 0) { throw 'Cannot locate Rust standard library notices' }
    $rustDocs = Join-Path $rustRoot 'share/doc/rust'
    $rustDestination = Join-Path $licenseDir 'rust-standard-library'
    New-Item -ItemType Directory -Path $rustDestination -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $rustDocs 'COPYRIGHT-library.html') -Destination $rustDestination
    Copy-Item -LiteralPath (Join-Path $rustDocs 'licenses') -Destination $rustDestination -Recurse
    $notices.Add('- Rust standard library: licenses/rust-standard-library/COPYRIGHT-library.html and its license texts.')
    $locks = Join-Path $OutputDir 'dependency-locks'
    New-Item -ItemType Directory -Path $locks -Force | Out-Null
    Copy-Item -LiteralPath 'go.mod','go.sum','crypto/Cargo.lock' -Destination $locks
} finally {
    Pop-Location
}
[System.IO.File]::WriteAllText((Join-Path $OutputDir 'THIRD_PARTY_NOTICES.md'), ($notices -join "`n") + "`n", [System.Text.UTF8Encoding]::new($false))
