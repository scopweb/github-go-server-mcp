#!/bin/bash

# Script de construcción y configuración del MCP GitHub

echo "Configurando GitHub Go Server MCP..."

# Verificar Go
if ! command -v go &> /dev/null; then
    echo "Go no está instalado. Por favor instale Go 1.23 o superior."
    exit 1
fi

# Verificar variables de entorno
if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN no está configurado."
    echo "Configúrelo con: export GITHUB_TOKEN='your_token'"
fi

# Instalar dependencias
echo "Instalando dependencias..."
go mod tidy

# Compilar servidor
echo "Compilando servidor MCP..."
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    go build -o github-mcp.exe main.go
    echo "Compilado: github-mcp.exe"
else
    go build -o github-mcp main.go
    echo "Compilado: github-mcp"
fi

# Validar sintaxis
echo "Validando sintaxis..."
go vet ./...

echo ""
echo "MCP GitHub configurado correctamente!"
echo ""
echo "Próximos pasos:"
echo "1. Configurar variable de entorno GITHUB_TOKEN"
echo "2. Copiar claude_desktop_config.json a la configuración de Claude Desktop"
echo "3. Reiniciar Claude Desktop"
echo ""
echo "Para probar manualmente:"
echo "   ./github-mcp (Linux/Mac) o github-mcp.exe (Windows)"