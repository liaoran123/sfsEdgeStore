// sfsEdgeStore Local Monitoring Dashboard JS
let realtimeData = [];
let historicalData = [];
let selectedDevice = '';
let selectedReading = '';
let updateInterval = 3000; // 3 seconds refresh
let ws;
let wsConnected = false;
let pollingTimers = [];

function startPolling() {
    stopPolling();
    pollingTimers.push(setInterval(() => { if (!wsConnected) fetchData(); }, updateInterval));
    pollingTimers.push(setInterval(fetchMetrics, updateInterval));
    pollingTimers.push(setInterval(() => { if (!wsConnected) fetchDeviceAlerts(); }, updateInterval * 3));
}

function stopPolling() {
    pollingTimers.forEach(clearInterval);
    pollingTimers = [];
}

// One-click Config
function oneClickConfig() {
    fetch('/api/config/oneclick', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            alert('One-click config successful!');
            fetchData();
        } else {
            alert('Config failed: ' + (data.message || 'Unknown error'));
        }
    })
    .catch(error => {
        console.error('Config failed:', error);
        alert('Config failed, please check network connection');
    });
}

// Refresh Data
function refreshData() {
    fetchData();
    fetchMetrics();
    fetchDeviceAlerts();
    alert('Data refreshed!');
}

// Export Data
function exportData() {
    const exportType = prompt('Please select export format:\n1. CSV\n2. JSON', '1');

    if (exportType === '1') {
        window.open('/api/export/csv', '_blank');
    } else if (exportType === '2') {
        window.open('/api/export/json', '_blank');
    }
}

// Settings Modal
var currentThresholds = {};

function openSettings() {
    fetch('/api/config/get')
        .then(response => response.json())
        .then(data => {
            const config = data.data || data.config || data;

            const modal = new bootstrap.Modal(document.getElementById('settingsModal'));
            modal.show();

            setTimeout(() => {
                document.getElementById('settingMqttBroker').value = config.mqtt_broker || 'tcp://localhost:1883';
                document.getElementById('settingDbPath').value = config.db_path || 'data/sfs.db';
                document.getElementById('settingDbScenario').value = config.db_scenario || 'edge';
                document.getElementById('settingHttpPort').value = config.http_port || '8081';

                document.getElementById('settingResourceMonitoring').checked = config.enable_resource_monitoring || false;
                document.getElementById('settingMaxMemory').value = config.max_memory_mb || 45;
                document.getElementById('settingRetentionPolicy').checked = config.enable_retention_policy || false;
                document.getElementById('settingRetentionDays').value = config.retention_days || 30;
                document.getElementById('settingEnableAnalyzer').checked = config.enable_analyzer || false;

                currentThresholds = config.analyzer_thresholds || {};
                renderThresholdTable();
            }, 100);
        })
        .catch(error => {
            console.error('Failed to get config:', error);

            const modal = new bootstrap.Modal(document.getElementById('settingsModal'));
            modal.show();

            setTimeout(() => {
                document.getElementById('settingMqttBroker').value = 'tcp://localhost:1883';
                document.getElementById('settingDbPath').value = 'data/sfs.db';
                document.getElementById('settingDbScenario').value = 'edge';
                document.getElementById('settingHttpPort').value = '8081';
                document.getElementById('settingResourceMonitoring').checked = false;
                document.getElementById('settingMaxMemory').value = 45;
                document.getElementById('settingRetentionPolicy').checked = false;
                document.getElementById('settingRetentionDays').value = 30;
                document.getElementById('settingEnableAnalyzer').checked = false;

                currentThresholds = {};
                renderThresholdTable();
            }, 100);
        });
}

function saveSettings() {
    const configData = {
        mqtt_broker: document.getElementById('settingMqttBroker').value,
        db_path: document.getElementById('settingDbPath').value,
        db_scenario: document.getElementById('settingDbScenario').value,
        http_port: document.getElementById('settingHttpPort').value,
        enable_resource_monitoring: document.getElementById('settingResourceMonitoring').checked,
        max_memory_mb: parseInt(document.getElementById('settingMaxMemory').value),
        enable_retention_policy: document.getElementById('settingRetentionPolicy').checked,
        retention_days: parseInt(document.getElementById('settingRetentionDays').value),
        enable_analyzer: document.getElementById('settingEnableAnalyzer').checked,
        analyzer_thresholds: currentThresholds
    };

    fetch('/api/config/update', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(configData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            alert('Settings saved successfully! A restart may be required for some changes to take effect.');
            const modal = bootstrap.Modal.getInstance(document.getElementById('settingsModal'));
            modal.hide();
            fetchData();
            fetchMetrics();
        } else {
            alert('Save failed: ' + (data.message || 'Unknown error'));
        }
    })
    .catch(error => {
        console.error('Save settings failed:', error);
        alert('Save failed, please check network connection');
    });
}

