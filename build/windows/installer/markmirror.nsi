; =====================================================================
;  MarkMirror - installer NSIS autonomo (dal modello di Zenith)
;
;  Non usa wails_tools.nsh (rigenerato da Wails a ogni build): tutto
;  il comportamento è in questo file e si personalizza liberamente.
;
;  Funzionalità:
;    - scelta della lingua all'avvio (italiano / inglese), ricordata
;      anche dal disinstallatore
;    - installazione per utente, senza diritti di amministratore
;      (%LOCALAPPDATA%\Programs\MarkMirror)
;    - componenti: programma, collegamento sul desktop, collegamento
;      nel menu Start
;    - voce in "App installate" / "Programmi e funzionalità"
;    - controllo del runtime WebView2
;    - rimozione (su conferma) della vecchia installazione per tutti gli
;      utenti fatta dal setup di Wails (fino alla v2.0.0, Program Files)
;    - disinstallatore, con opzione per rimuovere anche le impostazioni
;
;  Build (dalla root del progetto):
;    wails build
;    makensis build\windows\installer\markmirror.nsi
;  Versione da riga di comando:
;    makensis /DVERSION=2.1.0 build\windows\installer\markmirror.nsi
;  Eseguibile con un altro nome (es. wails build -o MarkMirror_win.exe):
;    makensis /DSOURCE_EXE=..\..\bin\MarkMirror_win.exe build\windows\installer\markmirror.nsi
;  Output: build\bin\MarkMirror-<versione>-amd64-installer.exe
;  (stesso nome di update.AssetNames in internal/update/update.go)
;
;  makensis lavora nella cartella di questo file: i percorsi relativi
;  qui sotto partono da build\windows\installer.
; =====================================================================

Unicode true
ManifestDPIAware true
SetCompressor /SOLID lzma

; ---------------------------------------------------------------------
;  Configurazione - personalizza qui
; ---------------------------------------------------------------------
!define APP_NAME     "MarkMirror"
!define APP_EXE      "MarkMirror.exe"
!define COMPANY      "Alessio Abrugiati"
!define APP_URL      "https://github.com/alexis-82/markmirror"
!ifndef VERSION
  !define VERSION    "1.0.0"          ; formato x.y.z
!endif

!ifndef SOURCE_EXE                    ; eseguibile prodotto da `wails build`
  !define SOURCE_EXE "..\..\bin\${APP_EXE}"   ; altro nome: /DSOURCE_EXE=..\..\bin\MarkMirror_win.exe
!endif
!define ICON_FILE    "..\icon.ico"

; Chiavi di registro (HKCU: installazione per utente)
!define APP_REG_KEY  "Software\${APP_NAME}"
!define UNINST_KEY   "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

; Installazione per tutti gli utenti del vecchio setup di Wails (HKLM,
; vista a 64 bit): "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
!define OLD_UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\Alessio AbrugiatiMarkMirror"

; Cartelle dati dell'app (i documenti dell'utente non si toccano mai)
!define SETTINGS_DIR "$APPDATA\MarkMirror"      ; settings.json di MarkMirror
!define WEBVIEW_DIR  "$APPDATA\${APP_EXE}"      ; cache WebView2 creata da Wails

; Runtime WebView2 (Evergreen)
!define WEBVIEW2_GUID "{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}"
!define WEBVIEW2_URL  "https://go.microsoft.com/fwlink/p/?LinkId=2124703"

; ---------------------------------------------------------------------
;  Impostazioni generali
; ---------------------------------------------------------------------
Name "${APP_NAME}"
OutFile "..\..\bin\${APP_NAME}-${VERSION}-amd64-installer.exe"
InstallDir "$LOCALAPPDATA\Programs\${APP_NAME}"
InstallDirRegKey HKCU "${APP_REG_KEY}" "InstallDir"
RequestExecutionLevel user
BrandingText "${APP_NAME} ${VERSION}"
ShowInstDetails show
ShowUninstDetails show

!include "MUI2.nsh"
!include "FileFunc.nsh"
!include "LogicLib.nsh"

