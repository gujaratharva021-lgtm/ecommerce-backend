$handlerPath = "internal\handlers\admin_handler.go"
$lines = Get-Content $handlerPath

$inserts = @{
    266 = "`tif err := cache.DeleteByPrefix(c.Request.Context(), `"products:list:`"); err != nil { }`n`t_ = cache.Delete(c.Request.Context(), `"products:id:`"+id)"
    222 = "`t_ = cache.DeleteByPrefix(c.Request.Context(), `"products:list:`")`n`t_ = cache.Delete(c.Request.Context(), `"products:id:`"+id)"
    191 = "`t_ = cache.DeleteByPrefix(c.Request.Context(), `"products:list:`")`n`t_ = cache.Delete(c.Request.Context(), `"products:id:`"+id)"
    148 = "`t_ = cache.DeleteByPrefix(c.Request.Context(), `"products:list:`")"
    95  = "`t_ = cache.Delete(c.Request.Context(), `"categories:all`")"
    63  = "`t_ = cache.Delete(c.Request.Context(), `"categories:all`")"
    33  = "`t_ = cache.Delete(c.Request.Context(), `"categories:all`")"
}

$newLines = New-Object System.Collections.Generic.List[string]
for ($i = 0; $i -lt $lines.Length; $i++) {
    $lineNum = $i + 1
    $newLines.Add($lines[$i])
    if ($inserts.ContainsKey($lineNum)) {
        $inserts[$lineNum] -split "`n" | ForEach-Object { $newLines.Add($_) }
    }
}

Set-Content -Path $handlerPath -Value $newLines -Encoding UTF8
