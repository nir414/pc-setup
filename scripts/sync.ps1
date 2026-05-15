#requires -Version 5.1
<#
PC 설정 동기화 스크립트.
sync.psd1 에 정의된 환경변수 루트별 폴더를 저장소의 SyncData/ 와 robocopy 로 동기화한다.
백업본은 git 으로 버전 관리한다.

  .\sync.ps1 -Mode Backup            # PC -> SyncData
  .\sync.ps1 -Mode Restore           # SyncData -> PC (확인 prompt)
  .\sync.ps1 -Mode Status            # 변경 사항 미리보기
  .\sync.ps1 -Mode Backup -Folder Notepad++   # 단일 폴더
  .\sync.ps1 -Mode Restore -Force             # prompt 생략
  .\sync.ps1 -Mode Backup -BackupRoot D:\Backup
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [ValidateSet('Backup','Restore','Status')]
    [string]$Mode,

    [string]$BackupRoot,
    [string]$Folder,
    [switch]$Force
)

$ErrorActionPreference = 'Stop'

function Resolve-Base {
    param([string]$Base)
    switch ($Base) {
        'APPDATA'      { return $env:APPDATA }
        'LOCALAPPDATA' { return $env:LOCALAPPDATA }
        'USERPROFILE'  { return $env:USERPROFILE }
        'ProgramData'  { return $env:ProgramData }
        default        { throw "알 수 없는 Base: $Base" }
    }
}

function Split-Excludes {
    # robocopy 의 /XD(폴더), /XF(파일) 로 패턴을 분리.
    # 확장자 와일드카드(*.xxx) 또는 명시적 파일명은 /XF, 나머지는 /XD.
    param([string[]]$Patterns)
    $dirs  = @()
    $files = @()
    foreach ($p in $Patterns) {
        if ($p -match '^\*\.[^\\\/]+$' -or $p -match '\.[A-Za-z0-9]+$') {
            $files += $p
        } else {
            $dirs += $p
        }
    }
    return @{ Dirs = $dirs; Files = $files }
}

function Invoke-FolderSync {
    param(
        [string]$Base,
        [string]$RelPath,
        [string[]]$Excludes,
        [string]$BackupRoot,
        [string]$Mode
    )

    $baseDir   = Resolve-Base -Base $Base
    $pcPath    = Join-Path $baseDir $RelPath
    $repoPath  = Join-Path (Join-Path $BackupRoot $Base) $RelPath

    if ($Mode -eq 'Backup' -or $Mode -eq 'Status') {
        $src = $pcPath;   $dst = $repoPath
    } else {
        $src = $repoPath; $dst = $pcPath
    }

    if (-not (Test-Path $src)) {
        Write-Host ("  [SKIP] {0}: 원본 없음 ({1})" -f $RelPath, $src) -ForegroundColor DarkYellow
        return
    }

    $split = Split-Excludes -Patterns $Excludes
    $rcArgs = @($src, $dst, '/E', '/R:1', '/W:1', '/NFL', '/NDL', '/NP')
    if ($Mode -eq 'Status') { $rcArgs += '/L' }
    if ($split.Dirs.Count  -gt 0) { $rcArgs += '/XD'; $rcArgs += $split.Dirs }
    if ($split.Files.Count -gt 0) { $rcArgs += '/XF'; $rcArgs += $split.Files }

    Write-Host ("`n=== [{0}] {1} ===" -f $Base, $RelPath) -ForegroundColor Cyan
    Write-Host ("    {0} -> {1}" -f $src, $dst) -ForegroundColor DarkGray
    & robocopy @rcArgs | Out-Host
    $code = $LASTEXITCODE
    # robocopy: Status(/L) 모드는 0~16이 정상 (변경사항 있음은 1-16), Restore/Backup 모드는 0-7만 정상.
    if ($Mode -ne 'Status' -and $code -ge 8) {
        Write-Warning ("{0}\{1}: robocopy 실패 (exit {2})" -f $Base, $RelPath, $code)
    }
}

# --- main ---
$scriptDir  = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot   = Split-Path -Parent $scriptDir
$configPath = Join-Path $scriptDir 'sync.psd1'
if (-not (Test-Path $configPath)) { throw "sync.psd1 을 찾을 수 없습니다: $configPath" }

$config = Import-PowerShellDataFile -Path $configPath
if (-not $BackupRoot) { $BackupRoot = Join-Path $repoRoot 'SyncData' }

# Sections -> 평면화된 (Base, RelPath, Excludes) 작업 목록.
# Folders 의 각 entry 는 @{ Path='...'; Excludes=@(...) } 형태.
# 최종 Excludes = GlobalExcludes + 폴더별 Excludes.
$globalExcludes = if ($config.GlobalExcludes) { @($config.GlobalExcludes) } else { @() }

$jobs = foreach ($base in $config.Sections.Keys) {
    $section = $config.Sections[$base]
    foreach ($entry in $section.Folders) {
        $folderExcludes = if ($entry.Excludes) { @($entry.Excludes) } else { @() }
        $mergedExcludes = @($globalExcludes) + @($folderExcludes) | Select-Object -Unique
        [pscustomobject]@{
            Base     = $base
            RelPath  = $entry.Path
            Excludes = $mergedExcludes
        }
    }
}

if ($Folder) {
    $jobs = @($jobs | Where-Object { $_.RelPath -eq $Folder -or (Split-Path $_.RelPath -Leaf) -eq $Folder })
    if ($jobs.Count -eq 0) { throw "sync.psd1 에 '$Folder' 항목이 없습니다." }
}

Write-Host ("Mode      : {0}" -f $Mode)
Write-Host ("BackupRoot: {0}" -f $BackupRoot)
Write-Host ("Targets   : {0}" -f (($jobs | ForEach-Object { "{0}\{1}" -f $_.Base, $_.RelPath }) -join ', '))

if ($Mode -eq 'Restore' -and -not $Force) {
    $reply = Read-Host "Restore 는 PC 의 설정을 SyncData 백업으로 덮어씁니다. 진행할까요? (y/N)"
    if ($reply -notmatch '^[Yy]') { Write-Host '취소됨.'; return }
}

if ($Mode -eq 'Backup' -and -not (Test-Path $BackupRoot)) {
    New-Item -ItemType Directory -Path $BackupRoot -Force | Out-Null
}

foreach ($job in $jobs) {
    Invoke-FolderSync -Base $job.Base -RelPath $job.RelPath -Excludes $job.Excludes `
                      -BackupRoot $BackupRoot -Mode $Mode
}

Write-Host "`n완료." -ForegroundColor Green
