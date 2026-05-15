@{
	# 동기화 대상 앱 목록.
	# Base   : APPDATA | LOCALAPPDATA | USERPROFILE
	# Path   : Base 아래 상대 경로 (앱 설정 폴더)
	# ExcludeDirs : robocopy /XD 로 전달될 폴더 이름 (단순 이름 또는 와일드카드)
	# ExcludeFiles: robocopy /XF 로 전달될 파일 이름/와일드카드
	Apps = @(
		@{ Name = 'CopyQ';      Base = 'APPDATA'; Path = 'CopyQ';
			ExcludeDirs = @('cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'Everything'; Base = 'APPDATA'; Path = 'Everything';
			ExcludeDirs = @('cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'FileZilla';  Base = 'APPDATA'; Path = 'FileZilla';
			ExcludeDirs = @('cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'lghub';      Base = 'APPDATA'; Path = 'lghub';
			ExcludeDirs = @('cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'Notepad++';  Base = 'APPDATA'; Path = 'Notepad++';
			ExcludeDirs = @('backup','cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'PicPick';    Base = 'APPDATA'; Path = 'PicPick';
			ExcludeDirs = @('cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'WinMerge';   Base = 'APPDATA'; Path = 'WinMerge';
			ExcludeDirs = @('Backup','cache'); ExcludeFiles = @('*.log') },

		@{ Name = 'PowerToys';  Base = 'USERPROFILE'; Path = 'Documents\PowerToys';
			ExcludeDirs = @(); ExcludeFiles = @() }
	)
}
