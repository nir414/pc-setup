<#
간단 설정 동기화 스크립트

사용법 (PowerShell):
  ./sync.ps1                 # AppData + USERPROFILE 동시 적용
  ./sync.ps1 -DryRun         # 실제 복사 없이 미리보기(WhatIf)
  ./sync.ps1 -AppDataOnly    # AppData만 적용
  ./sync.ps1 -UserprofileOnly# USERPROFILE만 적용
  ./sync.ps1 -Backup         # 덮어쓰기 전 기존 파일 백업 (타임스탬프 폴더)
  ./sync.ps1 -BackupPath "D:\Backups" # 백업 경로 지정
  ./sync.ps1 -Force          # 확인 없이 바로 실행 (기본은 확인 필요)
	./sync.ps1 -FullDirBackup  # (선택) 폴더 전체 백업 (기본은 파일 단위만 백업)
#>

param(
	[switch]$AppDataOnly,
	[switch]$UserprofileOnly,
	[switch]$DryRun,
	[switch]$Backup,
	[string]$BackupPath = "",
	[switch]$Force,
	[switch]$FullDirBackup,  # 기존 방식(폴더 전체 백업)이 필요할 때만 사용 (현재 기본)
	[switch]$ListBackups     # 기존 백업 목록 출력
)$ErrorActionPreference = 'Stop'

# .syncconfig 파일 읽기 (exclude/include 패턴 분리)
function Get-BackupPatterns {
	param([string]$repoRoot)
	
	$ignoreFile = Join-Path $repoRoot '.syncconfig'
	$excludePatterns = @()
	$includePatterns = @()
	
	if (Test-Path -LiteralPath $ignoreFile) {
		$lines = Get-Content -LiteralPath $ignoreFile -ErrorAction SilentlyContinue
		foreach ($line in $lines) {
			$trimmed = $line.Trim()
			# 주석이나 빈 줄 제외
			if ($trimmed -and -not $trimmed.StartsWith('#')) {
				if ($trimmed.StartsWith('+')) {
					# 포함 패턴 (USERPROFILE allowlist)
					$includePatterns += $trimmed.Substring(1)  # + 제거
				} else {
					# 제외 패턴 (AppData blacklist)
					$excludePatterns += $trimmed
				}
			}
		}
		
		if ($excludePatterns.Count -gt 0 -or $includePatterns.Count -gt 0) {
			Write-Host "백업 제어 패턴 로드: 제외 $($excludePatterns.Count)개, 포함 $($includePatterns.Count)개" -ForegroundColor DarkGray
		}
	}
	
	return @{
		Exclude = $excludePatterns
		Include = $includePatterns
	}
}

# 패턴 매칭 함수
function Test-ExcludePattern {
	param(
		[string]$itemPath,
		[string]$baseLabel,
		[array]$patterns
	)

	$relativePath = "$baseLabel/$itemPath"
	foreach ($pattern in $patterns) {
		# 패턴 일치 또는 디렉터리 prefix 일치(하위 모두 제외)
		if ($relativePath -like $pattern) { return $true }
		if ($relativePath.StartsWith($pattern)) { return $true }
	}
	return $false
}

