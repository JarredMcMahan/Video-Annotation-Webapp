cd backend
C:\programs\go\bin\go build
IF %ERRORLEVEL% NEQ 0 (
	exit %ERRORLEVEL%
)

C:\programs\go\bin\go test -coverprofile cp.out
IF %ERRORLEVEL% NEQ 0 (
	exit %ERRORLEVEL%
)

C:\programs\go\bin\go tool cover -html=cp.out
IF %ERRORLEVEL% NEQ 0 (
	exit %ERRORLEVEL%
)

cd ..\frontend
call npm install
IF %ERRORLEVEL% NEQ 0 (
  exit %ERRORLEVEL%
)

call npm run test:coverage
IF %ERRORLEVEL% NEQ 0 (
  exit %ERRORLEVEL%
)

call npm run lint
exit %ERRORLEVEL%
