; TheBoys Launcher Windows Installer Script
; Requires NSIS (Nullsoft Scriptable Install System)

!define APPNAME "TheBoys Launcher"
!define APPVERSION "1.0.0"
!define APPID "TheBoysLauncher"
!define APPEXE "TheBoysLauncher.exe"
!define APPURL "https://github.com/dilllxd/theboys-launcher"
!define APPDESCRIPTION "Modern Minecraft Modpack Launcher"
!define APPPUBLISHER "TheBoys"
!define APPYEAR "2024"

; Include Modern UI
!include "MUI2.nsh"

; General settings
Name "${APPNAME}"
OutFile "TheBoysLauncher-Setup-${APPVERSION}.exe"
InstallDir "$PROGRAMFILES\${APPNAME}"
InstallDirRegKey HKLM "Software\${APPID}" "InstallPath"
RequestExecutionLevel admin

; Directory selection with custom text
DirText "Please select the folder where you would like to install ${APPNAME}.$\n$\nThis will install the application files, but your saved games and settings will be stored in your user profile directory." \
     "Select Installation Folder" \
     "Select the folder to install ${APPNAME} in:"

; Interface settings
!define MUI_ABORTWARNING
!define MUI_ICON "icon.ico"
!define MUI_UNICON "icon.ico"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_BITMAP "header.bmp"
!define MUI_COMPONENTSPAGE_SMALLDESC
!define MUI_FINISHPAGE_LINK "Visit our website for support and updates"
!define MUI_FINISHPAGE_LINK_LOCATION "${APPURL}"
!define MUI_FINISHPAGE_RUN "$INSTDIR\${APPEXE}"
!define MUI_FINISHPAGE_RUN_TEXT "Launch TheBoys Launcher"

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE.txt"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Languages
!insertmacro MUI_LANGUAGE "English"

; Installer sections
Section "Core Files" SecCore
    SectionIn RO

    SetOutPath "$INSTDIR"

    ; Main executable
    File "TheBoysLauncher.exe"

    ; Create installation marker
    FileOpen $0 "$INSTDIR\.theboys-installed" w
    FileWrite $0 "installed_at=$\r$\n"
    FileWrite $0 "version=${APPVERSION}$\r$\n"
    FileWrite $0 "platform=windows$\r$\n"
    FileClose $0

    ; Create directories
    CreateDirectory "$INSTDIR\logs"
    CreateDirectory "$INSTDIR\config"

    ; Write registry keys
    WriteRegStr HKLM "Software\${APPID}" "InstallPath" "$INSTDIR"
    WriteRegStr HKLM "Software\${APPID}" "Version" "${APPVERSION}"
    WriteRegStr HKLM "Software\${APPID}" "DataDir" "$APPDATA\TheBoysLauncher"

    ; Write uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"

    ; Add to Add/Remove Programs
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "DisplayName" "${APPNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "DisplayIcon" "$INSTDIR\${APPEXE}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "Publisher" "TheBoys"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "DisplayVersion" "${APPVERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "HelpLink" "${APPURL}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "URLInfoAbout" "${APPURL}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "InstallLocation" "$INSTDIR"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "NoRepair" 1

    ; Calculate estimated size
    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}" "EstimatedSize" "$0"
SectionEnd

Section "Desktop Shortcut" SecDesktop
    CreateShortCut "$DESKTOP\${APPNAME}.lnk" "$INSTDIR\${APPEXE}" "" "$INSTDIR\${APPEXE}" 0
SectionEnd

Section "Start Menu Shortcuts" SecStartMenu
    CreateDirectory "$SMPROGRAMS\${APPNAME}"
    CreateShortCut "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk" "$INSTDIR\${APPEXE}" "" "$INSTDIR\${APPEXE}" 0
    CreateShortCut "$SMPROGRAMS\${APPNAME}\Uninstall.lnk" "$INSTDIR\Uninstall.exe" "" "$INSTDIR\Uninstall.exe" 0
    CreateShortCut "$SMPROGRAMS\${APPNAME}\Website.lnk" "${APPURL}" "" "${APPURL}" 0
SectionEnd

Section "File Associations" SecAssociations
    ; Associate .modpack files with TheBoys Launcher
    WriteRegStr HKCR ".modpack" "" "TheBoysModpack"
    WriteRegStr HKCR "TheBoysModpack" "" "TheBoys Modpack File"
    WriteRegStr HKCR "TheBoysModpack\DefaultIcon" "" "$INSTDIR\${APPEXE},0"
    WriteRegStr HKCR "TheBoysModpack\shell\open\command" "" '"$INSTDIR\${APPEXE}" "%1"'
SectionEnd

; Uninstaller section
Section "Uninstall"
    ; Remove files and directories
    Delete "$INSTDIR\Uninstall.exe"
    Delete "$INSTDIR\${APPEXE}"
    Delete "$INSTDIR\.theboys-installed"
    RMDir /r "$INSTDIR"

    ; Remove shortcuts
    Delete "$DESKTOP\${APPNAME}.lnk"
    Delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
    Delete "$SMPROGRAMS\${APPNAME}\Uninstall.lnk"
    Delete "$SMPROGRAMS\${APPNAME}\Website.lnk"
    RMDir "$SMPROGRAMS\${APPNAME}"

    ; Remove registry keys
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPID}"
    DeleteRegKey HKLM "Software\${APPID}"
    DeleteRegKey HKCR ".modpack"
    DeleteRegKey HKCR "TheBoysModpack"

    ; Remove user data (optional - uncomment if you want to clean user data)
    ; RMDir /r "$APPDATA\TheBoysLauncher"