# 백업 폴더 생성 (필요 시)
$backupRoot = $null
function Backup-ExistingFiles {
	param(
		[Parameter(Mandatory=$true)][string]$sourcePath,  # 저장소 경로 (예: repo/AppData)
		[Parameter(Mandatory=$true)][string]$targetPath,  # 시스템 경로 (예: %APPDATA%)
		[Parameter(Mandatory=$true)][string]$backupBase,  # 백업 루트
		[Parameter(Mandatory=$true)][string]$label,       # 레이블(AppData|USERPROFILE)
		[array]$excludePatterns = @(),
		[array]$includePatterns = @()
	)

	$backupDest = Join-Path $backupBase $label
	if (-not (Test-Path -LiteralPath $sourcePath)) { return }

	# 저장소의 최상위 항목(폴더/파일) 목록을 기준으로 전체 대상 백업
	$topItems = Get-ChildItem -LiteralPath $sourcePath -Force -ErrorAction SilentlyContinue
	if (-not $topItems) { return }

	# 동작 모드 결정
	$useAllowlist = ($label -eq 'USERPROFILE' -and $includePatterns.Count -gt 0)
	
	if ($useAllowlist) {
		Write-Host "  선택적 백업 모드 (포함 패턴: $($includePatterns.Count)개)" -ForegroundColor DarkYellow
		
		# allowlist 모드: 지정된 경로만 직접 백업
		$copied = 0; $skipped = 0; $missing = 0
		
		foreach ($pattern in $includePatterns) {
			# 패턴에서 label 제거하여 실제 경로 얻기
			$relativePath = $pattern.Replace("$label/", "")
			$targetPath = Join-Path $targetPath $relativePath
			$backupPath = Join-Path $backupDest $relativePath
			
			if (-not (Test-Path -LiteralPath $targetPath)) {
				Write-Host "    - 존재하지 않음: $relativePath" -ForegroundColor DarkGray
				$missing++
				continue
			}
			
			Write-Host "    → 직접 백업: $relativePath" -ForegroundColor Cyan
			
			$destDir = Split-Path -Parent $backupPath
			if ($destDir -and -not (Test-Path -LiteralPath $destDir)) { 
				New-Item -ItemType Directory -Path $destDir -Force | Out-Null 
			}
			
			try {
				if (Test-Path -LiteralPath $targetPath -PathType Container) {
					Copy-Item -LiteralPath $targetPath -Destination $backupPath -Recurse -Force -ErrorAction Stop
				} else {
					Copy-Item -LiteralPath $targetPath -Destination $backupPath -Force -ErrorAction Stop
				}
				$copied++
			} catch { 
				Write-Warning "    ✗ 실패: $relativePath - $_"
			}
		}
		
		Write-Host "  결과: 직접 백업 $copied, 제외 $skipped, 없음 $missing" -ForegroundColor DarkYellow
		return
	}
	
	# 기존 blacklist 모드 (AppData)
	Write-Host "  폴더 단위 백업 시작...(최상위 기준)" -ForegroundColor DarkYellow
	$copied = 0; $skipped = 0; $missing = 0

	foreach ($repoItem in $topItems) {
		$name = $repoItem.Name
		$targetFull = Join-Path $targetPath $name
		$backupFull = Join-Path $backupDest $name

		# USERPROFILE: Allowlist 모드 (포함 패턴에 없으면 스킵)
		if ($useAllowlist) {
			# allowlist에서 이 최상위 폴더와 관련된 패턴이 있는지 확인
			$hasRelevantPattern = $false
			foreach ($pattern in $includePatterns) {
				$topFolder = "$label/$name"
				# 패턴이 이 폴더 또는 하위를 대상으로 하는가?
				if ($pattern -eq $topFolder -or $pattern.StartsWith($topFolder + "/")) {
					$hasRelevantPattern = $true
					break
				}
			}
			if (-not $hasRelevantPattern) {
				Write-Host "    ⊘ 포함 목록 외: $name" -ForegroundColor DarkGray
				$skipped++
				continue
			}
		} else {
			# AppData: Blacklist 모드 (제외 패턴에 매칭되면 스킵)
			if (Test-ExcludePattern -itemPath $name -baseLabel $label -patterns $excludePatterns) {
				Write-Host "    ⊗ 상위 제외: $name" -ForegroundColor DarkGray
				$skipped++
				continue
			}
		}

		if (-not (Test-Path -LiteralPath $targetFull)) {
			Write-Host "    - 존재하지 않음(시스템): $name" -ForegroundColor DarkGray
			$missing++
			continue
		}

		if ($repoItem.PSIsContainer) {
			Write-Host "    → 디렉터리 처리: $name" -ForegroundColor Cyan
			# 하위 파일 나열 (디렉터리 전체 백업, 단 패턴 적용)
			$allFiles = Get-ChildItem -LiteralPath $targetFull -Recurse -Force -ErrorAction SilentlyContinue | Where-Object { -not $_.PSIsContainer }
			foreach ($file in $allFiles) {
				$relSub = ($file.FullName.Substring($targetFull.Length) -replace '^[\\/]+','')
				$relUnified = ($name + '/' + $relSub).Replace('\\','/')
				$fullRelPath = "$label/$relUnified"
				
				# allowlist 모드: 포함 패턴에 매칭되는 파일만 백업
				if ($useAllowlist) {
					$shouldInclude = $false
					foreach ($pattern in $includePatterns) {
						if ($fullRelPath -like $pattern -or $fullRelPath.StartsWith($pattern) -or $pattern.StartsWith($fullRelPath)) {
							$shouldInclude = $true
							break
						}
					}
					if (-not $shouldInclude) {
						Write-Host "       ⊘ 포함 목록 외: $relUnified" -ForegroundColor DarkGray
						continue
					}
				} else {
					# 기존 blacklist 모드: 제외 패턴에 매칭되면 스킵
					if (Test-ExcludePattern -itemPath $relUnified -baseLabel $label -patterns $excludePatterns) {
						Write-Host "       ⊗ 제외: $relUnified" -ForegroundColor DarkGray
						continue
					}
				}
				
				# 대상 백업 경로 생성
				$destFilePath = Join-Path $backupFull $relSub
				$destDir = Split-Path -Parent $destFilePath
				if ($destDir -and -not (Test-Path -LiteralPath $destDir)) { New-Item -ItemType Directory -Path $destDir -Force | Out-Null }
				try {
					Copy-Item -LiteralPath $file.FullName -Destination $destFilePath -Force -ErrorAction Stop
				} catch { Write-Warning "       ✗ 실패: $relUnified - $_" }
			}
			$copied++
		} else {
			# 단일 파일
			$relFile = $name
			if (-not $useAllowlist -and (Test-ExcludePattern -itemPath $relFile -baseLabel $label -patterns $excludePatterns)) {
				Write-Host "    ⊗ 제외: $relFile" -ForegroundColor DarkGray
				$skipped++
				continue
			}
			$destDir = Split-Path -Parent $backupFull
			if ($destDir -and -not (Test-Path -LiteralPath $destDir)) { New-Item -ItemType Directory -Path $destDir -Force | Out-Null }
			try { Copy-Item -LiteralPath $targetFull -Destination $backupFull -Force -ErrorAction Stop; $copied++ } catch { Write-Warning "    ✗ 실패: $name - $_" }
		}
	}

	Write-Host "  결과: 상위단위 백업 $copied, 제외 $skipped, 없음 $missing" -ForegroundColor DarkYellow
}

