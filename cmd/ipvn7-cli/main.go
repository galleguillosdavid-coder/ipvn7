package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"ipv7/adapters"
	"ipv7/core"
	"ipv7/dht"
)

const (
	Version = "ipvn7-nos-v2.0"
	Banner  = `
  ██╗██████╗ ██╗   ██╗███╗   ██╗███████╗
  ██║██╔══██╗██║   ██║████╗  ██║╚════██║
  ██║██████╔╝██║   ██║██╔██╗ ██║    ██╔╝
  ██║██╔═══╝ ╚██╗ ██╔╝██║╚██╗██║   ██╔╝ 
  ██║██║      ╚████╔╝ ██║ ╚████║   ██║  
  ╚═╝╚═╝       ╚═══╝  ╚═╝  ╚═══╝   ╚═╝  
   Network Operating System · Control Plane CLI
`
)

func getIPv7Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ipv7"
	}
	dir := filepath.Join(home, ".ipv7")
	_ = os.MkdirAll(dir, 0700)
	return dir
}

func printHelp() {
	fmt.Print(Banner)
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Uso: ipvn7-cli <comando> [argumentos]")
	fmt.Print("\nComandos disponibles:\n")
	fmt.Println("  status                         Muestra la identidad local DID y estado del sistema")
	fmt.Println("  petname list                   Lista los nombres amigables (dDNS) registrados")
	fmt.Println("  petname add <nombre> <did>     Asocia un alias legible a un DID soberano")
	fmt.Println("  petname resolve <nombre>       Resuelve un alias a su DID y direcciones virtuales")
	fmt.Println("  petname delete <nombre>        Elimina un alias de la libreta local")
	fmt.Println("  petname export-hosts           Exporta formato compatible /etc/hosts")
	fmt.Println("  firewall list                  Muestra las reglas de micro-segmentación ZTNA")
	fmt.Println("  firewall add <did> <puerto> <action>  Agrega una regla ZTNA (action: ALLOW|DENY)")
	fmt.Println("  firewall check <did> <puerto>  Evalúa si un paquete sería admitido o descartado")
	fmt.Println("  kuzu <cypher>                  Ejecuta una consulta Cypher en el grafo local")
	fmt.Println("  version                        Muestra la versión de ingeniería de ipvn7 NOS")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	ipv7Dir := getIPv7Dir()
	contactsPath := filepath.Join(ipv7Dir, "contacts.json")
	firewallPath := filepath.Join(ipv7Dir, "firewall_rules.json")

	petStore := dht.NewPetnameStore(contactsPath, true)
	fw := adapters.NewZTNAFirewall(adapters.ActionDeny)
	_ = fw.LoadFromFile(firewallPath)

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Printf("ipvn7 Network Operating System (NOS) — Control Plane CLI\nVersión: %s\n", Version)

	case "status":
		fmt.Print(Banner)
		keyPath := filepath.Join(ipv7Dir, "identity.key")
		id, _, err := core.LoadOrCreatePersistentIdentity(keyPath)
		if err != nil {
			fmt.Printf("[ERROR] No se pudo cargar la identidad: %v\n", err)
			return
		}
		fmt.Println("==================================================================")
		fmt.Println("               ESTADO DEL NODO ipvn7 NETWORK OS                   ")
		fmt.Println("==================================================================")
		fmt.Printf("[ID]  Sovereign DID   : did:ipv7:%s\n", id.String())
		fmt.Printf("[KEY] Clave en disco  : %s\n", keyPath)
		fmt.Printf("[IP4] Virtual IPv4    : %s\n", dht.DeriveVirtualIPv4(id.String()))
		fmt.Printf("[IP6] Virtual IPv6    : %s\n", dht.DeriveVirtualIPv6(id.String()))
		allowed, dropped, totalRules := fw.Stats()
		fmt.Printf("[FW]  Firewall ZTNA   : %d reglas activas (Permitidos: %d, Descartados: %d)\n", totalRules, allowed, dropped)
		fmt.Printf("[DNS] Petnames (dDNS) : %d alias registrados en libreta\n", len(petStore.List()))
		fmt.Println("==================================================================")

	case "petname":
		if len(os.Args) < 3 {
			fmt.Println("Uso: ipvn7-cli petname [list|add|resolve|delete|export-hosts]")
			return
		}
		sub := os.Args[2]
		switch sub {
		case "list":
			entries := petStore.List()
			if len(entries) == 0 {
				fmt.Println("No hay petnames registrados. Agrega uno con: ipvn7-cli petname add <nombre> <did>")
				return
			}
			fmt.Println("\n--- Libreta de Nombres Descentralizados (dDNS Petnames) ---")
			fmt.Printf("%-20s %-22s %s\n", "ALIAS", "IP VIRTUAL", "SOVEREIGN DID")
			fmt.Println(strings.Repeat("-", 80))
			for _, e := range entries {
				vIP := dht.DeriveVirtualIPv4(e.DID)
				fmt.Printf("%-20s %-22s %s\n", e.Name+".ipv7", vIP, e.DID)
			}
			fmt.Println()

		case "add":
			if len(os.Args) < 5 {
				fmt.Println("Uso: ipvn7-cli petname add <nombre> <did> [notas]")
				return
			}
			name := os.Args[3]
			did := os.Args[4]
			notes := ""
			if len(os.Args) >= 6 {
				notes = strings.Join(os.Args[5:], " ")
			}
			if err := petStore.Set(name, did, notes); err != nil {
				fmt.Printf("[ERROR] %v\n", err)
				return
			}
			fmt.Printf("[OK] Alias '%s.ipv7' registrado exitosamente hacia %s\n", dht.NormalizeName(name), did)

		case "resolve":
			if len(os.Args) < 4 {
				fmt.Println("Uso: ipvn7-cli petname resolve <nombre>")
				return
			}
			name := os.Args[3]
			did, ok := petStore.Resolve(name)
			if !ok {
				fmt.Printf("[!] Alias '%s' no encontrado en la libreta local.\n", name)
				return
			}
			fmt.Printf("\nResolución dDNS exitosa:\n")
			fmt.Printf("  Alias        : %s.ipv7\n", dht.NormalizeName(name))
			fmt.Printf("  Sovereign DID: %s\n", did)
			fmt.Printf("  Virtual IPv4 : %s\n", dht.DeriveVirtualIPv4(did))
			fmt.Printf("  Virtual IPv6 : %s\n\n", dht.DeriveVirtualIPv6(did))

		case "delete":
			if len(os.Args) < 4 {
				fmt.Println("Uso: ipvn7-cli petname delete <nombre>")
				return
			}
			name := os.Args[3]
			if petStore.Delete(name) {
				fmt.Printf("[OK] Alias '%s' eliminado.\n", name)
			} else {
				fmt.Printf("[!] Alias '%s' no encontrado.\n", name)
			}

		case "export-hosts":
			fmt.Print(petStore.ExportHostsFile(false))
		}

	case "firewall":
		if len(os.Args) < 3 {
			fmt.Println("Uso: ipvn7-cli firewall [list|add|check]")
			return
		}
		sub := os.Args[2]
		switch sub {
		case "list":
			rules := fw.ListRules()
			if len(rules) == 0 {
				fmt.Println("Firewall ZTNA activo. Política por defecto: DENY (0 reglas explícitas).")
				return
			}
			fmt.Println("\n--- Reglas de Micro-segmentación ZTNA Activas ---")
			fmt.Printf("%-18s %-10s %-8s %-10s %s\n", "ID REGLA", "PROTO", "PUERTO", "ACCIÓN", "SOURCE DID")
			fmt.Println(strings.Repeat("-", 75))
			for _, r := range rules {
				portStr := fmt.Sprintf("%d", r.DestPort)
				if r.DestPort == 0 {
					portStr = "*"
				}
				fmt.Printf("%-18s %-10s %-8s %-10s %s\n", r.ID, strings.ToUpper(r.Protocol), portStr, r.Action, r.SourceDID)
			}
			fmt.Println()

		case "add":
			if len(os.Args) < 6 {
				fmt.Println("Uso: ipvn7-cli firewall add <source_did|*> <puerto|0> <ALLOW|DENY> [descripcion]")
				return
			}
			srcDID := os.Args[3]
			portNum, err := strconv.ParseUint(os.Args[4], 10, 16)
			if err != nil {
				fmt.Println("[ERROR] Puerto inválido (0 para cualquier puerto)")
				return
			}
			actionStr := strings.ToUpper(os.Args[5])
			action := adapters.ActionDeny
			if actionStr == "ALLOW" {
				action = adapters.ActionAllow
			}
			desc := "Configurada via ipvn7-cli"
			if len(os.Args) >= 7 {
				desc = strings.Join(os.Args[6:], " ")
			}

			rule := adapters.FirewallRule{
				SourceDID:   srcDID,
				Protocol:    "*",
				DestPort:    uint16(portNum),
				Action:      action,
				Description: desc,
			}
			if err := fw.AddRule(rule); err != nil {
				fmt.Printf("[ERROR] No se pudo agregar la regla: %v\n", err)
				return
			}
			_ = fw.SaveToFile(firewallPath)
			fmt.Printf("[OK] Regla ZTNA guardada: %s tráfico desde '%s' hacia puerto %d\n", action, srcDID, portNum)

		case "check":
			if len(os.Args) < 5 {
				fmt.Println("Uso: ipvn7-cli firewall check <source_did> <puerto>")
				return
			}
			srcDID := os.Args[3]
			portNum, _ := strconv.ParseUint(os.Args[4], 10, 16)
			allowed, action, reason := fw.Inspect(srcDID, "tcp", uint16(portNum))
			fmt.Println("--- Evaluación ZTNA ---")
			fmt.Printf("  DID Origen : %s\n", srcDID)
			fmt.Printf("  Puerto Dest: %d\n", portNum)
			fmt.Printf("  Decisión   : %s (Permitido: %v)\n", action, allowed)
			fmt.Printf("  Motivo     : %s\n", reason)
		}

	case "kuzu":
		if len(os.Args) < 3 {
			fmt.Println("Uso: ipvn7-cli kuzu \"MATCH (n) RETURN n LIMIT 5;\"")
			return
		}
		query := os.Args[2]
		cmd := exec.Command("python", "tools/indexer/audit_kuzu.py")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("[*] Consultando grafo Kùzu para: %s\n\n", query)
		_ = cmd.Run()

	default:
		fmt.Printf("Comando desconocido: %s\n", command)
		printHelp()
	}
}
