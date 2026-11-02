$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$rootPrefix = $repoRoot.TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
$catalog = Get-Content -LiteralPath (Join-Path $repoRoot 'docs/languages.json') -Raw -Encoding UTF8 | ConvertFrom-Json
$failures = [Collections.Generic.List[string]]::new()

function Get-PagePath([string]$canonical, [string]$language) {
    if ($language -eq 'en') { return $canonical }
    if ($canonical.StartsWith('docs/')) { return "docs/$language/" + $canonical.Substring(5) }
    if ($canonical -in @('CONTRIBUTING.md', 'SECURITY.md', 'CHANGELOG.md')) { return "docs/$language/project/$canonical" }
    return "docs/$language/$canonical"
}

function Get-RelativeLink([string]$source, [string]$target) {
    $sourceUri = [Uri]::new([IO.Path]::GetFullPath((Join-Path $repoRoot $source)))
    $targetUri = [Uri]::new([IO.Path]::GetFullPath((Join-Path $repoRoot $target)))
    return [Uri]::UnescapeDataString($sourceUri.MakeRelativeUri($targetUri).ToString())
}

if ($catalog.primaryLanguage -cne 'en' -or (@($catalog.languages.code) -join ',') -cne 'en,es,it') {
    throw 'Expected English primary documentation and en,es,it languages.'
}
if (@($catalog.pages).Count -eq 0 -or @($catalog.pages | Select-Object -Unique).Count -ne @($catalog.pages).Count) {
    throw 'The documentation catalog is empty or contains duplicate pages.'
}

$expected = [Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
foreach ($canonical in $catalog.pages) {
    if ($canonical -notmatch '^[A-Za-z0-9_./-]+\.md$' -or $canonical -match '(^|/)\.\.(/|$)|^/|^docs/(es|it)/') {
        throw "Invalid canonical page: $canonical"
    }
    foreach ($language in $catalog.languages) {
        $page = Get-PagePath $canonical $language.code
        if (!$expected.Add($page)) { throw "Duplicate output page: $page" }
        $fullPath = Join-Path $repoRoot $page
        if (!(Test-Path -LiteralPath $fullPath -PathType Leaf)) {
            $failures.Add("Missing translation: $page")
            continue
        }
        $body = [IO.File]::ReadAllText($fullPath)
        if ($body -notmatch '(?m)^# [^\r\n]+\r?\n') { $failures.Add("Missing page title: $page") }
        $links = foreach ($other in $catalog.languages) {
            if ($other.code -eq $language.code) { '**' + $other.name + '**' }
            else { '[' + $other.name + '](' + (Get-RelativeLink $page (Get-PagePath $canonical $other.code)) + ')' }
        }
        $bar = $links -join (' ' + [char]0x00B7 + ' ')
        if ($body -notmatch ('(?m)^# [^\r\n]+\r?\n\r?\n' + [Regex]::Escape($bar) + '\r?\n\r?\n')) {
            $failures.Add("Missing or incorrect same-page language navigation: $page")
        }

        # Ignore fenced examples when checking Markdown links. Reject unfinished fences.
        $prose = [Collections.Generic.List[string]]::new()
        $fence = ''
        $fenceLength = 0
        foreach ($line in ($body -split '\r?\n')) {
            if ($line -match '^ {0,3}(`{3,}|~{3,})(.*)$') {
                $marker = $Matches[1]
                if ($fence -eq '') { $fence = $marker.Substring(0,1); $fenceLength = $marker.Length }
                elseif ($marker.Substring(0,1) -eq $fence -and $marker.Length -ge $fenceLength -and $Matches[2].Trim() -eq '') { $fence = '' }
                continue
            }
            if ($fence -eq '') { $prose.Add($line) }
        }
        if ($fence -ne '') { $failures.Add("Unclosed code fence: $page") }
        foreach ($link in [Regex]::Matches(($prose -join "`n"), '\]\(([^\s)]+)\)')) {
            $href = $link.Groups[1].Value
            if ($href -match '^[a-zA-Z][a-zA-Z0-9+.-]*:|^#|^//') { continue }
            $relative = [Uri]::UnescapeDataString(($href -split '[?#]',2)[0])
            if ($relative -eq '') { continue }
            $resolved = [IO.Path]::GetFullPath((Join-Path ([IO.Path]::GetDirectoryName($fullPath)) $relative))
            if (!$resolved.StartsWith($rootPrefix, [StringComparison]::OrdinalIgnoreCase)) {
                $failures.Add("Link escapes the repository: $page -> $href")
            } elseif (!(Test-Path -LiteralPath $resolved)) {
                $failures.Add("Broken relative link: $page -> $href")
            }
        }
    }
}

# Include untracked documentation so new pages cannot bypass translation coverage.
$markdown = @(& git -C $repoRoot ls-files --cached --others --exclude-standard -- '*.md')
if ($LASTEXITCODE -ne 0) { throw 'Cannot enumerate repository documentation.' }
foreach ($page in $markdown) {
    if (!$expected.Contains($page)) { $failures.Add("Markdown page is absent from the language catalog: $page") }
}
if ($failures.Count -gt 0) {
    foreach ($failure in $failures) { Write-Output $failure }
    throw "$($failures.Count) documentation check(s) failed."
}
Write-Output "Documentation OK: $($catalog.pages.Count) pages x $($catalog.languages.Count) languages; navigation, coverage, fences and relative links checked."
