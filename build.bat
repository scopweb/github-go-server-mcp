@echo off
REM Script de construcción para Windows

echo Configurando GitHub MCP Server...

REM Verificar Go
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Go no está instalado. Por favor instale Go 1.23 o superior.
    pause
    exit /b 1
)

REM Verificar variables de entorno
if "%GITHUB_TOKEN%"=="" (
    echo GITHUB_TOKEN no está configurado.
    echo Configúrelo con: set GITHUB_TOKEN=your_token
)

REM Instalar dependencias
echo Instalando dependencias...
go mod tidy

REM Compilar servidor
echo Compilando servidor MCP...
go build -o github-mcp.exe main.go
if %errorlevel% neq 0 (
    echo Error al compilar
    pause
    exit /b 1
)

echo Compilado: github-mcp.exe

REM Validar sintaxis
echo Validando sintaxis...
go vet ./...

echo.
echo GitHub MCP Server configurado correctamente!
echo.
echo Próximos pasos:
echo 1. Configurar variable de entorno GITHUB_TOKEN
echo 2. Copiar claude_desktop_config.json a la configuración de Claude Desktop
echo 3. Reiniciar Claude Desktop
echo.
echo Para probar manualmente:
echo    github-mcp.exe
echo.
pause