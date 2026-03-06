
#!/bin/bash

# sfsEdgeStore 负载测试脚本
# 使用方法：./load_test.sh

URL="http://localhost:8081"
CONCURRENCY=10
REQUESTS=100
TEST_TYPE="health"
DEVICE_NAME="TestDevice-001"
OUTPUT_JSON=""
USE_GO_TOOL=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --url)
            URL="$2"
            shift 2
            ;;
        -c|--concurrency)
            CONCURRENCY="$2"
            shift 2
            ;;
        -n|--requests)
            REQUESTS="$2"
            shift 2
            ;;
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -d|--device)
            DEVICE_NAME="$2"
            shift 2
            ;;
        -j|--json)
            OUTPUT_JSON="$2"
            shift 2
            ;;
        --use-go)
            USE_GO_TOOL=true
            shift
            ;;
        -h|--help)
            echo "sfsEdgeStore 负载测试工具"
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --url URL           服务器URL (默认: http://localhost:8081)"
            echo "  -c, --concurrency N  并发数 (默认: 10)"
            echo "  -n, --requests N   总请求数 (默认: 100)"
            echo "  -t, --type TYPE    测试类型: health, ready, query (默认: health)"
            echo "  -d, --device NAME  设备名称 (默认: TestDevice-001)"
            echo "  -j, --json FILE    输出JSON到文件"
            echo "  --use-go           使用Go负载测试工具"
            echo "  -h, --help         显示帮助"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 1
            ;;
    esac
done

echo "======================================"
echo "  sfsEdgeStore 负载测试工具"
echo "======================================"
echo ""

if [ "$USE_GO_TOOL" = true ]; then
    echo "使用 Go 负载测试工具..."
    GO_TOOL_PATH="./cmd/loadtest/loadtest"
    
    if [ ! -f "$GO_TOOL_PATH" ]; then
        echo "正在编译 Go 负载测试工具..."
        go build -o "$GO_TOOL_PATH" ./cmd/loadtest/main.go
        if [ $? -ne 0 ]; then
            echo "编译失败！"
            exit 1
        fi
    fi
    
    ARGS=("-url" "$URL" "-c" "$CONCURRENCY" "-n" "$REQUESTS" "-type" "$TEST_TYPE")
    if [ -n "$DEVICE_NAME" ]; then
        ARGS+=("-device" "$DEVICE_NAME")
    fi
    if [ -n "$OUTPUT_JSON" ]; then
        ARGS+=("-json" "$OUTPUT_JSON")
    fi
    
    "$GO_TOOL_PATH" "${ARGS[@]}"
    exit $?
fi

echo "使用 Bash 简单负载测试..."
echo "目标 URL: $URL"
echo "并发数: $CONCURRENCY"
echo "总请求数: $REQUESTS"
echo "测试类型: $TEST_TYPE"
echo ""

# 生成测试URL
get_test_url() {
    case "$TEST_TYPE" in
        health)
            echo "$URL/health"
            ;;
        ready)
            echo "$URL/ready"
            ;;
        query)
            echo "$URL/api/readings?deviceName=$DEVICE_NAME"
            ;;
    esac
}

TEST_URL=$(get_test_url)

# 初始化计数器
SUCCESS_COUNT=0
FAIL_COUNT=0
LATENCIES=()
declare -A ERRORS
START_TIME=$(date +%s.%N)

