/**
 * sfsEdgeStore 国际化翻译系统
 * 提供多语言支持，默认语言为英文（en-US）
 * 
 * 使用方法：
 * 1. 在HTML元素中添加 data-i18n 属性，值为翻译键
 *    例如：<span data-i18n="nav.dashboard">Dashboard</span>
 * 
 * 2. 在JavaScript中使用 t() 函数获取翻译
 *    例如：const translatedText = t('nav.dashboard');
 * 
 * 3. 切换语言
 *    例如：switchLanguage('zh-CN');
 * 
 * 如何添加新翻译：
 * 1. 在 i18n 对象中添加新的翻译键值对
 * 2. 确保所有语言都有对应的翻译
 * 3. 在开发环境下，系统会自动检测翻译缺失
 */

const i18n = {
    // 中文翻译
    'zh-CN': {
        // 导航栏
        'nav.dashboard': '仪表盘',
        'nav.mqtt.config': 'MQTT配置',
        'nav.subscription': '订阅主题',
        'nav.auth.config': '认证配置',
        'nav.api.keys': 'API密钥',
        'nav.data.records': '数据记录',
        'nav.subscription.status': '订阅状态',
        'nav.system.health': '系统健康',
        'nav.settings': '设置',

        // 页面标题和副标题
        'dashboard.title': '仪表盘',
        'dashboard.subtitle': '系统运行状态和关键指标监控',
        'mqtt.config.title': 'MQTT配置',
        'mqtt.config.subtitle': '配置MQTT连接参数和订阅选项',
        'subscription.title': 'MQTT订阅管理',
        'subscription.subtitle': '管理MQTT订阅主题，支持添加、删除和测试自定义主题',

        // 连接状态
        'connection.status': '连接状态',
        'connected': '已连接',
        'disconnected': '未连接',
        'connecting': '连接中...',
        'unknown': '未知',

        // MQTT配置
        'broker.address': 'Broker地址',
        'client.id': '客户端ID',
        'topic': '主题',
        'connection.timeout': '连接超时（秒）',
        'keep.alive': '保持连接（秒）',

        // TLS配置
        'enable.tls': '启用TLS/SSL',
        'ca.cert': 'CA证书',
        'client.cert': '客户端证书',
        'client.key': '客户端密钥',

        // 订阅设置
        'enable.auto.subscribe': '启用自动订阅标准主题',

        // 按钮和操作
        'save': '保存',
        'cancel': '取消',
        'add': '添加',
        'delete': '删除',
        'test': '测试',
        'reset': '重置',
        'refresh': '刷新数据',
        'fullscreen': '全屏模式',
        'export': '导出数据',
        'config.management': '配置管理',

        // 主题分类
        'standard.topics': '标准订阅主题',
        'custom.topics': '自定义订阅主题',
        'topic.test': '主题测试',

        // 统计信息
        'total.topics': '总主题数',
        'standard.topics.count': '标准主题',
        'custom.topics.count': '自定义主题',

        // 状态和操作
        'status': '状态',
        'action': '操作',
        'subscribed': '已订阅',
        'not.active': '未激活',

        // 主题测试
        'add.topic': '添加订阅主题',
        'test.topic': '测试主题',
        'test.result': '测试结果',
        'waiting.test': '等待测试...',
        'testing': '正在测试...',
        'test.success': '主题测试成功！',
        'test.failed': '测试失败',

        // 帮助信息
        'topic.format.info': '主题格式说明',
        'topic.format.wildcard': '通配符说明',
        'auto.subscribe.info': '自动订阅说明',

        // 空状态
        'no.custom.topics': '暂无自定义主题，点击上方按钮添加',
        'loading': '加载中...',

        // 通用状态
        'success': '成功',
        'error': '错误',
        'warning': '警告',

        // 操作结果
        'config.saved': '配置保存成功',
        'config.save.failed': '配置保存失败',
        'topic.added': '主题添加成功',
        'topic.add.failed': '主题添加失败',
        'topic.deleted': '主题删除成功',
        'topic.delete.failed': '主题删除失败',

        // 提示信息
        'confirm.delete': '确定要删除这个主题吗？',
        'please.input.topic': '请输入主题',
        'topic.exists': '主题已存在',
        'topic.not.found': '主题不存在',

        // 语言选择
        'language': '语言',
        'chinese': '中文',
        'english': 'English',

        // 设备列表
        'device.list': '设备列表',
        'device.name': '设备名称',
        'last.active': '最后活跃',
        'data.count': '数据量',
        'search.device': '搜索设备...',
        'devices': '台设备',

        // 仪表盘
        'realtime.data': '实时数据',
        'records': '条记录',
        'history.trend': '历史趋势',
        'system.status': '系统状态',
        'mqtt.status': 'MQTT 状态',
        'error.statistics': '错误统计',
        'quick.actions': '快速操作',
        'oneclick.config': '一键配置',
        'select.device': '选择设备',
        'select.sensor': '选择传感器',

        // 系统信息
        'mqtt.broker': 'MQTT Broker',
        'mqtt.topic': 'MQTT 主题',
        'http.port': 'HTTP端口',
        'edgex.version': 'EdgeX版本',
        'memory.usage': '内存占用',

        // 配置提示
        'config.hint': '配置提示',
        'config.hint.content': '留空将使用智能默认值。配置修改后需要重启服务才能生效。',

        // 安全配置
        'security.config': '安全配置',
        'username': '用户名',
        'password': '密码',

        // 通用字段
        'type': '类型',
        'description': '描述',
        'added.time': '添加时间',
        'active': '激活',
        'inactive': '未激活',

        // 页面标题
        'api.keys.title': 'API密钥管理',
        'api.keys.subtitle': '管理和查看API密钥',
        'auth.config.title': '认证配置',
        'auth.config.subtitle': '配置用户名和密码认证',
        'data.records.title': '数据记录',
        'data.records.subtitle': '查看和管理数据记录',
        'settings.title': '设置',
        'settings.subtitle': '系统设置和配置',
        'subscription.status.title': '订阅状态',
        'subscription.status.subtitle': '查看MQTT订阅状态和统计',
        'system.health.title': '系统健康',
        'system.health.subtitle': '系统运行状态和性能指标',
        'tls.config.title': 'TLS配置',
        'tls.config.subtitle': '配置TLS/SSL安全连接',

        // 通配符说明
        'wildcard.plus': '+ - 匹配单个层级',
        'wildcard.hash': '# - 匹配任意层级',
        'auto.subscribe.description': '自动订阅 - 系统根据EdgeX版本自动订阅相应的标准主题',

        // 占位符
        'topic.placeholder': '例如: devices/+/data',
        'wildcard.support': '支持通配符 + 和 #',
        
        // 实时数据表格
        'time': '时间',
        'device': '设备',
        'sensor': '传感器',
        'value': '数值',
        'type': '类型',
        
        // 搜索框
        'search.device.name': '搜索设备名称...',
        
        // 系统状态
        'cpu': 'CPU',
        'memory': '内存',
        'uptime': '运行时间',
        
        // MQTT状态
        'received': '已接收',
        'processed': '已处理',
        'total.messages': '总消息数',
        'valid.data': '有效数据',
        'filtered': '已过滤',
        'filtered.non.event': '非事件消息',
        'filtered.invalid': '无效/缺失值',
        'process.rate': '处理率',
        'db.operations': '数据库操作',
        'records.stored': '已存储记录',
        
        // 不健康设备
        'unhealthy.devices': '不健康设备',
        'no.unhealthy.devices': '暂无不健康设备',
        
        // 告警中心
        'alert.center': '告警中心',
        'clear': '清除',
        'loading': '加载中...',
        
        // 数据保留
        'data.retention': '数据保留',
        'retention.period': '保留周期（天）',
        'cleanup.interval': '清理间隔',
        'daily': '每日',
        'weekly': '每周',
        'monthly': '每月',
        'save.settings': '保存设置',
        
        // 自定义主题
        'custom.topics.description': '自定义主题用于订阅非标准MQTT主题或第三方系统主题',
        'auto.subscribe.after.add': '添加后将自动订阅，无需重启服务',
        
        // Settings Modal
        'settings': '设置',
        'settings.general': '通用',
        'settings.thresholds': '阈值',
        'connection.settings': '🔗 连接配置',
        'storage.settings': '💾 存储配置',
        'resource.retention': '📊 资源与保留策略',
        'db.path': '数据库路径',
        'db.scenario': '数据库场景',
        'resource.monitoring': '资源监控',
        'max.memory': '最大内存 (MB)',
        'retention.policy': '保留策略',
        'retention.days': '保留天数',
        'analyzer.enable': '🔍 分析引擎',
        'analyzer.enable.label': '启用',
        'threshold.add': '➕ 添加阈值',
        'threshold.add.btn': '添加',
        'threshold.list': '📋 阈值列表',
        'threshold.device': '设备',
        'threshold.reading': '读数',
        'threshold.min': '下限',
        'threshold.max': '上限',
        'threshold.action': '操作',
        'threshold.empty': '未配置阈值',
        'apply.recommended': '✨ 推荐配置',
        'note': '💡 MQTT Topic',
        'mqtt.topic.note': '在「主题订阅」页面管理'
    },
    
    // 英文翻译
    'en-US': {
        // 导航栏
        'nav.dashboard': 'Dashboard',
        'nav.mqtt.config': 'MQTT Config',
        'nav.subscription': 'Subscription',
        'nav.auth.config': 'Auth Config',
        'nav.api.keys': 'API Keys',
        'nav.data.records': 'Data Records',
        'nav.subscription.status': 'Subscription Status',
        'nav.system.health': 'System Health',
        'nav.settings': 'Settings',

        // 页面标题和副标题
        'dashboard.title': 'Dashboard',
        'dashboard.subtitle': 'System status and key metrics monitoring',
        'mqtt.config.title': 'MQTT Configuration',
        'mqtt.config.subtitle': 'Configure MQTT connection parameters and subscription options',
        'subscription.title': 'MQTT Subscription Management',
        'subscription.subtitle': 'Manage MQTT subscription topics, support adding, deleting and testing custom topics',

        // 连接状态
        'connection.status': 'Connection Status',
        'connected': 'Connected',
        'disconnected': 'Disconnected',
        'connecting': 'Connecting...',
        'unknown': 'Unknown',

        // MQTT配置
        'broker.address': 'Broker Address',
        'client.id': 'Client ID',
        'topic': 'Topic',
        'connection.timeout': 'Connection Timeout (s)',
        'keep.alive': 'Keep Alive (s)',

        // TLS配置
        'enable.tls': 'Enable TLS/SSL',
        'ca.cert': 'CA Certificate',
        'client.cert': 'Client Certificate',
        'client.key': 'Client Key',

        // 订阅设置
        'enable.auto.subscribe': 'Enable Auto Subscribe Standard Topics',

        // 按钮和操作
        'save': 'Save',
        'cancel': 'Cancel',
        'add': 'Add',
        'delete': 'Delete',
        'test': 'Test',
        'reset': 'Reset',
        'refresh': 'Refresh Data',
        'fullscreen': 'Fullscreen',
        'export': 'Export Data',
        'config.management': 'Config Management',

        // 主题分类
        'standard.topics': 'Standard Subscription Topics',
        'custom.topics': 'Custom Subscription Topics',
        'topic.test': 'Topic Test',

        // 统计信息
        'total.topics': 'Total Topics',
        'standard.topics.count': 'Standard Topics',
        'custom.topics.count': 'Custom Topics',

        // 状态和操作
        'status': 'Status',
        'action': 'Action',
        'subscribed': 'Subscribed',
        'not.active': 'Not Active',

        // 主题测试
        'add.topic': 'Add Subscription Topic',
        'test.topic': 'Test Topic',
        'test.result': 'Test Result',
        'waiting.test': 'Waiting for test...',
        'testing': 'Testing...',
        'test.success': 'Topic test successful!',
        'test.failed': 'Test failed',

        // 帮助信息
        'topic.format.info': 'Topic Format Info',
        'topic.format.wildcard': 'Wildcard Info',
        'auto.subscribe.info': 'Auto Subscribe Info',

        // 空状态
        'no.custom.topics': 'No custom topics, click the button above to add',
        'loading': 'Loading...',

        // 通用状态
        'success': 'Success',
        'error': 'Error',
        'warning': 'Warning',

        // 操作结果
        'config.saved': 'Configuration saved successfully',
        'config.save.failed': 'Configuration save failed',
        'topic.added': 'Topic added successfully',
        'topic.add.failed': 'Topic add failed',
        'topic.deleted': 'Topic deleted successfully',
        'topic.delete.failed': 'Topic delete failed',

        // 提示信息
        'confirm.delete': 'Are you sure you want to delete this topic?',
        'please.input.topic': 'Please input topic',
        'topic.exists': 'Topic already exists',
        'topic.not.found': 'Topic not found',

        // 语言选择
        'language': 'Language',
        'chinese': '中文',
        'english': 'English',

        // 设备列表
        'device.list': 'Device List',
        'device.name': 'Device Name',
        'last.active': 'Last Active',
        'data.count': 'Data Count',
        'search.device': 'Search device...',
        'devices': 'devices',

        // 仪表盘
        'realtime.data': 'Real-time Data',
        'records': 'records',
        'history.trend': 'History Trend',
        'system.status': 'System Status',
        'mqtt.status': 'MQTT Status',
        'error.statistics': 'Error Statistics',
        'quick.actions': 'Quick Actions',
        'oneclick.config': 'One-click Config',
        'select.device': 'Select Device',
        'select.sensor': 'Select Sensor',

        // 系统信息
        'mqtt.broker': 'MQTT Broker',
        'mqtt.topic': 'MQTT Topic',
        'http.port': 'HTTP Port',
        'edgex.version': 'EdgeX Version',
        'memory.usage': 'Memory Usage',

        // 配置提示
        'config.hint': 'Configuration Hint',
        'config.hint.content': 'Empty values will use smart defaults. Configuration changes require restart to take effect.',

        // 安全配置
        'security.config': 'Security Configuration',
        'username': 'Username',
        'password': 'Password',

        // 通用字段
        'type': 'Type',
        'description': 'Description',
        'added.time': 'Added Time',
        'active': 'Active',
        'inactive': 'Inactive',

        // 页面标题
        'api.keys.title': 'API Keys Management',
        'api.keys.subtitle': 'Manage and view API keys',
        'auth.config.title': 'Authentication Configuration',
        'auth.config.subtitle': 'Configure username and password authentication',
        'data.records.title': 'Data Records',
        'data.records.subtitle': 'View and manage data records',
        'settings.title': 'Settings',
        'settings.subtitle': 'System settings and configuration',
        'subscription.status.title': 'Subscription Status',
        'subscription.status.subtitle': 'View MQTT subscription status and statistics',
        'system.health.title': 'System Health',
        'system.health.subtitle': 'System running status and performance metrics',
        'tls.config.title': 'TLS Configuration',
        'tls.config.subtitle': 'Configure TLS/SSL secure connection',

        // 通配符说明
        'wildcard.plus': '+ - Matches single level',
        'wildcard.hash': '# - Matches multiple levels',
        'auto.subscribe.description': 'Auto Subscribe - System automatically subscribes to standard topics based on EdgeX version',

        // 占位符
        'topic.placeholder': 'e.g.: devices/+/data',
        'wildcard.support': 'Supports wildcards + and #',
        
        // 实时数据表格
        'time': 'Time',
        'device': 'Device',
        'sensor': 'Sensor',
        'value': 'Value',
        'type': 'Type',
        
        // 搜索框
        'search.device.name': 'Search device name...',
        
        // 系统状态
        'cpu': 'CPU',
        'memory': 'Memory',
        'uptime': 'Uptime',
        
        // MQTT Status
        'received': 'Received',
        'processed': 'Processed',
        'total.messages': 'Total Messages',
        'valid.data': 'Valid Data',
        'filtered': 'Filtered',
        'filtered.non.event': 'Non-Event Messages',
        'filtered.invalid': 'Invalid/Missing Values',
        'process.rate': 'Process Rate',
        'db.operations': 'DB Operations',
        'records.stored': 'Records Stored',
        
        // 不健康设备
        'unhealthy.devices': 'Unhealthy Devices',
        'no.unhealthy.devices': 'No unhealthy devices',
        
        // 告警中心
        'alert.center': 'Alert Center',
        'clear': 'Clear',
        'loading': 'Loading...',
        
        // 数据保留
        'data.retention': 'Data Retention',
        'retention.period': 'Retention Period (days)',
        'cleanup.interval': 'Cleanup Interval',
        'daily': 'Daily',
        'weekly': 'Weekly',
        'monthly': 'Monthly',
        'save.settings': 'Save Settings',
        
        // 自定义主题
        'custom.topics.description': 'Custom topics are used to subscribe to non-standard MQTT topics or third-party system topics',
        'auto.subscribe.after.add': 'Will automatically subscribe after adding, no service restart required',
        
        // Settings Modal
        'settings': 'Settings',
        'settings.general': 'General',
        'settings.thresholds': 'Thresholds',
        'connection.settings': '🔗 Connection',
        'storage.settings': '💾 Storage',
        'resource.retention': '📊 Resource & Retention',
        'db.path': 'Database Path',
        'db.scenario': 'DB Scenario',
        'resource.monitoring': 'Resource Monitor',
        'max.memory': 'Max Memory (MB)',
        'retention.policy': 'Retention Policy',
        'retention.days': 'Retention Days',
        'analyzer.enable': '🔍 Analyzer',
        'analyzer.enable.label': 'Enable',
        'threshold.add': '➕ Add Threshold',
        'threshold.add.btn': 'Add',
        'threshold.list': '📋 Threshold List',
        'threshold.device': 'Device',
        'threshold.reading': 'Reading',
        'threshold.min': 'Min',
        'threshold.max': 'Max',
        'threshold.action': 'Action',
        'threshold.empty': 'No thresholds configured',
        'apply.recommended': '✨ Recommended',
        'note': '💡 MQTT Topic',
        'mqtt.topic.note': 'Managed in Topic Subscription page'
    }
};