function Sync-Path {
	param(
		[Parameter(Mandatory=$true)][string]$src,
		[Parameter(Mandatory=$true)][string]$dest,
		[string]$backupRoot = "",
		[string]$label = "",
		[array]$excludePatterns = @(),
		[array]$includePatterns = @()
	)

	if (-not (Test-Path -LiteralPath $src)) {
		Write-Warning "건너뜀: 원본 경로가 없습니다 -> $src"
		return
	}

	# 백업 수행 (Backup 플래그가 켜져 있고 DryRun이 아닐 때)
	if ($backupRoot -and $label) {
		Backup-ExistingFiles -sourcePath $src -targetPath $dest -backupBase $backupRoot -label $label -excludePatterns $excludePatterns -includePatterns $includePatterns
	}

	Write-Host "복사: $src -> $dest" -ForegroundColor Cyan
	
	# 상세 복사 로그 (항목별)
	$items = Get-ChildItem -LiteralPath $src -Force -ErrorAction SilentlyContinue
	if ($items) {
		Write-Host "  항목별 복사 진행 중..." -ForegroundColor DarkYellow
		
		foreach ($item in $items) {
			# 대상 경로 (디렉터리는 내용 병합 복사)
			$destDir = Join-Path $dest $item.Name
			$sourceToCopy = if ($item.PSIsContainer) { Join-Path $item.FullName '*' } else { $item.FullName }

			# 크기 정보
			$sizeInfo = ""
			if ($item.PSIsContainer) {
				$itemCount = (Get-ChildItem -LiteralPath $item.FullName -Recurse -Force -ErrorAction SilentlyContinue | Measure-Object).Count
				$sizeInfo = "($itemCount 개 항목)"
			} else {
				$sizeKB = [Math]::Round((Get-Item -LiteralPath $item.FullName).Length / 1KB, 2)
				$sizeInfo = "($sizeKB KB)"
			}

			Write-Host "    → 복사: $($item.Name) $sizeInfo" -ForegroundColor Cyan

			$startTime = Get-Date

			if ($DryRun) {
				Copy-Item -Path $sourceToCopy -Destination $destDir -Recurse -Force -WhatIf
			} else {
				try {
					if ($item.PSIsContainer -and -not (Test-Path -LiteralPath $destDir)) { New-Item -ItemType Directory -Path $destDir -Force | Out-Null }
					Copy-Item -Path $sourceToCopy -Destination $destDir -Recurse -Force -ErrorAction Stop
					$elapsed = ((Get-Date) - $startTime).TotalSeconds
					Write-Host "      ✓ 완료 ($([Math]::Round($elapsed, 2))초)" -ForegroundColor Green
				} catch {
					Write-Warning "      ✗ 실패: $_"
				}
			}
		}
		Write-Host ""
	}
}

# 저장소 루트
$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path

# 백업 패턴 로드 (exclude/include 분리)
$backupPatterns = Get-BackupPatterns -repoRoot $repoRoot
$excludePatterns = $backupPatterns.Exclude
$includePatterns = $backupPatterns.Include