# 使用临时文件存储结果
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# 执行单个请求
run_request() {
    local id=$1
    local req_start=$(date +%s.%N)
    
    if command -v curl &gt;/dev/null 2&gt;&amp;1; then
        response=$(curl -s -w "%{http_code}" -o /dev/null --max-time 30 "$TEST_URL" 2&gt;&amp;1)
        http_code=$?
        if [ $http_code -eq 0 ]; then
            http_code=$(echo "$response" | tail -n1)
            if [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
                success=true
            else
                success=false
                error_msg="HTTP $http_code"
            fi
        else
            success=false
            error_msg="curl error: $http_code"
        fi
    elif command -v wget &gt;/dev/null 2&gt;&amp;1; then
        wget -q -O /dev/null --timeout=30 "$TEST_URL" 2&gt;"$TEMP_DIR/error_$id"
        if [ $? -eq 0 ]; then
            success=true
        else
            success=false
            error_msg=$(cat "$TEMP_DIR/error_$id")
        fi
    else
        success=false
        error_msg="No HTTP client (curl/wget) not found"
    fi
    
    local req_end=$(date +%s.%N)
    local latency=$(echo "$req_end - $req_start" | bc -l)
    local latency_ms=$(echo "$latency * 1000" | bc -l)
    
    echo "$success $latency_ms $error_msg" &gt; "$TEMP_DIR/result_$id"
}

export -f run_request
export TEST_URL TEMP_DIR

# 运行并发请求
echo "正在发送请求..."
for ((i=0; i&lt;REQUESTS; i++)); do
    while [ $(jobs -p | wc -l) -ge $CONCURRENCY ]; do
        sleep 0.1
    done
    run_request $i &amp;
done

wait

END_TIME=$(date +%s.%N)
TOTAL_DURATION=$(echo "$END_TIME - $START_TIME" | bc -l)

# 收集结果
for ((i=0; i&lt;REQUESTS; i++)); do
    if [ -f "$TEMP_DIR/result_$i" ]; then
        read success latency error_msg &lt; "$TEMP_DIR/result_$i"
        if [ "$success" = "true" ]; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            LATENCIES+=($latency)
        else
            FAIL_COUNT=$((FAIL_COUNT + 1))
            if [ -n "$error_msg" ]; then
                if [ -z "${ERRORS[$error_msg]}" ]; then
                    ERRORS[$error_msg]=1
                else
                    ERRORS[$error_msg]=$((ERRORS[$error_msg] + 1))
                fi
            fi
        fi
    fi
done

# 打印报告
echo ""
echo "======================================"
echo "  负载测试报告"
echo "======================================"
echo ""

TOTAL_REQUESTS=$((SUCCESS_COUNT + FAIL_COUNT))
echo "总请求数:        $TOTAL_REQUESTS"
echo "成功:            $SUCCESS_COUNT"
echo "失败:            $FAIL_COUNT"

if [ $TOTAL_REQUESTS -gt 0 ]; then
    SUCCESS_RATE=$(echo "scale=2; ($SUCCESS_COUNT / $TOTAL_REQUESTS) * 100" | bc -l)
    echo "成功率:          $SUCCESS_RATE%"
fi

echo "总耗时:          $(printf "%.3f" $TOTAL_DURATION) 秒"

if [ ${#LATENCIES[@]} -gt 0 ]; then
    # 计算统计数据
    sum=0
    min=${LATENCIES[0]}
    max=${LATENCIES[0]}
    for lat in "${LATENCIES[@]}"; do
        sum=$(echo "$sum + $lat" | bc -l)
        if (( $(echo "$lat &lt; $min" | bc -l) )); then
            min=$lat
        fi
        if (( $(echo "$lat &gt; $max" | bc -l) )); then
            max=$lat
        fi
    done
    avg=$(echo "scale=2; $sum / ${#LATENCIES[@]}" | bc -l)
    rps=$(echo "scale=2; $TOTAL_REQUESTS / $TOTAL_DURATION" | bc -l)
    
    echo "平均延迟:        $(printf "%.2f" $avg) ms"
    echo "最小延迟:        $(printf "%.2f" $min) ms"
    echo "最大延迟:        $(printf "%.2f" $max) ms"
    echo "请求/秒:         $(printf "%.2f" $rps)"
fi

if [ ${#ERRORS[@]} -gt 0 ]; then
    echo ""
    echo "错误统计:"
    for err in "${!ERRORS[@]}"; do
        echo "  - $err : ${ERRORS[$err]}"
    done
fi

echo ""
echo "======================================"

# 输出JSON
if [ -n "$OUTPUT_JSON" ]; then
    cat &gt; "$OUTPUT_JSON" &lt;&lt;EOF
{
  "total_requests": $TOTAL_REQUESTS,
  "success_requests": $SUCCESS_COUNT,
  "failed_requests": $FAIL_COUNT,
  "success_rate": ${SUCCESS_RATE:-0},
  "total_duration_ms": $(echo "$TOTAL_DURATION * 1000" | bc -l),
  "average_latency_ms": ${avg:-0},
  "min_latency_ms": ${min:-0},
  "max_latency_ms": ${max:-0},
  "requests_per_second": ${rps:-0},
  "error_counts": {
EOF
    first=true
    for err in "${!ERRORS[@]}"; do
        if [ "$first" = true ]; then
            first=false
        else
            echo "," &gt;&gt; "$OUTPUT_JSON"
        fi
        echo "    \"$err\": ${ERRORS[$err]}" &gt;&gt; "$OUTPUT_JSON"
    done
    cat &gt;&gt; "$OUTPUT_JSON" &lt;&lt;EOF
  }
}
EOF
    echo "结果已保存到: $OUTPUT_JSON"
fi
