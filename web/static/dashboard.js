// sfsEdgeStore 本地监控 Dashboard JS
let realtimeData = [];
let historicalData = [];
let selectedDevice = '';
let selectedReading = '';
let updateInterval = 5000; // 5秒刷新一次
let ws;

document.addEventListener('DOMContentLoaded', function() {
    connectWebSocket();
    fetchData();
    fetchMetrics();
    fetchLicenseInfo();
    fetchDeviceStatus();
    fetchDeviceAlerts();
    setInterval(fetchData, updateInterval);
    setInterval(fetchMetrics, updateInterval);
    setInterval(fetchDeviceStatus, updateInterval);
    setInterval(fetchDeviceAlerts, updateInterval);
});

// WebSocket 连接
function connectWebSocket() {
    ws = new WebSocket('ws://' + window.location.host + '/ws');
    
    ws.onopen = function() {
        console.log('WebSocket connected');
    };
    
    ws.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
        } catch (e) {
            console.error('Error parsing WebSocket message:', e);
        }
    };
    
    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
    
    ws.onclose = function() {
        console.log('WebSocket disconnected, reconnecting...');
        setTimeout(connectWebSocket, 3000);
    };
}

function handleWebSocketMessage(data) {
    switch (data.type) {
        case 'device_data':
            updateRealtimeData(data.data);
            break;
        case 'alerts':
            updateDeviceAlerts(data.data);
            updateUnhealthyDevices([data.data]);
            break;
        case 'device_status':
            updateDeviceStatus(data.data);
            break;
    }
}

function updateRealtimeData(deviceData) {
	const deviceName = deviceData.deviceName;
	const records = deviceData.records;
	
	// 添加新数据到实时数据列表
	records.forEach(record => {
		realtimeData.unshift(record);
		if (realtimeData.length > 21) {
			realtimeData.pop();
		}
	});
	
	// 更新表格
	updateTable();
	// 更新波形图
	updateWaveform();
	
	// 更新设备计数
	const dataCountElement = document.getElementById('dataCount');
	if (dataCountElement) {
		dataCountElement.textContent = realtimeData.length + ' 条记录';
	}
}

function updateDeviceAlerts(alertData) {
    const deviceName = alertData.deviceName;
    const alerts = alertData.alerts;
    
    // 直接刷新告警列表，不再使用 alertList 缓存
    fetchDeviceAlerts();
}

function updateUnhealthyDevices(alertGroups) {
    const container = document.getElementById('unhealthyDevices');
    if (!container) return;
    
    // 收集有告警的设备
    const unhealthyDeviceMap = new Map();
    alertGroups.forEach(g => {
        if (g.alerts && g.alerts.length > 0) {
            g.alerts.forEach(alert => {
                if (!unhealthyDeviceMap.has(g.deviceName)) {
                    unhealthyDeviceMap.set(g.deviceName, []);
                }
                unhealthyDeviceMap.get(g.deviceName).push(alert.message || '告警');
            });
        }
    });
    
    if (unhealthyDeviceMap.size === 0) {
        container.innerHTML = '<div class="text-white-50 text-center">暂无问题设备</div>';
        return;
    }
    
    let html = '<div class="row g-2">';
    unhealthyDeviceMap.forEach((alerts, name) => {
        const badgeClass = 'bg-warning';
        html += `<div class="col-12 col-sm-6 col-md-4">
            <div class="p-2 rounded" style="background: rgba(255,107,107,0.15); border: 1px solid rgba(255,107,107,0.3);">
                <div class="d-flex justify-content-between align-items-center">
                    <span class="text-danger fw-bold">${name}</span>
                    <span class="badge ${badgeClass} text-dark">${alerts.length}</span>
                </div>
                <div class="text-white-50 small mt-1">${alerts.slice(0, 2).join(', ')}${alerts.length > 2 ? '...' : ''}</div>
            </div>
        </div>`;
    });
    html += '</div>';
    container.innerHTML = html;
}

