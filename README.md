# GitHub MCP Server

Go-based MCP server that connects GitHub to Claude Desktop, enabling direct repository operations from Claude's interface.

## Permisos Necesarios del Token

Para que todas las funciones trabajen correctamente, tu **GitHub Personal Access Token** debe tener estos permisos:

### Mínimos Requeridos:
```
repo (Full control of private repositories)
  - Necesario para crear repos, issues, PRs
  - Permite lectura/escritura en repositorios
```

### Opcionales (para funcionalidad completa):
```
delete_repo (Delete repositories) - Solo si necesitas borrar repos
workflow (Update GitHub Action workflows) - Para trabajar con Actions
admin:repo_hook (Repository hooks) - Para webhooks
```

### Generar Token:
1. Ve a: [GitHub Settings → Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Selecciona los scopes necesarios arriba
4. Copia el token generado

## Instalación

```bash
# Instalar dependencias
go mod tidy

# Compilar
go build -o github-mcp.exe main.go
```

## Configuración Claude Desktop

Añade esto a tu `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "github-mcp": {
      "command": "C:\\MCPs\\clone\\github-go-server-mcp\\github-mcp.exe",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "tu_token_aqui_con_permisos_repo"
      }
    }
  }
}
```

## Herramientas Disponibles (Todas Testeadas)

| Función | Estado | Descripción |
|---------|---------|-------------|
| **github_list_repos** | Testeado | Lista repositorios del usuario |
| **github_create_repo** | Testeado | Crea nuevo repositorio |
| **github_get_repo** | Testeado | Obtiene información de repositorio |
| **github_list_branches** | Testeado | Lista ramas de un repositorio |
| **github_list_prs** | Testeado | Lista pull requests |
| **github_create_pr** | Testeado | Crea nuevo pull request |
| **github_list_issues** | Testeado | Lista issues de un repositorio |
| **github_create_issue** | Testeado | Crea nuevo issue |

## Uso

1. **Compilar el servidor**
2. **Generar token GitHub con permisos `repo`**
3. **Actualizar `GITHUB_TOKEN` en configuración**
4. **Añadir configuración a Claude Desktop**
5. **Reiniciar Claude Desktop**

## Solución de Problemas

### Error 403 "Resource not accessible by personal access token"
- Tu token no tiene permisos suficientes
- Genera nuevo token con scope `repo`
- Reinicia Claude Desktop después del cambio

### Error "null" en respuestas
- Normal para repos vacíos o sin PRs/issues
- El MCP funciona correctamente

## Estado del Proyecto

- **Funciones de lectura**: Completamente operativas
- **Funciones de escritura**: Completamente operativas  
- **Gestión de permisos**: Documentada y verificada
- **Testing completo**: Todas las funciones probadas
- **Listo para producción**: Stable release