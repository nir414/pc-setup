#requires -Version 5.1
<#
PC 설정 동기화 스크립트.
sync.psd1 에 정의된 앱 설정 폴더를 저장소의 SyncData/ 폴더와 robocopy 로 동기화한다.
백업본은 git 으로 버전 관리한다.

  .\sync.ps1 -Mode Backup            # PC -> SyncData
  .\sync.ps1 -Mode Restore           # SyncData -> PC (확인 prompt)
  .\sync.ps1 -Mode Status            # 변경 사항 미리보기
  .\sync.ps1 -Mode Backup -App Notepad++   # 단일 앱
  .\sync.ps1 -Mode Restore -Force          # prompt 생략
  .\sync.ps1 -Mode Backup -BackupRoot D:\Backup   # 위치 override
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [ValidateSet('Backup','Restore','Status')]
    [string]$Mode,

    [string]$BackupRoot,
    [string]$App,
    [switch]$Force
)

$ErrorActionPreference = 'Stop'

function Resolve-Base {
    param([string]$Base)
    switch ($Base) {
        'APPDATA'      { return $env:APPDATA }
        'LOCALAPPDATA' { return $env:LOCALAPPDATA }
        'USERPROFILE'  { return $env:USERPROFILE }
        default        { throw "알 수 없는 Base: $Base" }
    }
}

function Invoke-AppSync {
    param(
        [hashtable]$AppEntry,
        [string]$BackupRoot,
        [string]$Mode
    )

    $baseDir   = Resolve-Base -Base $AppEntry.Base
    $pcPath    = Join-Path $baseDir $AppEntry.Path
    $repoPath  = Join-Path (Join-Path $BackupRoot $AppEntry.Base) $AppEntry.Path

    if ($Mode -eq 'Backup' -or $Mode -eq 'Status') {
        $src = $pcPath;   $dst = $repoPath
    } else {
        $src = $repoPath; $dst = $pcPath
    }

    if (-not (Test-Path $src)) {
        Write-Host ("  [SKIP] {0}: 원본 없음 ({1})" -f $AppEntry.Name, $src) -ForegroundColor DarkYellow
        return
    }

    $rcArgs = @($src, $dst, '/E', '/R:1', '/W:1', '/NFL', '/NDL', '/NP')
    if ($Mode -eq 'Status') { $rcArgs += '/L' }
    if ($AppEntry.ExcludeDirs.Count  -gt 0) { $rcArgs += '/XD'; $rcArgs += $AppEntry.ExcludeDirs }
    if ($AppEntry.ExcludeFiles.Count -gt 0) { $rcArgs += '/XF'; $rcArgs += $AppEntry.ExcludeFiles }

    Write-Host ("`n=== {0} ({1} -> {2}) ===" -f $AppEntry.Name, $src, $dst) -ForegroundColor Cyan
    & robocopy @rcArgs | Out-Host
    $code = $LASTEXITCODE
    # robocopy: 0..7 정상, 8+ 실패
    if ($code -ge 8) {
        Write-Warning ("{0}: robocopy 실패 (exit {1})" -f $AppEntry.Name, $code)
    }
}

# --- main ---
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot  = Split-Path -Parent $scriptDir
$configPath = Join-Path $scriptDir 'sync.psd1'
if (-not (Test-Path $configPath)) { throw "sync.psd1 을 찾을 수 없습니다: $configPath" }

$config = Import-PowerShellDataFile -Path $configPath
$apps = $config.Apps
if ($App) {
    $apps = @($apps | Where-Object { $_.Name -eq $App })
    if ($apps.Count -eq 0) { throw "sync.psd1 에 '$App' 항목이 없습니다." }
}

if (-not $BackupRoot) { $BackupRoot = Join-Path $repoRoot 'SyncData' }

Write-Host ("Mode      : {0}" -f $Mode)
Write-Host ("BackupRoot: {0}" -f $BackupRoot)
Write-Host ("Apps      : {0}" -f (($apps | ForEach-Object Name) -join ', '))

if ($Mode -eq 'Restore' -and -not $Force) {
    $reply = Read-Host "Restore 는 PC 의 설정을 SyncData 백업으로 덮어씁니다. 진행할까요? (y/N)"
    if ($reply -notmatch '^[Yy]') { Write-Host '취소됨.'; return }
}

if ($Mode -eq 'Backup' -and -not (Test-Path $BackupRoot)) {
    New-Item -ItemType Directory -Path $BackupRoot -Force | Out-Null
}

foreach ($entry in $apps) {
    Invoke-AppSync -AppEntry $entry -BackupRoot $BackupRoot -Mode $Mode
}

Write-Host "`n완료." -ForegroundColor Green