function formatNumber(num) {
    if (num === undefined || num === null || isNaN(num)) return '0';
    if (num >= 1e9) return (num / 1e9).toFixed(1) + 'B';
    if (num >= 1e6) return (num / 1e6).toFixed(1) + 'M';
    if (num >= 1e3) return (num / 1e3).toFixed(1) + 'k';
    return num.toString();
}

function formatBytes(bytes) {
    if (bytes === undefined || bytes === null || isNaN(bytes)) return '0 B';
    if (bytes >= 1e9) return (bytes / 1e9).toFixed(2) + ' GB';
    if (bytes >= 1e6) return (bytes / 1e6).toFixed(2) + ' MB';
    if (bytes >= 1e3) return (bytes / 1e3).toFixed(2) + ' KB';
    return bytes + ' B';
}

async function fetchData() {
    try {
        const res = await fetch('/api/readings?limit=21');
        const data = await res.json();
        realtimeData = data.readings || [];
        updateTable();
        await updateDeviceList();
        updateDeviceSelect();
        updateWaveform();

        const dataCountElement = document.getElementById('dataCount');
        if (dataCountElement) {
            dataCountElement.textContent = realtimeData.length + ' 条记录';
        }

        const connectionStatusElement = document.getElementById('connectionStatus');
        const connectionDotElement = document.getElementById('connectionDot');
        if (connectionStatusElement) {
            connectionStatusElement.textContent = '已连接';
        }
        if (connectionDotElement) {
            connectionDotElement.className = 'status-dot online';
        }
    } catch (e) {
        console.error('获取数据失败:', e);
        const connectionStatusElement = document.getElementById('connectionStatus');
        const connectionDotElement = document.getElementById('connectionDot');
        if (connectionStatusElement) {
            connectionStatusElement.textContent = '连接断开';
        }
        if (connectionDotElement) {
            connectionDotElement.className = 'status-dot offline';
        }
    }
}

async function fetchMetrics() {
    try {
        const res = await fetch('/metrics');
        const data = await res.json();
        if (data.application) {
            const mqttReceivedElement = document.getElementById('mqttReceived');
            if (mqttReceivedElement) {
                mqttReceivedElement.textContent = formatNumber(data.application.mqtt_messages_received);
            }
            const mqttProcessedElement = document.getElementById('mqttProcessed');
            if (mqttProcessedElement) {
                mqttProcessedElement.textContent = formatNumber(data.application.mqtt_messages_processed);
            }
            const errorCountElement = document.getElementById('errorCount');
            if (errorCountElement) {
                errorCountElement.textContent = formatNumber(data.application.errors);
            }
        }
        if (data.system) {
            const goroutines = data.system.goroutines || 0;
            const goroutinesValueElement = document.getElementById('goroutinesValue');
            if (goroutinesValueElement) {
                goroutinesValueElement.textContent = formatNumber(goroutines);
            }
            const uptimeValueElement = document.getElementById('uptimeValue');
            if (uptimeValueElement) {
                uptimeValueElement.textContent = formatUptime(data.system.uptime_seconds);
            }
            const memUsage = data.system.memory_usage || 0;
            const cpuUsage = data.system.cpu_usage || 0;
            updateMetricBar('cpuBar', 'cpuValue', cpuUsage.toFixed(1), '%', 'cpu');
            updateMetricBar('memoryBar', 'memoryValue', memUsage.toFixed(1), ' MB', 'memory');
        }
    } catch (e) {
        console.error('获取指标失败:', e);
    }

    const lastUpdateElement = document.getElementById('lastUpdate');
    if (lastUpdateElement) {
        lastUpdateElement.textContent = '更新: ' + new Date().toLocaleTimeString();
    }
}

function formatUptime(seconds) {
    if (!seconds || seconds <= 0) return '0s';
    if (seconds < 60) return Math.floor(seconds) + 's';
    if (seconds < 3600) return Math.floor(seconds / 60) + 'm';
    if (seconds < 86400) return Math.floor(seconds / 3600) + 'h';
    return Math.floor(seconds / 86400) + 'd';
}

