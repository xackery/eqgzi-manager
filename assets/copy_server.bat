@echo off
set last=copy

copy out\* %EQSERVERPATH% || goto :error
goto :EOF

:error
echo failed during %last% with signal #%errorlevel%
exit /b %errorlevel%