// 缓存当前语言和翻译对象，减少重复查找
let currentLang = null;
let currentTexts = null;

function getCurrentLanguage() {
    return localStorage.getItem('language') || 'en-US';
}

function switchLanguage(lang) {
    localStorage.setItem('language', lang);
    // 清除缓存，强制重新加载
    currentLang = null;
    currentTexts = null;
    updatePageText();

    if (document.getElementById('languageSelector')) {
        document.getElementById('languageSelector').value = lang;
    }
}

function getCurrentTexts() {
    const lang = getCurrentLanguage();
    if (currentLang !== lang) {
        currentLang = lang;
        currentTexts = i18n[lang];
    }
    return currentTexts;
}

function updatePageText() {
    const texts = getCurrentTexts();

    // 优化DOM操作：一次性选择所有需要翻译的元素
    const elementsToTranslate = [
        ...document.querySelectorAll('[data-i18n]'),
        ...document.querySelectorAll('[data-i18n-placeholder]'),
        ...document.querySelectorAll('[data-i18n-title]')
    ];

    // 批量处理翻译，减少DOM操作次数
    elementsToTranslate.forEach(element => {
        // 处理data-i18n属性
        if (element.hasAttribute('data-i18n')) {
            const key = element.getAttribute('data-i18n');
            if (texts[key]) {
                element.textContent = texts[key];
            }
        }
        
        // 处理data-i18n-placeholder属性
        if (element.hasAttribute('data-i18n-placeholder')) {
            const key = element.getAttribute('data-i18n-placeholder');
            if (texts[key]) {
                element.placeholder = texts[key];
            }
        }
        
        // 处理data-i18n-title属性
        if (element.hasAttribute('data-i18n-title')) {
            const key = element.getAttribute('data-i18n-title');
            if (texts[key]) {
                element.title = texts[key];
            }
        }
    });
}