function updateMetricBar(barId, valueId, value, unit, type) {
    const bar = document.getElementById(barId);
    const valueElement = document.getElementById(valueId);
    if (!bar || !valueElement) return;

    const numValue = parseFloat(value);
    const percentage = type === 'cpu' ? Math.min(numValue, 100) : Math.min(numValue, 100);
    bar.style.width = percentage + '%';

    if (type === 'cpu') {
        if (numValue > 80) {
            bar.style.background = 'linear-gradient(90deg, #ff4757, #ff6b81)';
        } else if (numValue > 50) {
            bar.style.background = 'linear-gradient(90deg, #ffa502, #ffbe76)';
        } else {
            bar.style.background = 'linear-gradient(90deg, #2ed573, #7bed9f)';
        }
    }

    valueElement.textContent = value + unit;
}

async function fetchDeviceStatus() {
    try {
        const res = await fetch('/api/device-status');
        const data = await res.json();
        const devices = data.devices || [];

        const onlineDevices = devices.filter(d => d.isOnline).length;
        const totalDevices = devices.length;

        const statusElement = document.getElementById('systemStatusValue');
        if (statusElement) {
            const healthyRatio = totalDevices > 0 ? (onlineDevices / totalDevices) : 0;
            if (healthyRatio >= 0.8) {
                statusElement.textContent = '健康';
                statusElement.className = 'text-primary';
            } else if (healthyRatio >= 0.5) {
                statusElement.textContent = '警告';
                statusElement.className = 'text-warning';
            } else {
                statusElement.textContent = '异常';
                statusElement.className = 'text-danger';
            }
        }

        const onlineCountElement = document.getElementById('onlineCount');
        const totalCountElement = document.getElementById('totalCount');
        if (onlineCountElement) onlineCountElement.textContent = onlineDevices;
        if (totalCountElement) totalCountElement.textContent = totalDevices;
    } catch (e) {
        console.error('获取设备状态失败:', e);
    }
}

async function fetchLicenseInfo() {
    try {
        const res = await fetch('/api/license');
        const data = await res.json();

        const versionElement = document.getElementById('licenseVersion');
        if (versionElement) {
            versionElement.textContent = data.version || '未知';
        }

        const devicesElement = document.getElementById('licenseDevices');
        if (devicesElement) {
            devicesElement.textContent = (data.deviceLimit || 0) + ' 台';
        }
    } catch (e) {
        console.error('获取许可证信息失败:', e);
    }
}