function applyRecommendedSettings() {
    document.getElementById('settingMqttBroker').value = 'tcp://localhost:1883';
    document.getElementById('settingDbPath').value = 'data/sfs.db';
    document.getElementById('settingDbScenario').value = 'edge';
    document.getElementById('settingHttpPort').value = '8081';
    document.getElementById('settingResourceMonitoring').checked = true;
    document.getElementById('settingMaxMemory').value = 45;
    document.getElementById('settingRetentionPolicy').checked = true;
    document.getElementById('settingRetentionDays').value = 30;
    document.getElementById('settingEnableAnalyzer').checked = true;

    alert('Recommended settings applied. Please review and click Save & Apply to save.');
}

// Threshold Management
function renderThresholdTable() {
    var tbody = document.getElementById('thresholdTableBody');
    var noMsg = document.getElementById('noThresholdsMsg');
    tbody.innerHTML = '';

    var keys = Object.keys(currentThresholds);
    if (keys.length === 0) {
        noMsg.style.display = 'block';
        return;
    }

    noMsg.style.display = 'none';
    keys.forEach(function(key) {
        var t = currentThresholds[key];
        var isDeviceThreshold = key.includes(':');
        var device = isDeviceThreshold ? key.split(':')[0] : '(default)';
        var reading = isDeviceThreshold ? key.split(':')[1] : key;

        var tr = document.createElement('tr');
        tr.innerHTML =
            '<td>' + device + '</td>' +
            '<td>' + reading + '</td>' +
            '<td>' + t.min + '</td>' +
            '<td>' + t.max + '</td>' +
            '<td><button class="btn btn-sm btn-danger" onclick="deleteThreshold(\'' + key + '\')">Delete</button></td>';
        tbody.appendChild(tr);
    });
}

function addThreshold() {
    var device = document.getElementById('thresholdDevice').value.trim();
    var reading = document.getElementById('thresholdReading').value.trim();
    var min = parseFloat(document.getElementById('thresholdMin').value);
    var max = parseFloat(document.getElementById('thresholdMax').value);

    if (!reading || isNaN(min) || isNaN(max)) {
        alert('Please fill in Reading name, Min and Max values');
        return;
    }

    var key = device ? device + ':' + reading : reading;
    currentThresholds[key] = { min: min, max: max, device: device };

    document.getElementById('thresholdDevice').value = '';
    document.getElementById('thresholdReading').value = '';
    document.getElementById('thresholdMin').value = '';
    document.getElementById('thresholdMax').value = '';

    renderThresholdTable();
}

function deleteThreshold(key) {
    delete currentThresholds[key];
    renderThresholdTable();
}

// Data Retention
function openRetentionSettings() {
    fetch('/api/retention/status')
        .then(response => response.json())
        .then(data => {
            const modal = new bootstrap.Modal(document.getElementById('retentionSettingsModal'));
            modal.show();

            setTimeout(() => {
                const retentionStatus = data.data || data;

                const retentionDaysEl = document.getElementById('retentionDays');
                if (retentionDaysEl) {
                    if (retentionStatus.retention_days > 0) {
                        retentionDaysEl.value = retentionStatus.retention_days;
                    } else {
                        retentionDaysEl.value = 30;
                    }
                }

                const cleanupInterval = retentionStatus.cleanup_interval || 24;
                let intervalValue = 'daily';
                if (cleanupInterval === 168) {
                    intervalValue = 'weekly';
                } else if (cleanupInterval === 720) {
                    intervalValue = 'monthly';
                }

                const cleanupIntervalEl = document.getElementById('cleanupInterval');
                if (cleanupIntervalEl) {
                    cleanupIntervalEl.value = intervalValue;
                }
            }, 100);
        })
        .catch(error => {
            console.error('Failed to get data retention:', error);

            const modal = new bootstrap.Modal(document.getElementById('retentionSettingsModal'));
            modal.show();

            setTimeout(() => {
                const retentionDaysEl = document.getElementById('retentionDays');
                if (retentionDaysEl) {
                    retentionDaysEl.value = 30;
                }

                const cleanupIntervalEl = document.getElementById('cleanupInterval');
                if (cleanupIntervalEl) {
                    cleanupIntervalEl.value = 'daily';
                }
            }, 100);
        });
}

