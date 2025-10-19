; TheBoys Launcher Installer Script
; Creates a professional Windows installer with shortcuts, uninstaller, and certificate installation

#define MyAppName "TheBoysLauncher"
#define MyAppVersion "2.0.0"
#define MyAppPublisher "Dylan"
#define MyAppURL "https://github.com/dilllxd/theboys-launcher"
#define MyAppExeName "TheBoysLauncher.exe"
#define MyAppAssocName "TheBoys Minecraft Modpack Launcher"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
; Do not use the same AppId value in installers for other applications.
AppId={{D5E2B1A3-7C4F-4A2D-9E8F-1A2B3C4D5E6F}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={userappdata}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
LicenseFile=LICENSE.txt
OutputDir=installer
OutputBaseFilename=TheBoysLauncher-Setup-{#MyAppVersion}
SetupIconFile=icon.ico
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=none
ChangesAssociations=yes
; Create a dedicated app folder to avoid conflicts
CreateAppDir=yes
CreateUninstallRegKey=yes
; Ensure we can overwrite existing files
DirExistsWarning=no

; Sign the installer with our certificate (optional)
; SignTool=signtool $p

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; OnlyBelowVersion: 6.1; Check: IsAdminInstallMode
Name: "associate"; Description: "Associate .theboys files with {#MyAppName}"; GroupDescription: "File associations:"; Flags: unchecked

[Files]
Source: "TheBoysLauncher.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "icon.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "theboys-launcher-cert.pfx"; DestDir: "{app}"; Flags: ignoreversion; Tasks: ; OnlyBelowVersion: 0
Source: "LICENSE.txt"; DestDir: "{app}"; Flags: ignoreversion; Tasks:
; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"; Tasks: desktopicon
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"; Tasks: quicklaunchicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Registry]
; File association
Root: HKCR; Subkey: ".theboys"; ValueType: string; ValueName: ""; ValueData: "{#MyAppAssocName}"; Flags: uninsdeletevalue; Tasks: associate
Root: HKCR; Subkey: "{#MyAppAssocName}"; ValueType: string; ValueName: ""; ValueData: "{#MyAppAssocName}"; Flags: uninsdeletekey; Tasks: associate
Root: HKCR; Subkey: "{#MyAppAssocName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName},0"; Tasks: associate
Root: HKCR; Subkey: "{#MyAppAssocName}\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppExeName}"" ""%1"""; Tasks: associate

[UninstallDelete]
Type: filesandordirs; Name: "{app}\logs"
Type: filesandordirs; Name: "{app}\instances"
Type: filesandordirs; Name: "{app}\cache"
Type: filesandordirs; Name: "{app}\prism"
Type: filesandordirs; Name: "{app}\util"

; Also clean up user data directory (requires custom code)

[Code]
// Function to install our certificate to Trusted Root Certification Authorities
procedure InstallCertificate;
var
    ResultCode: Integer;
    CertPath: string;
begin
    CertPath := ExpandConstant('{app}\theboys-launcher-cert.pfx');

    // Try to install the certificate to Trusted Root Certification Authorities
    if FileExists(CertPath) then
    begin
        Log('Installing certificate to Trusted Root Certification Authorities...');

        // Use PowerShell to install the certificate
        if Exec('powershell.exe', '-Command "Import-PfxCertificate -FilePath ''' + CertPath + ''' -CertStoreLocation ''Cert:\CurrentUser\Root'' -Password (ConvertTo-SecureString -String ''TheBoys2025!'' -Force -AsPlainText)"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
        begin
            if ResultCode = 0 then
                Log('Certificate installed successfully')
            else
                Log('Certificate installation failed with code: ' + IntToStr(ResultCode));
        end
        else
            Log('Failed to execute certificate installation');
    end
    else
        Log('Certificate file not found: ' + CertPath);
end;

// Function to uninstall our certificate
procedure UninstallCertificate;
var
    ResultCode: Integer;
begin
    Log('Attempting to remove certificate from Trusted Root Certification Authorities...');

    // Use PowerShell to find and remove our certificate
    if Exec('powershell.exe', '-Command "Get-ChildItem -Path ''Cert:\CurrentUser\Root'' | Where-Object {$_.Subject -like ''*TheBoysLauncher*''} | Remove-Item"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
    begin
        if ResultCode = 0 then
            Log('Certificate removed successfully')
        else
            Log('Certificate removal failed with code: ' + IntToStr(ResultCode));
    end
    else
        Log('Failed to execute certificate removal');
end;

// Install certificate after files are copied
procedure CurStepChanged(CurStep: TSetupStep);
begin
    if CurStep = ssPostInstall then
    begin
        // Install certificate to trusted store
        InstallCertificate;

        // Create initial directories
        CreateDir(ExpandConstant('{app}\logs'));
        CreateDir(ExpandConstant('{app}\instances'));
        CreateDir(ExpandConstant('{app}\cache'));
    end;
end;

// Remove certificate during uninstallation
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
    UserDataPath: string;
    ResultCode: Integer;
begin
    if CurUninstallStep = usUninstall then
    begin
        // Kill any running launcher processes before uninstall
        Log('Stopping any running TheBoysLauncher processes...');
        Exec('taskkill.exe', '/F /IM TheBoysLauncher.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
        // Also kill any setup processes that might be running
        Exec('taskkill.exe', '/F /IM TheBoysLauncher-Setup*.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
        if ResultCode = 0 then
            Log('Successfully stopped launcher processes')
        else
            Log('No running launcher processes found or failed to stop them');

        // Wait a moment for processes to fully terminate
        Sleep(2000);

        // Remove certificate from trusted store
        UninstallCertificate;

        // Clean up app directory (everything is in one place now)
        UserDataPath := ExpandConstant('{app}');
        if DirExists(UserDataPath) then
        begin
            Log('Removing user data directory: ' + UserDataPath);
            if DelTree(UserDataPath, True, True, True) then
                Log('User data directory removed successfully')
            else
                Log('Failed to remove user data directory');
        end;
    end;
end;

// Custom page for certificate information
function ShouldSkipPage(PageID: Integer): Boolean;
begin
    Result := False;
end;

// Check if .NET Framework is available (required for PowerShell)
function InitializeSetup(): Boolean;
var
    ResultCode: Integer;
    PowerShellVersion: string;
begin
    Result := True;

    // Simple check: if AppData\TheBoysLauncher folder exists, warn user
    if DirExists(ExpandConstant('{userappdata}\TheBoysLauncher')) then
    begin
        Log('Found existing installation in AppData');
        if MsgBox('TheBoys Launcher appears to already be installed. Continue with installation?', mbConfirmation, MB_YESNO) = IDYES then
        begin
            Log('User chose to continue installation');
        end
        else
        begin
            Result := False;
            Exit;
        end;
    end;

    // Check if PowerShell is available
    if Exec('powershell.exe', '-Command "$PSVersionTable.PSVersion.Major"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
    begin
        if ResultCode <> 0 then
        begin
            MsgBox('PowerShell is required for this installation but is not available. Please install PowerShell and try again.', mbError, MB_OK);
            Result := False;
        end;
    end
    else
    begin
        MsgBox('PowerShell is required for this installation but is not available. Please install PowerShell and try again.', mbError, MB_OK);
        Result := False;
    end;
end;