async function fetchDeviceAlerts() {
    try {
        const res = await fetch('/api/alert-groups');
        const data = await res.json();
        const groups = data.groups || [];
        const alertsContainer = document.getElementById('deviceAlerts');
        const alertCountElement = document.getElementById('deviceAlertCount');
        if (!alertsContainer) return;

        // 直接使用 API 返回的 groups，不使用 alertList 缓存
        // 更新告警计数
        if (alertCountElement) {
            alertCountElement.textContent = groups.length;
        }

        if (groups.length === 0) {
            alertsContainer.innerHTML = '<div class="text-center text-white-50 p-3">暂无告警</div>';
            return;
        }

        let html = '';
        for (let i = 0; i < groups.length && i < 10; i++) {
            const g = groups[i];
            const icon = g.Severity === 'critical' ? '🔴' : (g.Severity === 'warning' ? '🟠' : '🔵');
            const colorClass = g.Severity === 'critical' ? 'text-danger' : (g.Severity === 'warning' ? 'text-warning' : 'text-info');
            const bgClass = g.Severity === 'critical' ? 'bg-danger/10 border-danger/30' : (g.Severity === 'warning' ? 'bg-warning/10 border-warning/30' : 'bg-info/10 border-info/30');

            let timeStr = '未知';
            if (g.LastTime) {
                try {
                    const d = new Date(g.LastTime);
                    if (d && !isNaN(d.getTime())) {
                        timeStr = d.toLocaleTimeString();
                    }
                } catch (e) {}
            }

            let infoLine = '';
            if (g.Devices && g.Devices.length > 0) {
                const devs = g.Devices.slice(0, 3).join(', ');
                const more = g.DeviceCount > 3 ? '...' : '';
                infoLine = '<div class="small text-white-50 mt-1">' + devs + more + ' (' + g.DeviceCount + '台, ' + g.Count + '次)</div>';
            } else if (g.Count > 1) {
                infoLine = '<div class="small text-white-50 mt-1">共' + g.Count + '次</div>';
            }

            html += '<div class="alert-item p-3 rounded border ' + bgClass + '">' +
                '<div class="d-flex justify-content-between items-center">' +
                '<span class="' + colorClass + ' fw-bold">' + icon + ' ' + (g.Message || '未知告警') + '</span>' +
                '<span class="text-secondary small">' + timeStr + '</span></div>' +
                infoLine + '</div>';
        }
        alertsContainer.innerHTML = html;
        // 更新问题设备列表
        updateUnhealthyDevices(groups);
    } catch (e) {
        console.error('获取告警失败:', e);
        const alertsContainer = document.getElementById('deviceAlerts');
        const alertCountElement = document.getElementById('deviceAlertCount');
        if (alertsContainer) {
            alertsContainer.innerHTML = '<div class="text-center text-danger p-3">获取告警失败</div>';
        }
        if (alertCountElement) {
            alertCountElement.textContent = '0';
        }
    }
}

// 清除告警
function clearDeviceAlerts() {
    alertList = [];
    const alertsContainer = document.getElementById('deviceAlerts');
    const alertCountElement = document.getElementById('deviceAlertCount');
    if (alertsContainer) {
        alertsContainer.innerHTML = '<div class="text-center text-white-50 p-3">暂无告警</div>';
    }
    if (alertCountElement) {
        alertCountElement.textContent = '0';
    }
}

function updateUnhealthyDevices(groups) {
    const container = document.getElementById('unhealthyDevices');
    if (!container) return;

    // 收集有告警的设备
    const unhealthyDeviceMap = new Map();
    groups.forEach(g => {
        if (g.Devices && g.Devices.length > 0) {
            g.Devices.forEach(d => {
                if (!unhealthyDeviceMap.has(d)) {
                    unhealthyDeviceMap.set(d, []);
                }
                unhealthyDeviceMap.get(d).push(g.Message || '告警');
            });
        }
    });

    if (unhealthyDeviceMap.size === 0) {
        container.innerHTML = '<div class="text-white-50 text-center">暂无问题设备</div>';
        return;
    }

    let html = '<div class="row g-2">';
    unhealthyDeviceMap.forEach((alerts, name) => {
        const badgeClass = 'bg-warning';
        html += `<div class="col-12 col-sm-6 col-md-4">
            <div class="p-2 rounded" style="background: rgba(255,107,107,0.15); border: 1px solid rgba(255,107,107,0.3);">
                <div class="d-flex justify-content-between align-items-center">
                    <span class="text-danger fw-bold">${name}</span>
                    <span class="badge ${badgeClass} text-dark">${alerts.length}</span>
                </div>
                <div class="text-white-50 small mt-1">${alerts.slice(0, 2).join(', ')}${alerts.length > 2 ? '...' : ''}</div>
            </div>
        </div>`;
    });
    html += '</div>';
    container.innerHTML = html;
}

