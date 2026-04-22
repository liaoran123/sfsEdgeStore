// sfsEdgeStore 本地监控 Dashboard JS
let trendChart = null;
let realtimeData = [];
let historicalData = [];
let selectedDevice = '';
let selectedReading = '';
let updateInterval = 5000; // 5秒刷新一次

// 初始化
document.addEventListener('DOMContentLoaded', function() {
    initChart();
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

// 获取实时数据
async function fetchData() {
    try {
        const res = await fetch('/api/readings?limit=100');
        const data = await res.json();
        realtimeData = data.readings || [];
        updateTable();
        updateDeviceList();
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
        const res = await fetch('/metrics');
        const data = await res.json();
        if (data.application) {
            document.getElementById('mqttReceived').textContent = formatNumber(data.application.mqtt_messages_received);
            document.getElementById('mqttProcessed').textContent = formatNumber(data.application.mqtt_messages_processed);
            document.getElementById('errorCount').textContent = formatNumber(data.application.errors);
        }
        if (data.system) {
            const goroutines = data.system.goroutines || 0;
            document.getElementById('goroutinesValue').textContent = goroutines;
            document.getElementById('uptimeValue').textContent = formatUptime(data.system.uptime_seconds);
            const memUsage = data.system.memory_usage || 0;
            const cpuUsage = data.system.cpu_usage || 0;
            updateMetricBar('cpuBar', 'cpuValue', cpuUsage.toFixed(1), '%', 'cpu');
            updateMetricBar('memoryBar', 'memoryValue', memUsage.toFixed(1), ' MB', 'memory');
        }
    } catch (e) {
        console.error('获取指标失败:', e);
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

// 更新设备列表
function updateDeviceList() {
    const deviceMap = new Map();
    
    // 统计每个设备的数据
    realtimeData.forEach(item => {
        const deviceName = formatDeviceName(item.deviceName);
        if (!deviceMap.has(deviceName)) {
            deviceMap.set(deviceName, {
                name: deviceName,
                lastActive: item.timestamp,
                count: 0
            });
        }
        const device = deviceMap.get(deviceName);
        device.count++;
        if (item.timestamp > device.lastActive) {
            device.lastActive = item.timestamp;
        }
    });
    
    // 转换为数组并排序（按最后活跃时间）
    const devices = Array.from(deviceMap.values()).sort((a, b) => b.lastActive - a.lastActive);
    
    // 更新设备数量
    document.getElementById('deviceCount').textContent = devices.length + ' 台设备';
    
    // 更新设备表格
    const tbody = document.getElementById('deviceTableBody');
    tbody.innerHTML = '';
    
    const searchText = document.getElementById('deviceSearch').value.toLowerCase();
    const filtered = devices.filter(d => d.name.toLowerCase().includes(searchText));
    
    filtered.forEach(device => {
        const tr = document.createElement('tr');
        const lastActive = new Date(device.lastActive / 1000000).toLocaleTimeString();
        const status = device.lastActive > Date.now() - 5 * 60 * 1000000 ? '在线' : '离线';
        const statusClass = status === '在线' ? 'text-primary' : 'text-muted';
        
        tr.innerHTML = `
            <td class="device-name">${device.name}</td>
            <td>${lastActive}</td>
            <td>${device.count}</td>
            <td class="${statusClass}">${status}</td>
        `;
        tbody.appendChild(tr);
    });
}

// 过滤设备表
function filterDeviceTable() {
    updateDeviceList();
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

// 导出数据
function exportData() {
    const format = prompt('请选择导出格式:\n1. CSV\n2. JSON', '1');
    if (!format) return;
    
    const isCSV = format === '1';
    const fileName = `sfsEdgeStore_${new Date().toISOString().slice(0,10)}.${isCSV ? 'csv' : 'json'}`;
    
    if (isCSV) {
        exportToCSV(fileName);
    } else {
        exportToJSON(fileName);
    }
}

// 导出为CSV
function exportToCSV(fileName) {
    let csv = '时间,设备,传感器,数值,类型\n';
    realtimeData.forEach(item => {
        const time = new Date(item.timestamp / 1000000).toISOString();
        const device = formatDeviceName(item.deviceName);
        const reading = item.reading || '';
        const value = formatValue(item.value);
        const type = item.valueType || '';
        csv += `"${time}","${device}","${reading}","${value}","${type}"\n`;
    });
    downloadFile(fileName, csv, 'text/csv');
}

// 导出为JSON
function exportToJSON(fileName) {
    const exportData = realtimeData.map(item => ({
        time: new Date(item.timestamp / 1000000).toISOString(),
        device: formatDeviceName(item.deviceName),
        sensor: item.reading || '',
        value: item.value,
        type: item.valueType || ''
    }));
    const jsonStr = JSON.stringify(exportData, null, 2);
    downloadFile(fileName, jsonStr, 'application/json');
}

// 下载文件
function downloadFile(fileName, content, mimeType) {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

// 打开配置管理
function openConfig() {
    // 这里可以打开配置管理页面或模态框
    alert('配置管理功能开发中...');
}

// 窗口大小改变时重新调整图表
window.addEventListener('resize', function() {
    if (trendChart) {
        trendChart.resize();
    }
});

