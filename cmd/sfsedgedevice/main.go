package main

import (
	"flag"
	"fmt"
	"os"

	"sfsEdgeStore/devconfig"
	"sfsEdgeStore/provision"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	configDir := "./devconfig"
	if envDir := os.Getenv("SFSCONFIG_DIR"); envDir != "" {
		configDir = envDir
	}

	switch os.Args[1] {
	case "config":
		handleConfig(os.Args[2:], configDir)
	case "provision":
		handleProvision(os.Args[2:], configDir)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`sfsEdgeStore Device Management CLI

Usage:
  sfsedgedevice <command> [options]

Commands:
  config              Manage configuration and templates
  provision           Provision devices to EdgeX
  help                Show this help message

Config Subcommands:
  sfsedgedevice config list-templates          List available device templates
  sfsedgedevice config list-devices           List configured devices
  sfsedgedevice config add <device>            Add a device from CSV
  sfsedgedevice config remove <name>           Remove a device
  sfsedgedevice config validate                Validate configuration

Provision Subcommands:
  sfsedgedevice provision add <name> <ip> [--protocol] [--template] [--interval]
  sfsedgedevice provision all                  Provision all devices from CSV
  sfsedgedevice provision list                 List devices in EdgeX
  sfsedgedevice provision remove <name>       Remove device from EdgeX
  sfsedgedevice provision validate             Validate EdgeX connection

Environment Variables:
  SFSCONFIG_DIR       Configuration directory (default: ./config)

Examples:
  # List available templates
  sfsedgedevice config list-templates

  # Add a device
  sfsedgedevice provision add sensor-001 192.168.1.10 --template modbus/temperature --interval 1s

  # Provision all devices from CSV
  sfsedgedevice provision all

  # Validate configuration
  sfsedgedevice config validate`)
}

func handleConfig(args []string, configDir string) {
	if len(args) < 1 {
		fmt.Println("Usage: sfsedgedevice config <subcommand>")
		os.Exit(1)
	}

	cm, err := devconfig.NewConfigManager(configDir)
	if err != nil {
		fmt.Printf("Failed to create config manager: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "list-templates":
		templates, err := cm.ListTemplates()
		if err != nil {
			fmt.Printf("Failed to list templates: %v\n", err)
			os.Exit(1)
		}
		if len(templates) == 0 {
			fmt.Println("No templates found")
			return
		}
		fmt.Println("Available templates:")
		for _, t := range templates {
			files, _ := cm.ListTemplateFiles(t)
			fmt.Printf("  %s/\n", t)
			for _, f := range files {
				fmt.Printf("    - %s\n", f)
			}
		}

	case "list-devices":
		mode := cm.GetConfigMode()
		if mode == "yaml" {
			fmt.Println("📄 Mode: YAML (GitOps Mode)")
		} else if mode == "csv" {
			fmt.Println("📄 Mode: CSV (Excel Mode)")
		}

		devices, err := cm.LoadDevices()
		if err != nil {
			fmt.Printf("Failed to load devices: %v\n", err)
			os.Exit(1)
		}
		if len(devices) == 0 {
			fmt.Println("No devices configured")
			return
		}
		fmt.Printf("Configured devices (%d):\n", len(devices))
		for _, d := range devices {
			fmt.Printf("  - %s (IP: %s, Protocol: %s, Template: %s, Interval: %s)\n",
				d.Name, d.IP, d.Protocol, d.Template, d.Interval)
		}

	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: sfsedgedevice config add <csv-file>")
			os.Exit(1)
		}
		fmt.Printf("Adding devices from %s...\n", args[1])
		fmt.Println("(Note: Use 'provision add' to provision a device to EdgeX)")

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: sfsedgedevice config remove <device-name>")
			os.Exit(1)
		}
		if err := cm.RemoveDevice(args[1]); err != nil {
			fmt.Printf("Failed to remove device: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Device '%s' removed from configuration\n", args[1])

	case "validate":
		errors, err := cm.ValidateConfig()
		if err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		}
		if len(errors) == 0 {
			fmt.Println("Configuration is valid!")
		} else {
			fmt.Println("Configuration errors found:")
			for _, e := range errors {
				fmt.Printf("  - %s\n", e)
			}
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown config subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleProvision(args []string, configDir string) {
	if len(args) < 1 {
		fmt.Println("Usage: sfsedgedevice provision <subcommand>")
		os.Exit(1)
	}

	cm, err := devconfig.NewConfigManager(configDir)
	if err != nil {
		fmt.Printf("Failed to create config manager: %v\n", err)
		os.Exit(1)
	}

	p := provision.NewProvisioner(cm, "")

	switch args[0] {
	case "add":
		handleProvisionAdd(args[1:], p, cm)

	case "all":
		fmt.Println("Provisioning all devices from configuration...")
		if err := p.ProvisionAll(); err != nil {
			fmt.Printf("Provisioning failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Provisioning completed!")

	case "list":
		if err := p.ListDevices(); err != nil {
			fmt.Printf("Failed to list devices: %v\n", err)
			os.Exit(1)
		}

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: sfsedgedevice provision remove <device-name>")
			os.Exit(1)
		}
		if err := p.RemoveDevice(args[1]); err != nil {
			fmt.Printf("Failed to remove device: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Device '%s' removed from EdgeX\n", args[1])

	case "validate":
		if err := p.ValidateConnection(); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("EdgeX connection is valid!")

	default:
		fmt.Printf("Unknown provision subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleProvisionAdd(args []string, p *provision.Provisioner, cm *devconfig.ConfigManager) {
	fs := flag.NewFlagSet("provision add", flag.ContinueOnError)
	protocol := fs.String("protocol", "modbus-tcp", "Device protocol (modbus-tcp, modbus-rtu, mqtt)")
	template := fs.String("template", "modbus/temperature", "Device template")
	interval := fs.String("interval", "15s", "AutoEvent interval (e.g., 1s, 5s, 1m)")
	onChange := fs.Bool("onChange", false, "Trigger only on value change")
	minInterval := fs.String("minInterval", "", "Minimum trigger interval")
	subscriptionTopic := fs.String("subscriptionTopic", "", "MQTT subscription topic")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("Invalid arguments: %v\n", err)
		os.Exit(1)
	}

	if fs.NArg() < 2 {
		fmt.Println("Usage: sfsedgedevice provision add <name> <ip> [options]")
		fs.PrintDefaults()
		os.Exit(1)
	}

	name := fs.Arg(0)
	ip := fs.Arg(1)

	fmt.Printf("Adding device '%s' at %s (protocol: %s, template: %s, interval: %s)...\n",
		name, ip, *protocol, *template, *interval)

	if err := p.ProvisionDevice(name, ip, *protocol, *template, *interval, *onChange, *minInterval, *subscriptionTopic); err != nil {
		fmt.Printf("Failed to provision device: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Device '%s' provisioned successfully!\n", name)
}