async function updateDeviceList() {
    try {
        const res = await fetch('/api/device-status');
        const data = await res.json();
        const devices = data.devices || [];

        const deviceCountElement = document.getElementById('deviceCount');
        if (deviceCountElement) {
            deviceCountElement.textContent = devices.length + ' 台设备';
        }

        const tbody = document.getElementById('deviceTableBody');
        if (!tbody) return;
        tbody.innerHTML = '';

        const deviceSearchElement = document.getElementById('deviceSearch');
        let searchText = '';
        if (deviceSearchElement) {
            searchText = deviceSearchElement.value.toLowerCase();
        }
        const filtered = devices.filter(d => d.deviceName.toLowerCase().includes(searchText));

        filtered.forEach(device => {
            const tr = document.createElement('tr');
            let lastActiveStr = '未知';
            try {
                const lastActiveDate = new Date(device.lastActive);
                if (!isNaN(lastActiveDate.getTime())) {
                    lastActiveStr = lastActiveDate.toLocaleTimeString();
                }
            } catch (e) {}
            const status = device.isOnline ? '在线' : '离线';
            const statusClass = status === '在线' ? 'text-primary' : 'text-muted';

            tr.innerHTML = `
                <td class="device-name">${device.deviceName}</td>
                <td>${lastActiveStr}</td>
                <td>${formatNumber(device.dataCount)}</td>
                <td class="${statusClass}">${status}</td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        console.error('获取设备列表失败:', e);
    }
}

function filterDeviceTable() {
    updateDeviceList();
}

function updateDeviceSelect() {
    const select = document.getElementById('deviceSelect');
    if (!select) return;

    const devices = [...new Set(realtimeData.map(d => d.deviceName))];
    const currentValue = select.value;

    select.innerHTML = '<option value="">选择设备</option>';
    devices.forEach(device => {
        const option = document.createElement('option');
        option.value = device;
        option.textContent = device;
        select.appendChild(option);
    });

    if (devices.includes(currentValue)) {
        select.value = currentValue;
    }
    selectedDevice = select.value;
}

function filterTable() {
    updateTable();
}

function updateTable() {
    const tbody = document.getElementById('tableBody');
    if (!tbody) return;
    tbody.innerHTML = '';

    const searchBox = document.getElementById('searchBox');
    let searchText = '';
    if (searchBox) {
        searchText = searchBox.value.toLowerCase();
    }

    const filtered = realtimeData.filter(d =>
        d.deviceName.toLowerCase().includes(searchText) ||
        d.reading.toLowerCase().includes(searchText)
    );

    filtered.forEach(item => {
        const tr = document.createElement('tr');
        const time = new Date(item.timestamp / 1000000).toLocaleTimeString();
        tr.innerHTML = `
            <td>${time}</td>
            <td class="device-name">${item.deviceName}</td>
            <td>${item.reading}</td>
            <td>${item.value}</td>
            <td>${item.valueType || 'Unknown'}</td>
        `;
        tbody.appendChild(tr);
    });
}

function updateWaveform() {
    const svg = document.getElementById('waveformSvg');
    if (!svg) return;

    const recentData = realtimeData.slice(-50);
    if (recentData.length < 2) {
        svg.innerHTML = '';
        return;
    }

    const width = svg.clientWidth || 800;
    const height = svg.clientHeight || 160;
    const padding = 10;
    const graphWidth = width - padding * 2;
    const graphHeight = height - padding * 2;

    const readingTypes = [...new Set(recentData.map(d => d.reading))].slice(0, 3);
    const colors = ['#ff6b6b', '#00d9ff', '#10b981'];

    let svgContent = '';

    readingTypes.forEach((reading, idx) => {
        const readingData = recentData.filter(d => d.reading === reading);
        if (readingData.length < 2) return;

        const values = readingData.map(d => parseFloat(d.value) || 0);
        const minVal = Math.min(...values);
        const maxVal = Math.max(...values);
        const range = maxVal - minVal || 1;

        const points = readingData.map((d, i) => {
            const x = padding + (i / (readingData.length - 1)) * graphWidth;
            const y = padding + graphHeight - ((parseFloat(d.value) - minVal) / range) * graphHeight;
            return `${x},${y}`;
        });

        const polyline = points.join(' ');
        svgContent += `<polyline points="${polyline}" fill="none" stroke="${colors[idx]}" stroke-width="2" stroke-linejoin="round"/>`;
    });

    svg.innerHTML = svgContent;
}