; ---------------------------------------------------------------------
;  Interfaccia (Modern UI 2)
; ---------------------------------------------------------------------
!define MUI_ICON   "${ICON_FILE}"
!define MUI_UNICON "${ICON_FILE}"
!define MUI_ABORTWARNING
!define MUI_UNABORTWARNING

; Lingua scelta: salvata in HKCU, così il disinstallatore la riusa
!define MUI_LANGDLL_REGISTRY_ROOT      HKCU
!define MUI_LANGDLL_REGISTRY_KEY       "${APP_REG_KEY}"
!define MUI_LANGDLL_REGISTRY_VALUENAME "Language"
!define MUI_LANGDLL_ALLLANGUAGES       ; mostra sempre tutte le lingue

; Ultima pagina: casella "Avvia MarkMirror"
!define MUI_FINISHPAGE_RUN "$INSTDIR\${APP_EXE}"

; Pagine dell'installer
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Pagine del disinstallatore
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_COMPONENTS
!insertmacro MUI_UNPAGE_INSTFILES

; Lingue (la prima è quella predefinita)
!insertmacro MUI_LANGUAGE "Italian"
!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_RESERVEFILE_LANGDLL

; ---------------------------------------------------------------------
;  Testi tradotti
; ---------------------------------------------------------------------
LangString SecAppName      ${LANG_ITALIAN} "${APP_NAME} (obbligatorio)"
LangString SecAppName      ${LANG_ENGLISH} "${APP_NAME} (required)"
LangString SecAppDesc      ${LANG_ITALIAN} "Il programma ${APP_NAME}."
LangString SecAppDesc      ${LANG_ENGLISH} "The ${APP_NAME} application."
LangString SecDesktopName  ${LANG_ITALIAN} "Collegamento sul desktop"
LangString SecDesktopName  ${LANG_ENGLISH} "Desktop shortcut"
LangString SecDesktopDesc  ${LANG_ITALIAN} "Crea un collegamento a ${APP_NAME} sul desktop."
LangString SecDesktopDesc  ${LANG_ENGLISH} "Creates a ${APP_NAME} shortcut on the desktop."
LangString SecStartName    ${LANG_ITALIAN} "Collegamento nel menu Start"
LangString SecStartName    ${LANG_ENGLISH} "Start menu shortcut"
LangString SecStartDesc    ${LANG_ITALIAN} "Aggiunge ${APP_NAME} al menu Start."
LangString SecStartDesc    ${LANG_ENGLISH} "Adds ${APP_NAME} to the Start menu."

LangString UnSecAppName       ${LANG_ITALIAN} "${APP_NAME}"
LangString UnSecAppName       ${LANG_ENGLISH} "${APP_NAME}"
LangString UnSecAppDesc       ${LANG_ITALIAN} "Rimuove il programma, i collegamenti e la cache dell'interfaccia."
LangString UnSecAppDesc       ${LANG_ENGLISH} "Removes the application, its shortcuts and the UI cache."
LangString UnSecSettingsName  ${LANG_ITALIAN} "Impostazioni"
LangString UnSecSettingsName  ${LANG_ENGLISH} "Settings"
LangString UnSecSettingsDesc  ${LANG_ITALIAN} "Rimuove anche le impostazioni (tema, scroll sincronizzato, posizione della finestra, controllo aggiornamenti). I documenti Markdown non vengono mai eliminati."
LangString UnSecSettingsDesc  ${LANG_ENGLISH} "Also removes the settings (theme, synchronized scrolling, window position, update check). Markdown documents are never deleted."

LangString MsgWebView2 ${LANG_ITALIAN} "${APP_NAME} richiede Microsoft Edge WebView2 Runtime, che non risulta installato.$\n$\nVuoi aprire la pagina di download di Microsoft?"
LangString MsgWebView2 ${LANG_ENGLISH} "${APP_NAME} requires the Microsoft Edge WebView2 Runtime, which does not seem to be installed.$\n$\nDo you want to open the Microsoft download page?"

