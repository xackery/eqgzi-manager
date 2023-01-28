@echo off
set last=blender

:: change soldungb to your zone name
set zone=%ZONE%

:: change C:\src\eqgzi\out\convert.py to your eqgzi's path with the file
blender --background %zone%.blend --python %EQGZI%\convert.py || goto :error

set last=eqgzi
eqgzi import %zone% || goto :error

set last=azone
cd out 
%EQGZI%\azone.exe %zone% || goto :error 
cd ..
del out\azone.log
rmdir /s /q map\
mkdir map\
move out\%zone%.map map\

set last=awater
cd out 
%EQGZI%\awater.exe %zone% || goto :error
cd ..
del out\awater.log
move out\%zone%.wtr map\

goto :EOF

:error
echo failed during %last% with signal #%errorlevel%
exit /b %errorlevel%