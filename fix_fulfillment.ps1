$handlerPath = "internal\handlers\warehouse_fulfillment_handler.go"
$content = Get-Content $handlerPath -Raw

$oldImports = "`t`"net/http`"`r`n`t`"time`"`r`n`r`n`t`"github.com/gin-gonic/gin`""
$newImports = "`t`"net/http`"`r`n`t`"strconv`"`r`n`r`n`t`"github.com/gin-gonic/gin`""
$content = $content.Replace($oldImports, $newImports)

$oldFunc = "func parsePositiveInt(s string) (int, error) {`r`n`tvar v int`r`n`t_, err := fmtSscanf(s, &v)`r`n`tif err != nil || v < 1 {`r`n`t`treturn 0, gorm.ErrInvalidData`r`n`t}`r`n`treturn v, nil`r`n}"
$newFunc = "func parsePositiveInt(s string) (int, error) {`r`n`tv, err := strconv.Atoi(s)`r`n`tif err != nil || v < 1 {`r`n`t`treturn 0, gorm.ErrInvalidData`r`n`t}`r`n`treturn v, nil`r`n}"
$content = $content.Replace($oldFunc, $newFunc)

Set-Content -Path $handlerPath -Value $content -Encoding UTF8 -NoNewline