LangString MsgOldInstall ${LANG_ITALIAN} "È installata una versione precedente di ${APP_NAME} ($1) per tutti gli utenti in:$\n$2$\n$\nVuoi rimuoverla? Servono i diritti di amministratore. Le impostazioni vengono conservate.$\n$\nSe scegli No restano installate due copie di ${APP_NAME}."
LangString MsgOldInstall ${LANG_ENGLISH} "An older version of ${APP_NAME} ($1) is installed for all users in:$\n$2$\n$\nDo you want to remove it? Administrator rights are required. Your settings are kept.$\n$\nIf you choose No, two copies of ${APP_NAME} stay installed."
LangString MsgOldRemoving ${LANG_ITALIAN} "Rimozione della versione precedente in corso..."
LangString MsgOldRemoving ${LANG_ENGLISH} "Removing the older version..."
LangString MsgOldNotRemoved ${LANG_ITALIAN} "La versione precedente non è stata rimossa. Puoi disinstallarla in seguito da Impostazioni > App installate."
LangString MsgOldNotRemoved ${LANG_ENGLISH} "The older version was not removed. You can uninstall it later from Settings > Installed apps."

; Proprietà dell'installer (tasto destro > Proprietà > Dettagli)
VIProductVersion "${VERSION}.0"
VIFileVersion    "${VERSION}.0"
VIAddVersionKey /LANG=${LANG_ITALIAN} "ProductName"     "${APP_NAME}"
VIAddVersionKey /LANG=${LANG_ITALIAN} "CompanyName"     "${COMPANY}"
VIAddVersionKey /LANG=${LANG_ITALIAN} "FileDescription" "${APP_NAME} Setup"
VIAddVersionKey /LANG=${LANG_ITALIAN} "FileVersion"     "${VERSION}"
VIAddVersionKey /LANG=${LANG_ITALIAN} "ProductVersion"  "${VERSION}"
VIAddVersionKey /LANG=${LANG_ITALIAN} "LegalCopyright"  "${COMPANY}"

; ---------------------------------------------------------------------
;  Installazione
; ---------------------------------------------------------------------
Section "!$(SecAppName)" SecApp
  SectionIn RO
  SetShellVarContext current

  Call RemoveOldInstall

  SetOutPath "$INSTDIR"

  File "/oname=${APP_EXE}" "${SOURCE_EXE}"
  WriteUninstaller "$INSTDIR\uninstall.exe"

  WriteRegStr HKCU "${APP_REG_KEY}" "InstallDir" "$INSTDIR"

  ; Voce in "App installate" / "Programmi e funzionalità"
  WriteRegStr   HKCU "${UNINST_KEY}" "DisplayName"          "${APP_NAME}"
  WriteRegStr   HKCU "${UNINST_KEY}" "DisplayVersion"       "${VERSION}"
  WriteRegStr   HKCU "${UNINST_KEY}" "DisplayIcon"          "$INSTDIR\${APP_EXE}"
  WriteRegStr   HKCU "${UNINST_KEY}" "Publisher"            "${COMPANY}"
  WriteRegStr   HKCU "${UNINST_KEY}" "URLInfoAbout"         "${APP_URL}"
  WriteRegStr   HKCU "${UNINST_KEY}" "InstallLocation"      "$INSTDIR"
  WriteRegStr   HKCU "${UNINST_KEY}" "UninstallString"      "$\"$INSTDIR\uninstall.exe$\""
  WriteRegStr   HKCU "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
  WriteRegDWORD HKCU "${UNINST_KEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNINST_KEY}" "NoRepair" 1
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD HKCU "${UNINST_KEY}" "EstimatedSize" $0

  Call CheckWebView2
SectionEnd

Section "$(SecDesktopName)" SecDesktop
  SetShellVarContext current
  CreateShortcut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
SectionEnd

Section "$(SecStartName)" SecStart
  SetShellVarContext current
  CreateShortcut "$SMPROGRAMS\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
SectionEnd

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SecApp}     $(SecAppDesc)
  !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} $(SecDesktopDesc)
  !insertmacro MUI_DESCRIPTION_TEXT ${SecStart}   $(SecStartDesc)
!insertmacro MUI_FUNCTION_DESCRIPTION_END

Function .onInit
  !insertmacro MUI_LANGDLL_DISPLAY
FunctionEnd

