// sfsEdgeStore 本地监控 Dashboard JS
let trendChart = null;
let realtimeData = [];
let historicalData = [];
let selectedDevice = '';
let selectedReading = '';
let isEnterprise = false;
let updateInterval = 5000; // 5秒刷新一次

// 初始化
document.addEventListener('DOMContentLoaded', function() {
    initChart();
    fetchLicenseInfo();
    fetchData();
    fetchMetrics();
    setInterval(fetchData, updateInterval);
    setInterval(fetchMetrics, updateInterval);
});

// 初始化图表
function initChart() {
    trendChart = echarts.init(document.getElementById('trendChart'));
    const option = {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'axis',
            backgroundColor: '#16213e',
            borderColor: '#0f3460',
            textStyle: { color: '#eee' }
        },
        legend: {
            data: ['数值'],
            textStyle: { color: '#888' }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            top: '15%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: [],
            axisLine: { lineStyle: { color: '#0f3460' } },
            axisLabel: { color: '#888' }
        },
        yAxis: {
            type: 'value',
            axisLine: { lineStyle: { color: '#0f3460' } },
            axisLabel: { color: '#888' },
            splitLine: { lineStyle: { color: '#0f3460' } }
        },
        series: [{
            name: '数值',
            type: 'line',
            smooth: true,
            data: [],
            lineStyle: { color: '#00d9ff', width: 2 },
            areaStyle: {
                color: {
                    type: 'linear',
                    x: 0, y: 0, x2: 0, y2: 1,
                    colorStops: [
                        { offset: 0, color: 'rgba(0, 217, 255, 0.3)' },
                        { offset: 1, color: 'rgba(0, 217, 255, 0)' }
                    ]
                }
            },
            itemStyle: { color: '#00d9ff' }
        }]
    };
    trendChart.setOption(option);
}

// 获取许可证信息
async function fetchLicenseInfo() {
    try {
        const res = await fetch('/api/license');
        const data = await res.json();
        isEnterprise = data.license_type === 'enterprise';
        document.getElementById('licenseBadge').textContent = isEnterprise ? '企业版' : '开源版';
        document.getElementById('licenseBadge').className = 'license-badge ' + (isEnterprise ? 'license-enterprise' : 'license-opensource');
    } catch (e) {
        console.error('获取许可证信息失败:', e);
    }
}

// 获取实时数据
async function fetchData() {
    try {
        const res = await fetch('/api/readings?limit=100');
        const data = await res.json();
        realtimeData = data.readings || [];
        updateTable();
        updateDeviceSelect();
        updateChart();
        document.getElementById('dataCount').textContent = realtimeData.length + ' 条记录';
        document.getElementById('connectionStatus').textContent = '已连接';
        document.getElementById('connectionDot').className = 'status-dot online';
    } catch (e) {
        console.error('获取数据失败:', e);
        document.getElementById('connectionStatus').textContent = '连接断开';
        document.getElementById('connectionDot').className = 'status-dot offline';
    }
}

// 获取系统指标
async function fetchMetrics() {
    try {
        const res = await fetch('/api/resources/status');
        const data = await res.json();
        if (data.status === 'success' && data.data) {
            updateMetrics(data.data);
        }
    } catch (e) {
        console.error('获取指标失败:', e);
    }

    // 获取 MQTT 状态
    try {
        const res = await fetch('/metrics');
        const data = await res.json();
        if (data.application) {
            document.getElementById('mqttReceived').textContent = formatNumber(data.application.MQTTMessagesReceived);
            document.getElementById('mqttProcessed').textContent = formatNumber(data.application.MQTTMessagesProcessed);
            document.getElementById('dbOperations').textContent = formatNumber(data.application.DatabaseOperations);
            document.getElementById('httpRequests').textContent = formatNumber(data.application.HTTPRequests);
            document.getElementById('errorCount').textContent = formatNumber(data.application.Errors);
        }
    } catch (e) {
        console.error('获取MQTT状态失败:', e);
    }

    document.getElementById('lastUpdate').textContent = '更新: ' + new Date().toLocaleTimeString();
}

// 更新表格
function updateTable() {
    const tbody = document.getElementById('tableBody');
    tbody.innerHTML = '';
    const searchText = document.getElementById('searchBox').value.toLowerCase();
    const filtered = realtimeData.filter(d =>
        (d.deviceName || '').toLowerCase().includes(searchText)
    );
    filtered.forEach(item => {
        const tr = document.createElement('tr');
        const time = new Date(item.timestamp / 1000000).toLocaleTimeString();
        tr.innerHTML = `
            <td>${time}</td>
            <td class="device-name">${formatDeviceName(item.deviceName)}</td>
            <td>${item.reading || '-'}</td>
            <td class="reading-value">${formatValue(item.value)}</td>
            <td>${item.valueType || '-'}</td>
        `;
        tbody.appendChild(tr);
    });
}

