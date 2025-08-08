
# 윈도우 시스템 다크 모드 적용
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name AppsUseLightTheme -Type DWord -Value 0
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name SystemUsesLightTheme -Type DWord -Value 0

# 제목/테두리에 강조 색 표시
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\DWM" -Name ColorPrevalence -Type DWord -Value 1
# (선택) 시작/작업 표시줄에도 강조 색 표시
# Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name ColorPrevalence -Type DWord -Value 1

# 파일 탐색기에서 확장자·숨김 파일 표시
$adv = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced"
Set-ItemProperty -Path $adv -Name HideFileExt -Type DWord -Value 0
Set-ItemProperty -Path $adv -Name Hidden -Type DWord -Value 1
# (선택) 보호된 운영체제 파일까지 표시하려면 아래 주석 해제
# Set-ItemProperty -Path $adv -Name ShowSuperHidden -Type DWord -Value 1

# 시작 메뉴 > 폴더 전부 켬
$start = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Start"
$names = @(
    "ShowSettings",
    "ShowFileExplorer",
    "ShowDocuments",
    "ShowDownloads",
    "ShowMusic",
    "ShowPictures",
    "ShowVideos",
    "ShowNetwork",
    "ShowPersonalFolder"
)
if (-not (Test-Path $start)) { New-Item $start -Force | Out-Null }
foreach ($n in $names) {
    New-ItemProperty -Path $start -Name $n -PropertyType DWord -Value 1 -Force | Out-Null
}

# 변경 사항 즉시 적용 (탐색기 재시작)
Get-Process explorer -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Process explorer.exe