// Topic Subscription - opens settings modal to subscription tab
function openTopicSubscription() {
    openSettingsTab('subscription');
}

// Retention Settings - opens settings modal to retention tab
function openRetentionSettings() {
    openSettingsTab('retention');
}

function openSettingsTab(tabId) {
    const modal = new bootstrap.Modal(document.getElementById('settingsModal'));
    modal.show();

    setTimeout(() => {
        const tab = document.getElementById(tabId + '-tab');
        if (tab) {
            tab.click();
        }

        if (tabId === 'subscription') {
            subLoadData();
        } else if (tabId === 'retention') {
            subLoadRetention();
        }
    }, 200);
}

// ===== Subscription Tab Functions =====
let subCustomTopics = [];

function subLoadData() {
    subLoadConnectionStatus();
    subLoadTopics();
}

function subLoadConnectionStatus() {
    fetch('/api/subscription/status')
        .then(response => response.json())
        .then(result => {
            const data = result.data || {};
            const el = document.getElementById('subConnectionStatus');
            const brokerEl = document.getElementById('subBroker');
            if (el) el.textContent = data.connected ? 'Connected' : 'Disconnected';
            if (el) el.style.color = data.connected ? '#4CAF50' : '#ff6b6b';
            if (brokerEl) brokerEl.textContent = data.broker || '-';
        })
        .catch(() => {
            const el = document.getElementById('subConnectionStatus');
            if (el) el.textContent = 'Unknown';
        });
}

function subLoadTopics() {
    fetch('/api/subscription/themes')
        .then(response => response.json())
        .then(result => {
            if (result.status === 'success') {
                subCustomTopics = result.custom_topics || [];
                subRenderCustomTopics();
            }
        })
        .catch(() => {});
}

function subRenderCustomTopics() {
    const tbody = document.getElementById('subCustomTopics');
    if (!tbody) return;

    if (subCustomTopics.length === 0) {
        tbody.innerHTML = '<tr><td colspan="3" class="text-center text-muted">No custom topics</td></tr>';
        return;
    }

    tbody.innerHTML = subCustomTopics.map((t, i) =>
        `<tr>
            <td class="topic-name" style="font-family:'Courier New',monospace;color:var(--primary-color)">${t.topic}</td>
            <td><span class="badge ${t.active ? 'bg-success' : 'bg-warning'}">${t.active ? 'Subscribed' : 'Not Active'}</span></td>
            <td><button class="btn btn-danger btn-sm" onclick="subRemoveTopic(${i})">Delete</button></td>
        </tr>`
    ).join('');
}

function showAddTopicModal() {
    const modal = new bootstrap.Modal(document.getElementById('addTopicModal'));
    document.getElementById('subNewTopic').value = '';
    modal.show();
}

function subAddTopic() {
    const topic = document.getElementById('subNewTopic').value.trim();
    if (!topic) { alert('Please enter topic'); return; }

    fetch('/api/subscription/themes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic: topic })
    })
    .then(response => response.json())
    .then(result => {
        if (result.status === 'success') {
            const modal = bootstrap.Modal.getInstance(document.getElementById('addTopicModal'));
            if (modal) modal.hide();
            subLoadTopics();
        } else {
            alert('Add failed: ' + (result.error || 'Unknown error'));
        }
    })
    .catch(error => alert('Add topic failed: ' + error.message));
}

function subRemoveTopic(index) {
    if (!confirm('Are you sure you want to delete this topic?')) return;
    const topic = subCustomTopics[index];
    fetch('/api/subscription/themes', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic: topic.topic })
    })
    .then(response => response.json())
    .then(result => {
        if (result.status === 'success') subLoadTopics();
        else alert('Delete failed: ' + (result.error || 'Unknown error'));
    })
    .catch(error => alert('Delete failed: ' + error.message));
}

function subTestTopic() {
    const topic = document.getElementById('subTestTopic').value.trim();
    const resultEl = document.getElementById('subTestResult');
    if (!topic) { resultEl.innerHTML = '<span style="color:#ffc107;">Please enter topic</span>'; return; }

    resultEl.innerHTML = '<span style="color:#17a2b8;">Testing...</span>';
    fetch('/api/subscription/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic: topic })
    })
    .then(response => response.json())
    .then(result => {
        resultEl.innerHTML = result.status === 'success'
            ? '<span style="color:#4CAF50;">✅ Test successful!</span>'
            : '<span style="color:#dc3545;">❌ Test failed</span>';
    })
    .catch(error => resultEl.innerHTML = '<span style="color:#dc3545;">❌ Error: ' + error.message + '</span>');
}

