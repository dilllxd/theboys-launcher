; TheBoysLauncher Installer Script
; Creates a professional Windows installer with shortcuts and uninstaller

#define MyAppName "TheBoysLauncher"
#define MyAppVersion "v3.2.68"
#define MyAppPublisher "Dylan"
#define MyAppURL "https://github.com/dilllxd/theboyslauncher"
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
DefaultDirName={localappdata}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
LicenseFile=LICENSE.txt
OutputDir=installer
OutputBaseFilename=TheBoysLauncher-Setup-{#MyAppVersion}
SetupIconFile=icon.ico
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern

; Enable restart manager to handle files in use during uninstall
CloseApplications=yes
RestartApplications=no
CloseApplicationsFilter=*.exe,*.dll,*.log

; Ensure we can overwrite existing files
DirExistsWarning=no

; Enable custom directory selection
AppendDefaultDirName=no
UsePreviousAppDir=yes
UsePreviousGroup=yes
UsePreviousTasks=yes

; Registry entries for launcher to read installation path
ChangesAssociations=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1

[Files]
Source: "TheBoysLauncher.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "icon.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "LICENSE.txt"; DestDir: "{app}"; Flags: ignoreversion
; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"; Comment: "TheBoys Minecraft Modpack Launcher"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"; Comment: "Uninstall TheBoysLauncher"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"; Tasks: desktopicon; Comment: "TheBoys Minecraft Modpack Launcher"
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: quicklaunchicon; IconFilename: "{app}\icon.ico"; Comment: "TheBoys Minecraft Modpack Launcher"

[Run]
; Launch application after installation - controlled by checkbox
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent unchecked

[Registry]
; Store installation path for launcher to read
Root: HKLM; Subkey: "SOFTWARE\{#MyAppName}"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\{#MyAppName}"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey


[UninstallDelete]
; Delete specific subdirectories first
Type: filesandordirs; Name: "{app}\logs"
Type: filesandordirs; Name: "{app}\instances"
Type: filesandordirs; Name: "{app}\cache"
; Delete the entire application directory and all its contents
Type: filesandordirs; Name: "{app}"

[Code]
// Windows API constants and functions for process management
const
  WM_CLOSE = $0010;
  WM_QUIT = $0012;
  PROCESS_TERMINATE = $0001;
  SMTO_BLOCK = $0001;
  SMTO_ABORTIFHUNG = $0002;
  TIMEOUT_MS = 5000; // 5 seconds timeout for graceful shutdown
  
// Use Integer instead of Windows-specific types for Inno Setup compatibility
function FindWindowByClassName(WindowClassName: string): Integer;
external 'FindWindowA@user32.dll stdcall';

function SendMessageTimeout(hWnd: Integer; Msg: Integer; wParam: Integer; lParam: Integer; fuFlags: Integer; uTimeout: Integer; out lpdwResult: Integer): Integer;
external 'SendMessageTimeoutA@user32.dll stdcall';

function PostMessage(hWnd: Integer; Msg: Integer; wParam: Integer; lParam: Integer): Boolean;
external 'PostMessageA@user32.dll stdcall';

function IsWindow(hWnd: Integer): Boolean;
external 'IsWindow@user32.dll stdcall';

function GetWindowThreadProcessId(hWnd: Integer; out lpdwProcessId: Cardinal): Cardinal;
external 'GetWindowThreadProcessId@user32.dll stdcall';

function OpenProcess(dwDesiredAccess: Cardinal; bInheritHandle: Boolean; dwProcessId: Cardinal): Integer;
external 'OpenProcess@kernel32.dll stdcall';

function TerminateProcess(hProcess: Integer; uExitCode: Cardinal): Boolean;
external 'TerminateProcess@kernel32.dll stdcall';

function CloseHandle(hObject: Integer): Boolean;
external 'CloseHandle@kernel32.dll stdcall';

// Function to check if TheBoysLauncher is currently running
function IsAppRunning: Boolean;
var
  ResultCode: Integer;
begin
  // Use tasklist to check if TheBoysLauncher.exe is running
  Result := Exec(ExpandConstant('{cmd}'), '/C tasklist /FI "IMAGENAME eq TheBoysLauncher.exe" 2>NUL | find /I "TheBoysLauncher.exe" >NUL', '',
                 SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Result := (ResultCode = 0);
end;

// Function to find and close TheBoysLauncher windows gracefully
function CloseAppGracefully: Boolean;
var
  WindowHandle: Integer;
  ResultCode: Integer;
begin
  Result := False;
  
  // Try to find the main window by class name or title
  WindowHandle := FindWindowByClassName('TTheBoysLauncher');
  if WindowHandle = 0 then
    WindowHandle := FindWindowByClassName('TheBoysLauncher');
  
  if WindowHandle <> 0 then
  begin
    // Send WM_CLOSE message to request graceful shutdown
    if SendMessageTimeout(WindowHandle, WM_CLOSE, 0, 0, SMTO_BLOCK or SMTO_ABORTIFHUNG, TIMEOUT_MS, ResultCode) <> 0 then
    begin
      // Wait a bit for the application to close
      Sleep(2000);
      
      // Check if window still exists
      if not IsWindow(WindowHandle) then
      begin
        Result := True;
        Exit;
      end;
    end;
    
    // If WM_CLOSE didn't work, try WM_QUIT
    if PostMessage(WindowHandle, WM_QUIT, 0, 0) then
    begin
      Sleep(2000);
      if not IsWindow(WindowHandle) then
      begin
        Result := True;
        Exit;
      end;
    end;
  end;
end;

// Function to forcefully terminate TheBoysLauncher processes
function TerminateAppForcefully: Boolean;
var
  ResultCode: Integer;
begin
  // Use taskkill to forcefully terminate all TheBoysLauncher.exe processes
  Result := Exec(ExpandConstant('{cmd}'), '/C taskkill /F /IM "TheBoysLauncher.exe" /T', '',
                 SW_HIDE, ewWaitUntilTerminated, ResultCode);
  
  // Wait a moment for processes to terminate
  Sleep(1000);
  
  // Check if we were successful
  Result := (ResultCode = 0) or (ResultCode = 128); // 128 means no processes found
end;

// Function to close TheBoysLauncher with multiple methods and user notification
function EnsureAppClosed: Boolean;
var
  AppWasRunning: Boolean;
  RetryCount: Integer;
  UserResponse: Integer;
begin
  Result := True;
  AppWasRunning := False;
  RetryCount := 0;
  
  // Check if the application is running
  if IsAppRunning then
  begin
    AppWasRunning := True;
    
    // Notify user that we need to close the application
    UserResponse := MsgBox('TheBoysLauncher is currently running and needs to be closed for uninstallation to continue.' + #13#13 +
                          'Click "Yes" to close the application automatically.' + #13 +
                          'Click "No" to cancel the uninstallation.',
                          mbConfirmation, MB_YESNO);
    
    if UserResponse = IDNO then
    begin
      Result := False;
      Exit;
    end;
    
    // Try to close the application gracefully first
    if CloseAppGracefully then
    begin
      // Check if it actually closed
      if not IsAppRunning then
        Exit;
    end;
    
    // If graceful closing failed, try force termination
    if TerminateAppForcefully then
    begin
      // Check if it actually closed
      if not IsAppRunning then
        Exit;
    end;
    
    // If still running, try multiple times with user notification
    while (RetryCount < 3) and IsAppRunning do
    begin
      Inc(RetryCount);
      if RetryCount < 3 then
      begin
        UserResponse := MsgBox('TheBoysLauncher is still running. Attempt ' + IntToStr(RetryCount) + ' of 3.' + #13#13 +
                              'Click "Yes" to try again.' + #13 +
                              'Click "No" to cancel the uninstallation.',
                              mbConfirmation, MB_YESNO);
        
        if UserResponse = IDNO then
        begin
          Result := False;
          Exit;
        end;
      end;
      
      // Try force termination again
      TerminateAppForcefully;
      Sleep(2000); // Wait longer between retries
    end;
    
    // Final check
    if IsAppRunning then
    begin
      MsgBox('Unable to close TheBoysLauncher after multiple attempts.' + #13#13 +
             'Please close the application manually and try again.',
             mbError, MB_OK);
      Result := False;
    end;
  end;
end;

// Enhanced uninstall function with retry logic for locked files
function UninstallAppFolder: Boolean;
var
  ResultCode: Integer;
  RetryCount: Integer;
begin
  Result := False;
  RetryCount := 0;
  
  while (RetryCount < 3) and not Result do
  begin
    Inc(RetryCount);
    
    // Try to delete the application folder using cmd with force and quiet options
    if Exec(ExpandConstant('{cmd}'), '/C rmdir /S /Q "' + ExpandConstant('{app}') + '"', '',
            SW_HIDE, ewWaitUntilTerminated, ResultCode) then
    begin
      Result := True;
    end
    else
    begin
      // If the above fails, try with PowerShell as a fallback
      if Exec(ExpandConstant('{powershell}'), '-Command "Remove-Item -Path \"' + ExpandConstant('{app}') + '\" -Recurse -Force -ErrorAction SilentlyContinue"', '',
              SW_HIDE, ewWaitUntilTerminated, ResultCode) then
      begin
        Result := True;
      end
      else
      begin
        // Wait before retrying
        if RetryCount < 3 then
          Sleep(2000);
      end;
    end;
  end;
end;

// Custom checkbox for launch after installation
function ShouldLaunchAfterInstall: Boolean;
begin
  // Check if RunList has any items before accessing index 0
  if WizardForm.RunList.Items.Count > 0 then
    Result := WizardForm.RunList.Checked[0]
  else
    Result := False; // Default value if RunList is empty
end;

// Initialize the checkboxes on the finishing page
procedure InitializeWizard;
begin
  // Check if RunList has any items before accessing index 0
  if WizardForm.RunList.Items.Count > 0 then
    // Ensure the launch checkbox is unchecked by default
    WizardForm.RunList.Checked[0] := False;
end;

// Enhanced uninstall step to ensure complete cleanup
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
  begin
    // Before uninstallation starts, ensure the application is closed
    if not EnsureAppClosed then
    begin
      // Abort the uninstallation if we can't close the app
      Abort;
    end;
  end
  else if CurUninstallStep = usPostUninstall then
  begin
    // After the standard uninstall, try to remove any remaining files/folders
    UninstallAppFolder();
  end;
end;