function t(key) {
    const texts = getCurrentTexts();
    if (!texts[key]) {
        console.warn(`Translation missing for key: "${key}" in language: ${currentLang}`);
    }
    return texts[key] || key;
}

// 检测翻译缺失
function detectTranslationMissing() {
    const languages = Object.keys(i18n);
    const referenceLanguage = 'en-US';
    const referenceKeys = Object.keys(i18n[referenceLanguage]);
    
    console.log('=== Translation Missing Detection ===');
    
    languages.forEach(lang => {
        if (lang === referenceLanguage) return;
        
        const missingKeys = [];
        referenceKeys.forEach(key => {
            if (!i18n[lang][key]) {
                missingKeys.push(key);
            }
        });
        
        if (missingKeys.length > 0) {
            console.warn(`Missing translations for language ${lang}:`, missingKeys);
        } else {
            console.log(`All translations present for language ${lang}`);
        }
    });
    
    console.log('=== Translation Detection Complete ===');
}

document.addEventListener('DOMContentLoaded', function() {
    if (document.getElementById('languageSelector')) {
        document.getElementById('languageSelector').value = getCurrentLanguage();
    }
    updatePageText();
    
    // 开发环境下检测翻译缺失
    // 检查当前URL是否为本地开发环境
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        detectTranslationMissing();
    }
});
