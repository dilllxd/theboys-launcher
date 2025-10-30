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

// Custom uninstall function to ensure complete cleanup
function UninstallAppFolder: Boolean;
var
  ResultCode: Integer;
begin
  // Try to delete the application folder using cmd with force and quiet options
  // This handles files that might be in use and ensures complete deletion
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
      Result := False;
    end;
  end;
end;

// Enhanced uninstall step to ensure complete cleanup
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // After the standard uninstall, try to remove any remaining files/folders
    UninstallAppFolder();
  end;
end;
