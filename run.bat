
rem Bypass "Terminate Batch Job" prompt.
if "%~1"=="-FIXED_CTRL_C" (
   REM Remove the -FIXED_CTRL_C parameter
   SHIFT
) ELSE (
   REM Run the batch with <NUL and -FIXED_CTRL_C
   CALL <NUL %0 -FIXED_CTRL_C %*
   GOTO :EOF
)

mkdir bin || exit /b
cd bin && del quail-gui.exe && cd .. || exit /b
rsrc -ico quail-gui.ico -manifest quail-gui.exe.manifest || exit /b
copy /y quail-gui.exe.manifest bin\quail-gui.exe.manifest || exit /b
go build -buildmode=pie -ldflags="-s -w" -o quail-gui.exe main.go || exit /b
move quail-gui.exe bin/quail-gui.exe
cd bin && quail-mod-manager.exe || exit /b
go build -o bin || exit /b
cd bin && quail-mod-manager.exe || exit /b