// 搜索过滤
function filterTable() {
    updateTable();
}

// 格式化设备名称（移除64字符填充）
function formatDeviceName(name) {
    if (!name) return '-';
    return name.trim().replace(/ +$/, '');
}

// 格式化数值
function formatValue(value) {
    if (value === null || value === undefined) return '-';
    if (typeof value === 'number') {
        return value.toFixed(2);
    }
    return String(value);
}

// 格式化数字
function formatNumber(num) {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return String(num);
}

// 更新设备选择器
function updateDeviceSelect() {
    const select = document.getElementById('deviceSelect');
    const devices = [...new Set(realtimeData.map(d => d.deviceName).filter(Boolean))];
    const currentValue = select.value;
    select.innerHTML = '<option value="">选择设备</option>';
    devices.forEach(dev => {
        select.innerHTML += `<option value="${dev}">${formatDeviceName(dev)}</option>`;
    });
    if (devices.includes(currentValue)) {
        select.value = currentValue;
    }
    selectedDevice = select.value;
}

// 更新图表
async function updateChart() {
    selectedDevice = document.getElementById('deviceSelect').value;
    selectedReading = document.getElementById('readingSelect').value;

    if (!selectedDevice) {
        trendChart.setOption({
            xAxis: { data: [] },
            series: [{ data: [] }]
        });
        return;
    }

    // 获取该设备的历史数据
    try {
        const res = await fetch(`/api/readings?deviceName=${encodeURIComponent(selectedDevice)}&limit=100`);
        const data = await res.json();
        historicalData = data.readings || [];
        updateReadingSelect();

        let chartData = historicalData;
        if (selectedReading) {
            chartData = historicalData.filter(d => d.reading === selectedReading);
        }

        const times = chartData.map(d => new Date(d.timestamp / 1000000).toLocaleTimeString());
        const values = chartData.map(d => parseFloat(d.value) || 0);

        trendChart.setOption({
            xAxis: { data: times },
            series: [{ data: values }]
        });
    } catch (e) {
        console.error('获取历史数据失败:', e);
    }
}

// 更新传感器选择器
function updateReadingSelect() {
    const select = document.getElementById('readingSelect');
    const readings = [...new Set(historicalData.filter(d => d.deviceName === selectedDevice).map(d => d.reading).filter(Boolean))];
    const currentValue = select.value;
    select.innerHTML = '<option value="">选择传感器</option>';
    readings.forEach(rd => {
        select.innerHTML += `<option value="${rd}">${rd}</option>`;
    });
    if (readings.includes(currentValue)) {
        select.value = currentValue;
    }
    selectedReading = select.value;
}

// 更新系统指标
function updateMetrics(data) {
    // CPU 使用率（这里用内存使用率模拟，实际应该从系统获取）
    const memUsage = data.memoryUsage || 0;
    const memPercent = Math.min((memUsage / 1024) * 100, 100); // 假设总共 1GB
    updateMetricBar('cpuBar', 'cpuValue', 0, '%', 'cpu');
    updateMetricBar('memoryBar', 'memoryValue', memPercent.toFixed(1), ' MB', 'memory');

    document.getElementById('goroutinesValue').textContent = data.goroutines || 0;
    document.getElementById('uptimeValue').textContent = formatUptime(data.uptime_seconds);
}

// 更新指标条
function updateMetricBar(barId, valueId, percent, unit, type) {
    const bar = document.getElementById(barId);
    const value = document.getElementById(valueId);
    const percentNum = parseFloat(percent);

    bar.style.width = percentNum + '%';
    value.textContent = percent + unit;

    let fillClass = 'metric-fill-normal';
    if (type === 'cpu' && percentNum > 80) fillClass = 'metric-fill-danger';
    else if (type === 'memory' && percentNum > 70) fillClass = 'metric-fill-warning';
    else if (type === 'memory' && percentNum > 85) fillClass = 'metric-fill-danger';

    bar.className = 'metric-fill ' + fillClass;
}

// 格式化运行时间
function formatUptime(seconds) {
    if (!seconds) return '0s';
    if (seconds < 60) return seconds + 's';
    if (seconds < 3600) return Math.floor(seconds / 60) + 'm';
    if (seconds < 86400) return Math.floor(seconds / 3600) + 'h';
    return Math.floor(seconds / 86400) + 'd';
}

// 刷新数据
function refreshData() {
    fetchData();
    fetchMetrics();
}

// 进入全屏
function enterFullscreen() {
    if (document.documentElement.requestFullscreen) {
        document.documentElement.requestFullscreen();
    }
}

// 窗口大小改变时重新调整图表
window.addEventListener('resize', function() {
    if (trendChart) {
        trendChart.resize();
    }
});