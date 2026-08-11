$handlerPath = "internal\handlers\product_handler.go"
$lines = Get-Content $handlerPath

$newFunc = @(
    "func GetProductByID(c *gin.Context) {",
    "`tid := c.Param(`"id`")",
    "`tcacheKey := `"products:id:`" + id",
    "",
    "`tvar product models.Product",
    "`tif found, _ := cache.Get(c.Request.Context(), cacheKey, &product); found {",
    "`t`tc.JSON(http.StatusOK, product)",
    "`t`treturn",
    "`t}",
    "",
    "`tif err := database.DB.Preload(`"Category`").Preload(`"Inventories`").First(&product, id).Error; err != nil {",
    "`t`tc.JSON(http.StatusNotFound, gin.H{`"error`": `"Product not found`"})",
    "`t`treturn",
    "`t}",
    "",
    "`t_ = cache.Set(c.Request.Context(), cacheKey, product, 10*time.Minute)",
    "`tc.JSON(http.StatusOK, product)",
    "}"
)

$before = $lines[0..101]
$after = $lines[113..($lines.Length - 1)]
$newLines = $before + $newFunc + $after

Set-Content -Path $handlerPath -Value $newLines -Encoding UTF8
