// sfsEdgeStore Local Monitoring Dashboard JS
let realtimeData = [];
let historicalData = [];
let selectedDevice = '';
let selectedReading = '';
let updateInterval = 10000; // 10 seconds refresh (reduce CPU usage)
let ws;

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
            // Refresh data
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
    fetchDeviceStatus();
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
function openSettings() {
    fetch('/api/config/get')
        .then(response => response.json())
        .then(data => {
            const config = data.config || data;
            
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
        retention_days: parseInt(document.getElementById('settingRetentionDays').value)
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
    
    alert('Recommended settings applied. Please review and click Save & Apply to save.');
}

// Data Retention
function openRetentionSettings() {
    // Get current settings from server
    fetch('/api/retention/status')
        .then(response => response.json())
        .then(data => {
            // Show modal
            const modal = new bootstrap.Modal(document.getElementById('retentionSettingsModal'));
            modal.show();
            
            // Set values after modal is shown
            setTimeout(() => {
                // Adapt to backend API response format
                const retentionStatus = data.data || data;
                
                // Always use server returned values, even if enabled is false
                const retentionDaysEl = document.getElementById('retentionDays');
                if (retentionDaysEl) {
                    // Use server value if available, otherwise use default
                    if (retentionStatus.retention_days > 0) {
                        retentionDaysEl.value = retentionStatus.retention_days;
                    } else {
                        retentionDaysEl.value = 30;
                    }
                }
                
                // Convert cleanup interval
                const cleanupInterval = retentionStatus.cleanup_interval || 24;
                let intervalValue = 'daily';
                if (cleanupInterval === 168) { // 7 days
                    intervalValue = 'weekly';
                } else if (cleanupInterval === 720) { // 30 days
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
            
            // Show modal
            const modal = new bootstrap.Modal(document.getElementById('retentionSettingsModal'));
            modal.show();
            
            // Set default values after modal is shown
            setTimeout(() => {
                // Use default values
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

// Topic Subscription
function openTopicSubscription() {
    window.location.href = '/mqtt-subscription';
}

function saveRetentionSettings() {
    const retentionDays = document.getElementById('retentionDays').value;
    const cleanupInterval = document.getElementById('cleanupInterval').value;
    
    // Convert cleanup interval to hours
    let cleanupIntervalHours = 24; // Default daily
    if (cleanupInterval === 'weekly') {
        cleanupIntervalHours = 168; // 7 days
    } else if (cleanupInterval === 'monthly') {
        cleanupIntervalHours = 720; // 30 days
    }
    
    // Build config object
    const configData = {
        enable_retention_policy: true,
        retention_days: parseInt(retentionDays),
        cleanup_interval_hours: cleanupIntervalHours
    };
    
    // Send to server
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
            alert('Data retention saved!');
            
            // Close modal
            const modal = bootstrap.Modal.getInstance(document.getElementById('retentionSettingsModal'));
            modal.hide();
        } else {
            alert('Save failed: ' + (data.message || 'Unknown error'));
        }
    })
    .catch(error => {
        console.error('Save data retention failed:', error);
        alert('Save failed, please check network connection');
    });
}

document.addEventListener('DOMContentLoaded', function() {
    connectWebSocket();
    fetchData();
    fetchMetrics();
    fetchLicenseInfo();
    fetchDeviceStatus();
    fetchDeviceAlerts();
    // Stagger intervals to reduce CPU spikes
    setInterval(fetchData, updateInterval);
    setInterval(fetchMetrics, updateInterval);
    // Device status/alerts: refresh every 30s (less frequent)
    setInterval(fetchDeviceStatus, updateInterval * 3);
    setInterval(fetchDeviceAlerts, updateInterval * 3);
});

// WebSocket Connection
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
	
	// Add new data to realtime data list
	records.forEach(record => {
		realtimeData.unshift(record);
		if (realtimeData.length > 21) {
			realtimeData.pop();
		}
	});
	
	// Update table
	updateTable();
	// Update waveform
	updateWaveform();
	
	// Update data count
	const dataCountElement = document.getElementById('dataCount');
	if (dataCountElement) {
		dataCountElement.textContent = realtimeData.length + ' records';
	}
}

function updateDeviceAlerts(alertData) {
    const deviceName = alertData.deviceName;
    const alerts = alertData.alerts;
    
    // Directly refresh alert list, no longer using alertList cache
    fetchDeviceAlerts();
}

function updateUnhealthyDevices(alertGroups) {
    const container = document.getElementById('unhealthyDevices');
    if (!container) return;
    
    // Collect devices with alerts
    const unhealthyDeviceMap = new Map();
    alertGroups.forEach(g => {
        if (g.alerts && g.alerts.length > 0) {
            g.alerts.forEach(alert => {
                if (!unhealthyDeviceMap.has(g.deviceName)) {
                    unhealthyDeviceMap.set(g.deviceName, []);
                }
                unhealthyDeviceMap.get(g.deviceName).push(alert.message || 'Alert');
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
    
    // Convert reading to lowercase for case-insensitive matching
    const lowerReading = reading.toLowerCase();
    
    // Check if reading is in the unit map
    if (unitMap[lowerReading]) {
        return unitMap[lowerReading];
    }
    
    // Check for common prefixes or suffixes
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
    
    // Default to empty string if no unit found
    return '';
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
        console.error('Failed to fetch data:', e);
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
            const processed = data.application.mqtt_messages_processed || 0;
            const filtered = data.application.mqtt_messages_filtered || 0;

            // Update total messages
            const mqttTotalElement = document.getElementById('mqttTotal');
            if (mqttTotalElement) {
                mqttTotalElement.textContent = formatNumber(total);
            }

            // Update valid/processed data
            const mqttProcessedElement = document.getElementById('mqttProcessed');
            if (mqttProcessedElement) {
                mqttProcessedElement.textContent = formatNumber(processed);
            }

            // Update filtered count
            const mqttFilteredElement = document.getElementById('mqttFiltered');
            if (mqttFilteredElement) {
                mqttFilteredElement.textContent = formatNumber(filtered);
            }

            // Estimate breakdown (filtered is total filtered, split into non-event and invalid)
            const nonEventEstimate = Math.round(filtered * 0.6); // ~60% non-event
            const invalidEstimate = filtered - nonEventEstimate; // ~40% invalid/missing values

            const nonEventElement = document.getElementById('nonEventCount');
            if (nonEventElement) {
                nonEventElement.textContent = formatNumber(nonEventEstimate);
            }

            const invalidElement = document.getElementById('invalidValueCount');
            if (invalidElement) {
                invalidElement.textContent = formatNumber(invalidEstimate);
            }

            // Update total records stored
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
        console.error('Failed to fetch metrics:', e);
    }

    const lastUpdateElement = document.getElementById('lastUpdate');
    if (lastUpdateElement) {
        lastUpdateElement.textContent = 'Updated: ' + new Date().toLocaleTimeString();
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
                statusElement.textContent = 'Healthy';
                statusElement.className = 'text-primary';
            } else if (healthyRatio >= 0.5) {
                statusElement.textContent = 'Warning';
                statusElement.className = 'text-warning';
            } else {
                statusElement.textContent = 'Critical';
                statusElement.className = 'text-danger';
            }
        }

        const onlineCountElement = document.getElementById('onlineCount');
        const totalCountElement = document.getElementById('totalCount');
        if (onlineCountElement) onlineCountElement.textContent = onlineDevices;
        if (totalCountElement) totalCountElement.textContent = totalDevices;
    } catch (e) {
        console.error('Failed to fetch device status:', e);
    }
}

async function fetchLicenseInfo() {
    try {
        const res = await fetch('/api/license');
        const data = await res.json();

        const versionElement = document.getElementById('licenseVersion');
        if (versionElement) {
            versionElement.textContent = data.version || 'Unknown';
        }

        const devicesElement = document.getElementById('licenseDevices');
        if (devicesElement) {
            devicesElement.textContent = (data.deviceLimit || 0) + ' devices';
        }
    } catch (e) {
        console.error('Failed to fetch license info:', e);
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

        // Use API returned groups directly, no alertList cache
        // Update alert count
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
        // Update unhealthy devices list
        updateUnhealthyDevices(groups);
    } catch (e) {
        console.error('Failed to fetch alerts:', e);
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
    alertList = [];
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

    // Collect devices with alerts
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

function updateDeviceSelect() {
    const select = document.getElementById('deviceSelect');
    if (!select) return;

    const devices = [...new Set(realtimeData.map(d => d.deviceName))];
    const currentValue = select.value;

    select.innerHTML = '<option value="">Select Device</option>';
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
            // Refresh data
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
    fetchDeviceStatus();
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
