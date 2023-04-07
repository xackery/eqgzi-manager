@echo off
set last=copy

copy out\* %EQPATH% || goto :error
goto :EOF

:error
echo failed during %last% with signal #%errorlevel%
exit /b %errorlevel%