SectionEnd

; Component descriptions
LangString DESC_SecCore ${LANG_ENGLISH} "Core application files and required components"
LangString DESC_SecDesktop ${LANG_ENGLISH} "Add a shortcut to your desktop for easy access"
LangString DESC_SecStartMenu ${LANG_ENGLISH} "Add shortcuts to your Start Menu"
LangString DESC_SecAssociations ${LANG_ENGLISH} "Open .modpack files automatically with TheBoys Launcher"

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
    !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} $(DESC_SecCore)
    !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} $(DESC_SecDesktop)
    !insertmacro MUI_DESCRIPTION_TEXT ${SecStartMenu} $(DESC_SecStartMenu)
    !insertmacro MUI_DESCRIPTION_TEXT ${SecAssociations} $(DESC_SecAssociations)
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; Functions
Function .onInit
    ; Check for previous installation
    ReadRegStr $R0 HKLM "Software\${APPID}" "InstallPath"
    StrCmp $R0 "" check_portable

    ; Found previous installation
    ReadRegStr $R1 HKLM "Software\${APPID}" "Version"

    MessageBox MB_YESNO|MB_ICONQUESTION \
        "${APPNAME} v$R1 is already installed on this system. $\n$\nDo you want to upgrade to the new version (${APPVERSION})? $\n$\nClick Yes to upgrade, No to install alongside." \
        IDYES upgrade_existing
    IDNO install_alongside

    upgrade_existing:
        ; Uninstall previous version
        ClearErrors
        ExecWait '$R0\Uninstall.exe _?=$R0'
        IfErrors upgrade_failed
        Goto done

    upgrade_failed:
        MessageBox MB_YESNO|MB_ICONEXCLAMATION \
            "Unable to automatically uninstall the previous version. $\n$\nDo you want to continue with the installation anyway?" \
            IDYES install_alongside
        Abort

    check_portable:
        ; Check for portable installation in current directory
        IfFileExists "$EXEDIR\TheBoysLauncher.exe" 0 check_user_dir
        IfFileExists "$EXEDIR\.theboys-installed" 0 check_user_dir

        ; Found portable installation
        MessageBox MB_YESNO|MB_ICONQUESTION \
            "A portable installation of ${APPNAME} was detected in the current directory. $\n$\nDo you want to migrate it to a proper installation? $\n$\nClick Yes to migrate (recommended) or No to keep both." \
            IDYES migrate_portable
        Goto install_alongside

    migrate_portable:
        ; Create migration script
        FileOpen $2 "$INSTDIR\migrate-portable.bat" w
        FileWrite $2 "@echo off$\r$\n"
        FileWrite $2 "echo Migrating portable installation...$\r$\n"
        FileWrite $2 "mkdir \"%APPDATA%\.theboys-launcher\" 2>nul$\r$\n"
        FileWrite $2 "if exist \"$EXEDIR\instances\" xcopy /E /I /Y \"$EXEDIR\instances\" \"%APPDATA%\.theboys-launcher\instances\"$\r$\n"
        FileWrite $2 "if exist \"$EXEDIR\config\" xcopy /E /I /Y \"$EXEDIR\config\" \"%APPDATA%\.theboys-launcher\config\"$\r$\n"
        FileWrite $2 "if exist \"$EXEDIR\prism\" xcopy /E /I /Y \"$EXEDIR\prism\" \"%APPDATA%\.theboys-launcher\prism\"$\r$\n"
        FileWrite $2 "if exist \"$EXEDIR\util\" xcopy /E /I /Y \"$EXEDIR\util\" \"%APPDATA%\.theboys-launcher\util\"$\r$\n"
        FileWrite $2 "echo Migration completed. Press any key to continue...$\r$\n"
        FileWrite $2 "pause >nul$\r$\n"
        FileWrite $2 "del \"%~f0\"$\r$\n"
        FileClose $2
        Goto done

    check_user_dir:
        ; Check for existing user directory
        IfFileExists "$APPDATA\.theboys-launcher" 0 done
        MessageBox MB_YESNO|MB_ICONQUESTION \
            "Existing user data was found in: $\n$APPDATA\.theboys-launcher$\n$\nDo you want to continue with the installation? $\n$\nYour existing instances and settings will be preserved." \
            IDNO abort_install
        Goto done

    abort_install:
        Abort

    install_alongside:
        ; Install alongside existing version
        StrCpy $INSTDIR "$PROGRAMFILES\${APPNAME} ${APPVERSION}"
        Goto done

    done:
FunctionEnd

Function un.onInit
    ; Check if application is running
    FindProcDLL::FindProc "${APPEXE}"
    Pop $R0
    IntCmp $R0 1 0 notRunning

    MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION \
        "${APPNAME} is currently running. $\n$\nClick OK to attempt to close it and continue with the uninstall." \
        IDOK tryClose
    Abort

    tryClose:
        FindProcDLL::KillProc "${APPEXE}"
        Sleep 2000
        FindProcDLL::FindProc "${APPEXE}"
        Pop $R0
        IntCmp $R0 1 notRunning 0 notRunning

        MessageBox MB_OK|MB_ICONEXCLAMATION \
            "Unable to close ${APPNAME}. $\n$\nPlease close it manually and run the uninstaller again."
        Abort

    notRunning:
FunctionEnd