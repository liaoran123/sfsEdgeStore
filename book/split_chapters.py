#!/usr/bin/env python
# -*- coding: utf-8 -*-

import re
import os

source_file = r"d:\MyGo\src\sfsEdgeStore\book\EdgeX Foundry 与 sfsDb 结合：工业物联网边缘计算数据存储实战.md"
output_dir = r"d:\MyGo\src\sfsEdgeStore\book"

# 读取源文件
with open(source_file, 'r', encoding='utf-8') as f:
    lines = f.readlines()

# 定义章节边界
chapter_boundaries = [
    (4, 403, 596),
    (5, 597, 776),
    (6, 777, 1119),
    (7, 1120, 1376),
    (8, 1377, 1659),
    (9, 1660, 2085),
    (10, 2086, 2192),
    (11, 2193, len(lines))
]

chapter_titles = {
    4: "sfsEdgeStore快速开始",
    5: "与EdgeX-Foundry深度集成",
    6: "数据存储与查询",
    7: "监控与告警",
    8: "认证与安全",
    9: "备份与恢复",
    10: "生产部署最佳实践",
    11: "商业服务与支持",
    12: "成功案例"
}

# 处理每个章节
for chapter_num, start_line, end_line in chapter_boundaries:
    # 提取章节内容（注意：Python从0开始计数）
    chapter_content = lines[start_line-1:end_line]
    
    # 清理内容
    filtered_lines = []
    for line in chapter_content:
        # 跳过部分标题和分隔线
        if line.strip() == '---':
            continue
        if re.match(r'^## 第[一二三四]部分：', line):
            continue
        filtered_lines.append(line)
    
    # 确保有内容
    if not filtered_lines:
        continue
    
    # 生成文件名
    title = chapter_titles.get(chapter_num, f"第{chapter_num}章")
    filename = f"{chapter_num:02d}-第{chapter_num}章-{title}.md"
    filepath = os.path.join(output_dir, filename)
    
    # 写入文件
    with open(filepath, 'w', encoding='utf-8') as f:
        f.writelines(filtered_lines)
    
    print(f"Created: {filename}")

print("Done!")
