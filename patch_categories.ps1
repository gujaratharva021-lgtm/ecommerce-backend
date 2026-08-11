$handlerPath = "internal\handlers\product_handler.go"
$lines = Get-Content $handlerPath

$newFunc = @(
    "func GetCategories(c *gin.Context) {",
    "`tcacheKey := `"categories:all`"",
    "`tvar categories []models.Category",
    "",
    "`tif found, _ := cache.Get(c.Request.Context(), cacheKey, &categories); found {",
    "`t`tc.JSON(http.StatusOK, gin.H{`"categories`": categories})",
    "`t`treturn",
    "`t}",
    "",
    "`tif err := database.DB.Order(`"name ASC`").Find(&categories).Error; err != nil {",
    "`t`tc.JSON(http.StatusInternalServerError, gin.H{`"error`": `"Failed to fetch categories`"})",
    "`t`treturn",
    "`t}",
    "",
    "`t_ = cache.Set(c.Request.Context(), cacheKey, categories, 30*time.Minute)",
    "`tc.JSON(http.StatusOK, gin.H{`"categories`": categories})",
    "}"
)

$before = $lines[0..113]
$after = $lines[123..($lines.Length - 1)]
$newLines = $before + $newFunc + $after

Set-Content -Path $handlerPath -Value $newLines -Encoding UTF8