// ===== Retention Tab Functions =====
function subLoadRetention() {
    fetch('/api/retention/status')
        .then(response => response.json())
        .then(data => {
            const status = data.data || data;
            const daysEl = document.getElementById('subRetentionDays');
            const intervalEl = document.getElementById('subCleanupInterval');
            if (daysEl && status.retention_days > 0) daysEl.value = status.retention_days;
            if (intervalEl) {
                const h = status.cleanup_interval || 24;
                intervalEl.value = h === 168 ? 'weekly' : h === 720 ? 'monthly' : 'daily';
            }
        })
        .catch(() => {});
}

function subSaveRetention() {
    const days = document.getElementById('subRetentionDays').value;
    const interval = document.getElementById('subCleanupInterval').value;
    let hours = 24;
    if (interval === 'weekly') hours = 168;
    else if (interval === 'monthly') hours = 720;

    fetch('/api/config/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            enable_retention_policy: true,
            retention_days: parseInt(days),
            cleanup_interval_hours: hours
        })
    })
    .then(response => response.json())
    .then(data => {
        const el = document.getElementById('subRetentionResult');
        if (!el) return;
        if (data.status === 'success') {
            el.innerHTML = '<span style="color:#4CAF50;">✅ Retention settings saved</span>';
            setTimeout(() => el.innerHTML = '', 3000);
        } else {
            el.innerHTML = '<span style="color:#dc3545;">❌ ' + (data.error || 'Save failed') + '</span>';
        }
    })
    .catch(() => {
        const el = document.getElementById('subRetentionResult');
        if (el) el.innerHTML = '<span style="color:#dc3545;">❌ Save failed</span>';
    });
}

function subManualCleanup() {
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '🧹 Cleaning...';

    fetch('/api/retention/cleanup', { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            const el = document.getElementById('subRetentionResult');
            if (!el) return;
            if (data.status === 'success') {
                el.innerHTML = '<span style="color:#4CAF50;">✅ Cleanup completed, deleted ' + data.deleted_count + ' records</span>';
            } else {
                el.innerHTML = '<span style="color:#dc3545;">❌ ' + (data.error || 'Cleanup failed') + '</span>';
            }
        })
        .catch(() => {
            const el = document.getElementById('subRetentionResult');
            if (el) el.innerHTML = '<span style="color:#dc3545;">❌ Cleanup failed</span>';
        })
        .finally(() => {
            btn.disabled = false;
            btn.textContent = '🧹 Run Cleanup Now';
        });
}

document.addEventListener('DOMContentLoaded', function() {
    connectWebSocket();
    fetchMetrics();
    fetchDeviceAlerts();
    startPolling();

    // Load subscription data when the tab is shown
    document.getElementById('subscription-tab').addEventListener('shown.bs.tab', function () {
        subLoadData();
    });
    document.getElementById('retention-tab').addEventListener('shown.bs.tab', function () {
        subLoadRetention();
    });
});

document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        stopPolling();
    } else {
        fetchData();
        fetchMetrics();
        fetchDeviceAlerts();
        startPolling();
    }
});

// WebSocket Connection
function connectWebSocket() {
    ws = new WebSocket('ws://' + window.location.host + '/ws');

    ws.onopen = function() {
        wsConnected = true;
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
        wsConnected = false;
        console.log('WebSocket disconnected, reconnecting...');
        setTimeout(connectWebSocket, 3000);
    };
}

function handleWebSocketMessage(data) {
    try {
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
    } catch (e) {
        // Silent fail for WebSocket message errors
    }
}

function updateRealtimeData(deviceData) {
    try {
        if (!deviceData) return;

        const records = deviceData.records || [];
        records.forEach(record => {
            realtimeData.unshift(record);
            if (realtimeData.length > 21) {
                realtimeData.pop();
            }
        });

        updateTable();
        updateWaveform();
        updateDeviceAlerts(deviceData.alerts);
    } catch (e) {
        // Silent fail for UI update errors
    }
}

