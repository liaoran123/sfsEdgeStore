// cli/cli.go
// 命令行参数处理包
package cli

import (
	"flag"
	"fmt"
)

// Args 命令行参数结构体
type Args struct {
	MQTTBroker string
	MQTTTopic  string
	HTTPPort   string
	Help       bool
}

// Parse 解析命令行参数
func Parse() *Args {
	mqttBroker := flag.String("broker", "", "MQTT Broker地址 (例如: tcp://localhost:1883)")
	mqttTopic := flag.String("topic", "", "MQTT订阅主题 (例如: edgex/events/core/#)")
	httpPort := flag.String("port", "", "HTTP服务端口 (例如: 8081)")
	help := flag.Bool("help", false, "显示帮助信息")
	flag.Parse()

	return &Args{
		MQTTBroker: *mqttBroker,
		MQTTTopic:  *mqttTopic,
		HTTPPort:   *httpPort,
		Help:       *help,
	}
}

// ShowHelp 显示帮助信息
func (a *Args) ShowHelp() {
	fmt.Println("sfsEdgeStore - 轻量级工业物联网边缘数据存储适配器")
	fmt.Println("\n用法: sfsedgestore [选项]")
	fmt.Println("\n选项:")
	fmt.Println("  -broker <地址>   MQTT Broker地址 (例如: tcp://localhost:1883)")
	fmt.Println("  -topic <主题>    MQTT订阅主题 (例如: edgex/events/core/#)")
	fmt.Println("  -port <端口>     HTTP服务端口 (例如: 8081)")
	fmt.Println("  -help            显示帮助信息")
	fmt.Println("\n示例:")
	fmt.Println("  sfsedgestore")
	fmt.Println("  sfsedgestore -broker tcp://192.168.1.100:1883 -port 8082")
}
func PrintWelcome() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════╗
║                    sfsEdgeStore 启动成功                       ║
╠═══════════════════════════════════════════════════════════════╣
║  轻量级工业物联网边缘数据存储适配器                             ║
║  内存占用: <50MB | 极轻量 | 高可靠                              ║
╠═══════════════════════════════════════════════════════════════╣
║  Web 监控界面: http://localhost:8081                          ║
║  停止方法: Ctrl+C                                              ║
╚═══════════════════════════════════════════════════════════════╝
`)
}