; Fino alla v2.0.0 MarkMirror si installava con il setup di Wails, per
; tutti gli utenti (Program Files, HKLM). Se c'è, su conferma si lancia il
; suo disinstallatore (chiede l'elevazione) e si aspetta che la voce di
; registro sparisca: il disinstallatore NSIS si copia in una cartella
; temporanea e termina subito, quindi non basta attendere il processo.
Function RemoveOldInstall
  SetRegView 64
  ReadRegStr $0 HKLM "${OLD_UNINST_KEY}" "UninstallString"
  ReadRegStr $1 HKLM "${OLD_UNINST_KEY}" "DisplayVersion"
  ${If} $0 != ""
    ; Il setup di Wails scrive il percorso tra virgolette: si tolgono solo
    ; se ci sono davvero.
    StrCpy $3 $0
    StrCpy $5 $3 1
    ${If} $5 == '"'
      StrCpy $3 $3 "" 1
      StrCpy $5 $3 1 -1
      ${IfThen} $5 == '"' ${|} StrCpy $3 $3 -1 ${|}
    ${EndIf}
    ${GetParent} $3 $2
    MessageBox MB_YESNO|MB_ICONQUESTION "$(MsgOldInstall)" /SD IDNO IDNO old_done
    DetailPrint "$(MsgOldRemoving)"
    ClearErrors
    ExecShellWait "open" "$3" "/S"
    ; Errore = disinstallatore non avviato (es. UAC rifiutato): niente attesa
    ${IfNot} ${Errors}
      StrCpy $4 0
      ${Do}
        Sleep 500
        ReadRegStr $0 HKLM "${OLD_UNINST_KEY}" "UninstallString"
        IntOp $4 $4 + 1
        ${IfThen} $0 == "" ${|} ${ExitDo} ${|}
      ${LoopUntil} $4 >= 240          ; al massimo 2 minuti
    ${EndIf}
    ${If} $0 != ""
      MessageBox MB_OK|MB_ICONEXCLAMATION "$(MsgOldNotRemoved)" /SD IDOK
    ${EndIf}
  ${EndIf}
  old_done:
  SetRegView default
FunctionEnd

; WebView2 is preinstalled on up-to-date Windows 10/11; if missing, offer
; the official download page (the installer itself stays offline).
; NSIS is 32-bit, so HKLM\SOFTWARE is read through WOW6432Node.
Function CheckWebView2
  ReadRegStr $0 HKLM "SOFTWARE\Microsoft\EdgeUpdate\Clients\${WEBVIEW2_GUID}" "pv"
  ${If} $0 == ""
  ${OrIf} $0 == "0.0.0.0"
    ReadRegStr $0 HKCU "Software\Microsoft\EdgeUpdate\Clients\${WEBVIEW2_GUID}" "pv"
  ${EndIf}
  ${If} $0 == ""
  ${OrIf} $0 == "0.0.0.0"
    MessageBox MB_YESNO|MB_ICONEXCLAMATION "$(MsgWebView2)" /SD IDNO IDNO +2
      ExecShell "open" "${WEBVIEW2_URL}"
  ${EndIf}
FunctionEnd

; ---------------------------------------------------------------------
;  Disinstallazione
; ---------------------------------------------------------------------
Section "un.$(UnSecAppName)" UnSecApp
  SectionIn RO
  SetShellVarContext current

  Delete "$DESKTOP\${APP_NAME}.lnk"
  Delete "$SMPROGRAMS\${APP_NAME}.lnk"

  Delete "$INSTDIR\${APP_EXE}"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"                      ; solo se vuota: non cancella file estranei

  RMDir /r "${WEBVIEW_DIR}"

  DeleteRegKey HKCU "${UNINST_KEY}"
  DeleteRegKey HKCU "${APP_REG_KEY}"
SectionEnd

Section /o "un.$(UnSecSettingsName)" UnSecSettings
  RMDir /r "${SETTINGS_DIR}"
SectionEnd

!insertmacro MUI_UNFUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${UnSecApp}      $(UnSecAppDesc)
  !insertmacro MUI_DESCRIPTION_TEXT ${UnSecSettings} $(UnSecSettingsDesc)
!insertmacro MUI_UNFUNCTION_DESCRIPTION_END

Function un.onInit
  !insertmacro MUI_UNGETLANGUAGE
FunctionEnd
