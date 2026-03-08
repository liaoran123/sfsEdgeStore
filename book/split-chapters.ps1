# 电子书章节拆分脚本
$sourceFile = "d:\MyGo\src\sfsEdgeStore\book\EdgeX Foundry 与 sfsDb 结合：工业物联网边缘计算数据存储实战.md"
$outputDir = "d:\MyGo\src\sfsEdgeStore\book"

# 读取源文件
$content = Get-Content -Path $sourceFile -Raw -Encoding UTF8

# 定义章节信息
$chapters = @(
    @{ Num = 4; Title = "sfsEdgeStore 快速开始"; StartLine = 263; EndLine = 402 },
    @{ Num = 5; Title = "与 EdgeX Foundry 深度集成"; StartLine = 403; EndLine = 596 },
    @{ Num = 6; Title = "数据存储与查询"; StartLine = 597; EndLine = 776 },
    @{ Num = 7; Title = "监控与告警"; StartLine = 777; EndLine = 1119 },
    @{ Num = 8; Title = "认证与安全"; StartLine = 1120; EndLine = 1376 },
    @{ Num = 9; Title = "备份与恢复"; StartLine = 1377; EndLine = 1659 },
    @{ Num = 10; Title = "生产部署最佳实践"; StartLine = 1660; EndLine = 2085 },
    @{ Num = 11; Title = "商业服务与支持"; StartLine = 2086; EndLine = 2192 },
    @{ Num = 12; Title = "成功案例"; StartLine = 2193; EndLine = 9999 }
)

# 处理每个章节
foreach ($chapter in $chapters) {
    $fileName = "$outputDir\{0:D2}-第{1}章-{2}.md" -f $chapter.Num, $chapter.Num, ($chapter.Title -replace '[^a-zA-Z0-9\u4e00-\u9fff-]', '')
    
    Write-Host "Processing Chapter $($chapter.Num): $($chapter.Title) -> $fileName"
    
    # 读取指定行范围
    $lines = Get-Content -Path $sourceFile -Encoding UTF8
    $chapterLines = $lines[($chapter.StartLine - 1)..($chapter.EndLine - 1)]
    
    # 过滤掉分隔线（---）和部分标题
    $filteredLines = @()
    foreach ($line in $chapterLines) {
        if ($line -match '^## 第[0-9]+章：' -or $line -match '^### [0-9]+\.' -or $line -match '^#### ') {
            $filteredLines += $line
        } elseif ($line -ne '---' -and $line -notmatch '^## 第[一二三四]部分：') {
            $filteredLines += $line
        }
    }
    
    # 确保文件内容非空
    if ($filteredLines.Count -gt 0) {
        $filteredLines | Out-File -FilePath $fileName -Encoding UTF8 -NoNewline
    }
}

Write-Host "Chapter splitting completed!"