# 안전 확인 (DryRun이나 Force가 아닐 때)
if (-not $DryRun -and -not $Force) {
    Write-Host ""
    Write-Host "⚠️  주의: 기존 설정 파일을 덮어씁니다!" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "동기화 대상:" -ForegroundColor Cyan
    
    if (-not $UserprofileOnly) {
        Write-Host "  - AppData: $env:APPDATA" -ForegroundColor White
    }
    if (-not $AppDataOnly) {
        Write-Host "  - USERPROFILE: $env:USERPROFILE" -ForegroundColor White
    }
    
    if ($Backup) {
        Write-Host ""
        Write-Host "✓ 백업 활성화됨" -ForegroundColor Green
    } else {
        Write-Host ""
        Write-Host "✗ 백업 없음 (권장: -Backup 추가)" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "계속하시겠습니까? (Y/N): " -ForegroundColor Yellow -NoNewline
    $response = Read-Host
    
    if ($response -ne 'Y' -and $response -ne 'y') {
        Write-Host "취소되었습니다." -ForegroundColor Gray
        Write-Host ""
        Write-Host "💡 팁:" -ForegroundColor Cyan
        Write-Host "  - 미리보기: ./sync.ps1 -DryRun" -ForegroundColor White
        Write-Host "  - 백업하며 실행: ./sync.ps1 -Backup" -ForegroundColor White
        Write-Host "  - 확인 생략: ./sync.ps1 -Force" -ForegroundColor White
        exit 0
    }
    
    Write-Host ""
}

###############################################################################
# 백업 루트 결정 ( -Backup 플래그가 있을 때 1회 설정 )
# 기본 규칙:
#   - 사용자가 -BackupPath 를 지정하면 그 경로 아래에 timestamp 폴더 생성
#   - 아니면 AppData를 포함할 경우 AppData 상위(C:\Users\<User>\AppData)에 생성
#   - AppData를 제외하고 USERPROFILE만이면 USERPROFILE 루트에 생성
###############################################################################
if ($ListBackups -and -not $Backup) {
	Write-Host "\n📂 기존 백업 목록" -ForegroundColor Cyan
	$pathsToScan = @((Split-Path -Parent $env:APPDATA), $env:USERPROFILE)
	foreach ($scanRoot in $pathsToScan) {
		if (-not (Test-Path -LiteralPath $scanRoot)) { continue }
		$backups = Get-ChildItem -LiteralPath $scanRoot -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like 'backup_*' }
		if ($backups) {
			Write-Host "  위치: $scanRoot" -ForegroundColor DarkYellow
			foreach ($b in $backups) {
				# 간단 크기(파일 수) 측정 (무거울 수 있어 Recurse 제한 없음)
				$fileCount = (Get-ChildItem -LiteralPath $b.FullName -Recurse -Force -ErrorAction SilentlyContinue | Where-Object { -not $_.PSIsContainer } | Measure-Object).Count
				Write-Host "    - $($b.Name) (파일 $fileCount 개)" -ForegroundColor White
			}
		} else {
			Write-Host "  위치: $scanRoot -> (백업 없음)" -ForegroundColor DarkGray
		}
	}
	Write-Host "\n(목록만 표시했습니다. 동기화를 실행하려면 -ListBackups 를 제거하세요.)" -ForegroundColor Gray
	exit 0
}

if ($Backup) {
	$timestamp = (Get-Date).ToString('yyyyMMdd_HHmmss')
	if ($BackupPath -and $BackupPath.Trim()) {
		$backupRoot = Join-Path $BackupPath "backup_$timestamp"
	} else {
		if (-not $UserprofileOnly) { # AppData 포함
			$backupRoot = Join-Path (Split-Path -Parent $env:APPDATA) "backup_$timestamp"
		} else { # USERPROFILE만
			$backupRoot = Join-Path $env:USERPROFILE "backup_$timestamp"
		}
	}
	if (-not (Test-Path -LiteralPath $backupRoot)) { New-Item -ItemType Directory -Path $backupRoot -Force | Out-Null }
	Write-Host "백업 루트 준비됨: $backupRoot" -ForegroundColor Yellow
} elseif ($BackupPath) {
	Write-Warning "-BackupPath 는 -Backup 과 함께 사용될 때만 의미가 있습니다."
}

$doAppData    = -not $UserprofileOnly
$doUserFolder = -not $AppDataOnly

if ($doAppData) {
	$srcAppData  = Join-Path $repoRoot 'AppData'
	$destAppData = $env:APPDATA
	Sync-Path -src $srcAppData -dest $destAppData -backupRoot $backupRoot -label 'AppData' -excludePatterns $excludePatterns -includePatterns $includePatterns
}

if ($doUserFolder) {
	$srcUser  = Join-Path $repoRoot 'USERPROFILE'
	$destUser = $env:USERPROFILE
	Sync-Path -src $srcUser -dest $destUser -backupRoot $backupRoot -label 'USERPROFILE' -excludePatterns $excludePatterns -includePatterns $includePatterns
}

Write-Host "동기화 완료 (DryRun=$DryRun, Backup=$Backup)" -ForegroundColor Green
if ($backupRoot) {
	Write-Host "백업 위치: $backupRoot" -ForegroundColor Cyan
}
