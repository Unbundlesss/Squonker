@echo off
:start
ECHO    WELCOME TO SQUONKER. YOUR OPTIONS ARE AS FOLLOWS:
ECHO    1. discover an iOS device
ECHO    2. squonk a pack
ECHO    3. deploy a pack to iOS
ECHO    4. remove a pack from iOS
ECHO    5. clone a soundpack repository
ECHO    6. pull a soundpack repository
ECHO    7. OBLITERATE
ECHO    8. nope out of this program

set /p menu="pick a number:      "
IF '%menu%'=='1' GOTO ios-discover
IF '%menu%'=='2' GOTO squonk
IF '%menu%'=='3' GOTO ios-deploy
IF '%menu%'=='4' GOTO ios-remove
IF '%menu%'=='5' GOTO clone
IF '%menu%'=='6' GOTO pull
IF '%menu%'=='7' GOTO obliterate 
IF '%menu%'=='8' GOTO nope
IF '%menu%'=='9' GOTO well-done
ECHO "I SAID pick a NUMBER"
ECHO.
PAUSE
GOTO start

:ios-discover
echo    You should only need to run this once to teach Squonker how to break into your iOS 
echo    device and wangjangle all your files. 
echo.
echo    You'll need the IP address of your iOS device, and then the port number of the 
echo    WebDAV server on your iOS device (which you need to get from Filza - if you haven't 
echo    changed anything there, it's probably 11111). 
echo.


set /P ip="-------What is your iOS device's IP address?   "
set /P port="-------What is the port number of the WebDAV server on your device?   "
echo        thanks that should do it
.\squonker.exe ios-discover --webdav http://%ip%:%port%
PAUSE
GOTO start


:squonk
set /p victim="Give the folder name of the pack you want squonked    ->    "
set /p pack="Give the name of the specific pack you want squonked or press Enter to squonk the whole folder  ->    "
set /p ios-agree="type 'i' to deploy the pack directly to your ios device as it is squonked, or press Enter to skip this     "

IF '%pack%'=='' IF '%ios-agree%'=='i' GOTO squonkios

IF '%pack%'=='' IF NOT '%ios-agree%'=='i' GOTO squonkbasic

IF NOT '%pack%'=='' IF '%ios-agree%'=='i' GOTO squonkspecios

IF NOT '%pack%'=='' IF NOT '%ios-agree%'=='i' GOTO squonkspec


:squonkbasic
ECHO         COMMENCING SQUONK
.\squonker.exe compile -d %victim%
ECHO         SQUONK COMPLETE.
GOTO squonkagain

:squonkios
ECHO         COMMENCING SQUONK
.\squonker.exe compile -d %victim% --ios
ECHO         SQUONK COMPLETE.
GOTO squonkagain

:squonkspec
ECHO         COMMENCING SQUONK
.\squonker.exe compile -d %victim% -p %pack%
ECHO         SQUONK COMPLETE.
GOTO squonkagain

:squonkspecios
ECHO         COMMENCING SQUONK
.\squonker.exe compile -d %victim% -p %pack% --ios
ECHO         SQUONK COMPLETE.
GOTO squonkagain

:squonkagain
set /p repeat="Squonk again? [y/n]"
IF NOT '%repeat%'=='' SET repeat=%repeat:~0,1%
IF '%repeat%'=='y' GOTO squonk
GOTO start



:ios-deploy
set /p pack="Which pack do you want to deploy to your iOS device?  ->   "
set /p scrub="Would you like to clean up the files in the folder as you go? [y/n]  ->   "
IF '%scrub%'=='y' GOTO clean
GOTO unclean


:clean
echo         COMMENCING DEPLOYMENT
.\squonker.exe ios-deploy -p %pack% --clean
echo         DEPLOYMENT COMPLETE.
GOTO deployagain

:unclean
echo         COMMENCING DEPLOYMENT
.\squonker.exe ios-deploy -p %pack%
echo         DEPLOYMENT COMPLETE.
GOTO deployagain

:deployagain
set /p repeat="Deploy another? [y/n]"
IF NOT '%repeat%'=='' SET repeat=%repeat:~0,1%
IF '%repeat%'=='y' GOTO ios-deploy
GOTO start


:ios-remove

ECHO this will remove whichever pack you name from your iOS device.
set /p agree="Are you sure you want to do that? [y/n]"
IF NOT '%agree%'=='' SET agree=%agree:~0,1%
IF '%agree%'== 'y' GOTO rem
GOTO start

:rem
set /p destroy="give the name of the pack you would like to remove from your iOS device   ->   "
echo COMMENCING REMOVAL
.\squonker.exe ios-remove -p %destroy%
echo REMOVAL COMPLETE
set /p remagain="Would you like to remove another pack? [y/n]   "
IF NOT '%remagain%'=='' SET remagain=%remagain:~0,1%
IF '%remagain%'== 'y' GOTO rem
GOTO start

:clone

set /p address="What is the address of the repository you want to clone?  ->    "
set /p name="What would you like to name the local folder?  ->   "
set /p key="If you have already cloned any repositories, give the name of the repository which you want to reuse the key for, otherwise press Enter to continue without this  ->   "

IF NOT '%key%'=='' GOTO samekey
GOTO first

:first
ECHO Make sure to wait until the cloning has actually finished before exiting the program, otherwise things will not go well for you
.\squonker.exe sync-clone -g %address% -n %name% --key
GOTO :cloneagain

:samekey
ECHO Make sure to wait until the cloning has actually finished before exiting the program, otherwise things will not go well for you
.\squonker.exe sync-clone -g %address% -n %name% --key --samekey %key%
GOTO :cloneagain

:cloneagain
set /p cloneagain="Would you like to clone another repository? [y/n]   "
IF NOT '%cloneagain%'=='' SET cloneagain=%cloneagain:~0,1%
IF '%cloneagain%'=='y' GOTO clone
GOTO start

:pull
set /p pull="Would you like to pull all your cloned Squad repositories? [y/n]    "
IF NOT '%pull%'=='' SET pull=%pull:~0,1%
IF '%pull%'== 'y' GOTO pullyes
GOTO start

:pullyes
.\squonker.exe sync-pull
PAUSE
GOTO start 

:obliterate
ECHO ok what this one does is it DELETES ALL THE INSTRUMENTS FROM YOUR iOS DEVICE. It's less scary than it sounds.
ECHO but seriously, are you totally sure you want to do this?
set /p murder="[y/n]:     "
IF NOT '%murder%'=='' SET murder=%murder:~0,1%
IF '%murder%'== 'y' GOTO aaaaaa
GOTO start

:aaaaaa
.\squonker ios-obliterate
ECHO THE INSTRUMENTS HAVE BEEN OBLITERATED.
PAUSE
GOTO start

:well-done
ECHO you have found the secret easter egg in this shitty batch file, well done, I hope you're happy with yourself
PAUSE
GOTO start

:nope
echo      have a fabulous day!
PAUSE
EXIT