function updateDeviceAlerts(alertData) {
    // WebSocket 提供实时告警，HTTP 轮询 30s 周期负责同步
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

// Get unit based on reading type
function getUnitByReading(reading) {
    const unitMap = {
        'temperature': '°C',
        'humidity': '%',
        'pressure': 'hPa',
        'voltage': 'V',
        'current': 'A',
        'power': 'W',
        'energy': 'kWh',
        'speed': 'km/h',
        'flow': 'L/min',
        'level': 'm',
        'weight': 'kg',
        'distance': 'm',
        'frequency': 'Hz',
        'resistance': 'Ω',
        'conductivity': 'μS/cm',
        'pH': 'pH',
        'ORP': 'mV',
        'dissolvedOxygen': 'mg/L',
        'turbidity': 'NTU',
        'co2': 'ppm',
        'o2': 'ppm',
        'nox': 'ppm',
        'sox': 'ppm',
        'pm25': 'μg/m³',
        'pm10': 'μg/m³'
    };

    const lowerReading = reading.toLowerCase();

    if (unitMap[lowerReading]) {
        return unitMap[lowerReading];
    }

    if (lowerReading.includes('temp')) {
        return '°C';
    } else if (lowerReading.includes('humid')) {
        return '%';
    } else if (lowerReading.includes('press')) {
        return 'hPa';
    } else if (lowerReading.includes('volt')) {
        return 'V';
    } else if (lowerReading.includes('current')) {
        return 'A';
    } else if (lowerReading.includes('power')) {
        return 'W';
    } else if (lowerReading.includes('energy')) {
        return 'kWh';
    }

    return '';
}

async function fetchData() {
    try {
        const res = await fetch('/api/readings?limit=21');
        const data = await res.json();
        realtimeData = data.readings || [];
        updateTable();
        await updateDeviceList();
        updateWaveform();

        const dataCountElement = document.getElementById('dataCount');
        if (dataCountElement) {
            dataCountElement.textContent = realtimeData.length + ' records';
        }

        const connectionStatusElement = document.getElementById('connectionStatus');
        const connectionDotElement = document.getElementById('connectionDot');
        if (connectionStatusElement) {
            connectionStatusElement.textContent = 'Connected';
        }
        if (connectionDotElement) {
            connectionDotElement.className = 'status-dot online';
        }
    } catch (e) {
        const connectionStatusElement = document.getElementById('connectionStatus');
        const connectionDotElement = document.getElementById('connectionDot');
        if (connectionStatusElement) {
            connectionStatusElement.textContent = 'Disconnected';
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
            const total = data.application.mqtt_messages_received || 0;
            const filtered = data.application.mqtt_messages_filtered || 0;

            const mqttTotalElement = document.getElementById('mqttTotal');
            if (mqttTotalElement) {
                mqttTotalElement.textContent = formatNumber(total);
            }

            const mqttFilteredElement = document.getElementById('mqttFiltered');
            if (mqttFilteredElement) {
                mqttFilteredElement.textContent = formatNumber(filtered);
            }

            const nonEventEstimate = Math.round(filtered * 0.6);
            const invalidEstimate = filtered - nonEventEstimate;

            const nonEventElement = document.getElementById('nonEventCount');
            if (nonEventElement) {
                nonEventElement.textContent = formatNumber(nonEventEstimate);
            }

            const invalidElement = document.getElementById('invalidValueCount');
            if (invalidElement) {
                invalidElement.textContent = formatNumber(invalidEstimate);
            }

            const recordsStored = data.application.total_records_stored || 0;
            const recordsStoredElement = document.getElementById('totalRecordsStored');
            if (recordsStoredElement) {
                recordsStoredElement.textContent = formatNumber(recordsStored);
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
        // Silent fail
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

async function fetchDeviceAlerts() {
    try {
        const res = await fetch('/api/alert-groups');
        const data = await res.json();
        const groups = data.groups || [];
        const alertsContainer = document.getElementById('deviceAlerts');
        const alertCountElement = document.getElementById('deviceAlertCount');
        if (!alertsContainer) return;

        if (alertCountElement) {
            alertCountElement.textContent = groups.length;
        }

        if (groups.length === 0) {
            alertsContainer.innerHTML = '<div class="text-center text-white-50 p-3">No alerts</div>';
            return;
        }

        let html = '';
        for (let i = 0; i < groups.length && i < 10; i++) {
            const g = groups[i];
            const icon = g.Severity === 'critical' ? '🔴' : (g.Severity === 'warning' ? '🟠' : '🔵');
            const colorClass = g.Severity === 'critical' ? 'text-danger' : (g.Severity === 'warning' ? 'text-warning' : 'text-info');
            const bgClass = g.Severity === 'critical' ? 'bg-danger/10 border-danger/30' : (g.Severity === 'warning' ? 'bg-warning/10 border-warning/30' : 'bg-info/10 border-info/30');

            let timeStr = 'Unknown';
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
                infoLine = '<div class="small text-white-50 mt-1">' + devs + more + ' (' + g.DeviceCount + ' devices, ' + g.Count + ' times)</div>';
            } else if (g.Count > 1) {
                infoLine = '<div class="small text-white-50 mt-1">Total ' + g.Count + ' times</div>';
            }

            html += '<div class="alert-item p-3 rounded border ' + bgClass + '">' +
                '<div class="d-flex justify-content-between items-center">' +
                '<span class="' + colorClass + ' fw-bold">' + icon + ' ' + (g.Message || 'Unknown Alert') + '</span>' +
                '<span class="text-secondary small">' + timeStr + '</span></div>' +
                infoLine + '</div>';
        }
        alertsContainer.innerHTML = html;
        updateUnhealthyDevices(groups);
    } catch (e) {
        const alertsContainer = document.getElementById('deviceAlerts');
        const alertCountElement = document.getElementById('deviceAlertCount');
        if (alertsContainer) {
            alertsContainer.innerHTML = '<div class="text-center text-danger p-3">Failed to fetch alerts</div>';
        }
        if (alertCountElement) {
            alertCountElement.textContent = '0';
        }
    }
}

// Clear Alerts
function clearDeviceAlerts() {
    const alertsContainer = document.getElementById('deviceAlerts');
    const alertCountElement = document.getElementById('deviceAlertCount');
    if (alertsContainer) {
        alertsContainer.innerHTML = '<div class="text-center text-white-50 p-3">No alerts</div>';
    }
    if (alertCountElement) {
        alertCountElement.textContent = '0';
    }
}

function updateUnhealthyDevices(groups) {
    const container = document.getElementById('unhealthyDevices');
    if (!container) return;

    const unhealthyDeviceMap = new Map();
    groups.forEach(g => {
        if (g.Devices && g.Devices.length > 0) {
            g.Devices.forEach(d => {
                if (!unhealthyDeviceMap.has(d)) {
                    unhealthyDeviceMap.set(d, []);
                }
                unhealthyDeviceMap.get(d).push(g.Message || 'Alert');
            });
        }
    });

    if (unhealthyDeviceMap.size === 0) {
        container.innerHTML = '<div class="text-white-50 text-center">No unhealthy devices</div>';
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
            deviceCountElement.textContent = devices.length + ' devices';
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
            let lastActiveStr = 'Unknown';
            try {
                const lastActiveDate = new Date(device.lastActive);
                if (!isNaN(lastActiveDate.getTime())) {
                    lastActiveStr = lastActiveDate.toLocaleTimeString();
                }
            } catch (e) {}
            const status = device.isOnline ? 'Online' : 'Offline';
            const statusClass = status === 'Online' ? 'text-primary' : 'text-muted';

            tr.innerHTML = `
                <td class="device-name">${device.deviceName}</td>
                <td>${lastActiveStr}</td>
                <td>${formatNumber(device.dataCount)}</td>
                <td class="${statusClass}">${status}</td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        console.error('Failed to fetch device list:', e);
    }
}

function filterDeviceTable() {
    updateDeviceList();
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
        const unit = getUnitByReading(item.reading);
        const valueWithUnit = unit ? `${item.value} ${unit}` : item.value;
        tr.innerHTML = `
            <td>${time}</td>
            <td class="device-name">${item.deviceName}</td>
            <td>${item.reading}</td>
            <td>${valueWithUnit}</td>
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

        const values = readingData.map(d => parseFloat(d.value)).filter(v => !isNaN(v));
        if (values.length < 2) return;

        const minVal = Math.min(...values);
        const maxVal = Math.max(...values);
        const range = maxVal - minVal || 1;

        const points = readingData.map((d, i) => {
            const val = parseFloat(d.value);
            if (isNaN(val)) return null;
            const x = padding + (i / (readingData.length - 1)) * graphWidth;
            const y = padding + graphHeight - ((val - minVal) / range) * graphHeight;
            return `${x},${y}`;
        }).filter(p => p !== null);

        if (points.length < 2) return;

        const polyline = points.join(' ');
        svgContent += `<polyline points="${polyline}" fill="none" stroke="${colors[idx]}" stroke-width="2" stroke-linejoin="round"/>`;
    });

    svg.innerHTML = svgContent;
}