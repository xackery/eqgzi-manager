set last=copymap
copy map\*.map %EQSERVERPATH%\base || goto :error

set last=copywater
copy map\*.wtr %EQSERVERPATH%\water || goto :error
goto :EOF

:error
echo failed during %last% with signal #%errorlevel%
exit /b %errorlevel%