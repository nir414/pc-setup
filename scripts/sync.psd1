@{
    # 글로벌 제외 규칙 — 모든 폴더에 적용.
    # 확장자 패턴('*.log' 등)·명시적 파일명은 robocopy /XF, 그 외 폴더 이름은 /XD 로 분기.
    GlobalExcludes = @(
        'cache',
        '*.log',
        '*.backup.csv',
        '*.backup.ini',
        '*.backup.json',
        '*.tmp',
        '*.lock',
        '*.dat',
        '*.dat.*',
        '*.bak',
        '*.swp',
        '*.db'
    )

    Sections = @{
        APPDATA = @{
            Folders = @(
                @{ Path = 'CopyQ'; Excludes = @('copyq', 'copyq_tabs.ini.*') }
                @{ Path = 'DeepL_SE'; Excludes = @('logs', 'featureGates.db', 'featureUsageHistory.db', 'history.db', 'historyConsent.db') }
                @{ Path = 'Everything'; Excludes = @(
                    'Search History-1.5a.csv',
                    'Run History-1.5a.csv'
                )}
                @{ Path = 'FileZilla' }
                @{ Path = 'lghub' }
                @{ Path = 'Notepad++'; Excludes = @('backup') }
                @{ Path = 'PicPick' }
                @{ Path = 'WinMerge'; Excludes = @('Backup') }
            )
        }

        LOCALAPPDATA = @{
            Folders = @()
        }

        USERPROFILE = @{
            Folders = @(
                @{ Path = 'Documents\PowerToys'; Excludes = @('Backup') }
            )
        }

        ProgramData = @{
            Folders = @(
                @{ Path = 'Ant Renamer' }
            )
        }
    }
}
