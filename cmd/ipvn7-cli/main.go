package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	fmt.Println("  accounting list                Muestra el balance Tit-for-Tat y tiers de los peers")
	fmt.Println("  accounting status <did>        Muestra el consumo y reciprocidad de un peer")
	fmt.Println("  wot list                       Lista los avales emitidos en la red de confianza")
	fmt.Println("  wot vouch <did> <score> [tag]  Emite y firma digitalmente un aval de confianza")
	fmt.Println("  wot score <target_did>         Calcula la confianza transitiva hacia un nodo")
	fmt.Println("  wot export-kuzu                Exporta el grafo de confianza a sentencias Cypher")
	fmt.Println("  qos check <did> <bytes>        Evalúa si un paquete es admitido por el Token Bucket")
	fmt.Println("  qos challenge <did>            Genera un desafío Proof-of-Work anti-DDoS")
	fmt.Println("  qos demo                       Demuestra la resolución automática de PoW y crédito")
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

	case "accounting":
		acctPath := filepath.Join(ipv7Dir, "transit_accounting.json")
		acctEngine := adapters.NewAccountingEngine(acctPath, 10*1024*1024)

		if len(os.Args) < 3 {
			fmt.Println("Uso: ipvn7-cli accounting [list|record|status]")
			return
		}
		sub := os.Args[2]
		switch sub {
		case "list":
			accounts := acctEngine.ListAccounts()
			if len(accounts) == 0 {
				fmt.Println("No hay cuentas de tránsito registradas. Se crearán automáticamente durante el enrutamiento.")
				return
			}
			fmt.Println("\n--- Reciprocidad de Tránsito Tit-for-Tat ---")
			fmt.Printf("%-24s %-12s %-12s %-12s %-10s %s\n", "SOVEREIGN DID", "RELAYED FOR", "RELAYED BY", "BALANCE", "RATIO", "TIER")
			fmt.Println(strings.Repeat("-", 85))
			for _, a := range accounts {
				shortDID := a.DID
				if len(shortDID) > 22 {
					shortDID = shortDID[:22] + "..."
				}
				fmt.Printf("%-24s %-12s %-12s %-12d %-10.2f %s\n",
					shortDID,
					adapters.FormatTraffic(a.BytesRelayedFor),
					adapters.FormatTraffic(a.BytesRelayedBy),
					a.CreditBalance,
					a.ReciprocityRatio,
					a.Tier,
				)
			}
			fmt.Println()

		case "record":
			if len(os.Args) < 6 {
				fmt.Println("Uso: ipvn7-cli accounting record <peer_did> <bytes_relayed_for> <bytes_relayed_by>")
				return
			}
			peerDID := os.Args[3]
			bytesFor, err1 := strconv.ParseUint(os.Args[4], 10, 64)
			bytesBy, err2 := strconv.ParseUint(os.Args[5], 10, 64)
			if err1 != nil || err2 != nil {
				fmt.Println("[ERROR] Valores numéricos de bytes inválidos.")
				return
			}
			tier := acctEngine.RecordTransit(peerDID, bytesFor, bytesBy)
			_ = acctEngine.SaveToFile(acctPath)
			fmt.Printf("[OK] Tránsito registrado para %s. Nuevo Tier: %s\n", peerDID, tier)

		case "status":
			if len(os.Args) < 4 {
				fmt.Println("Uso: ipvn7-cli accounting status <peer_did>")
				return
			}
			peerDID := os.Args[3]
			acc, found := acctEngine.GetAccount(peerDID)
			if !found {
				fmt.Printf("[!] Peer %s no encontrado en el registro contable.\n", peerDID)
				return
			}
			fmt.Println("\n--- Estado Contable del Peer ---")
			fmt.Printf("  DID          : %s\n", acc.DID)
			fmt.Printf("  Relayed For  : %s\n", adapters.FormatTraffic(acc.BytesRelayedFor))
			fmt.Printf("  Relayed By   : %s\n", adapters.FormatTraffic(acc.BytesRelayedBy))
			fmt.Printf("  Credit Balance: %d\n", acc.CreditBalance)
			fmt.Printf("  Ratio Tit-Tat: %.2f\n", acc.ReciprocityRatio)
			fmt.Printf("  QoS Tier     : %s\n\n", acc.Tier)
		}

	case "wot":
		keyPath := filepath.Join(ipv7Dir, "identity.key")
		id, priv, err := core.LoadOrCreatePersistentIdentity(keyPath)
		if err != nil {
			fmt.Printf("[ERROR] No se pudo cargar la identidad local: %v\n", err)
			return
		}
		localDID := "did:ipv7:" + id.String()
		wotPath := filepath.Join(ipv7Dir, "wot.json")
		wotEngine := dht.NewWebOfTrust(localDID, wotPath)

		if len(os.Args) < 3 {
			fmt.Println("Uso: ipvn7-cli wot [list|vouch|score|export-kuzu]")
			return
		}
		sub := os.Args[2]
		switch sub {
		case "list":
			vouches := wotEngine.ListVouchesFrom(localDID)
			if len(vouches) == 0 {
				fmt.Println("No has emitido avales de confianza. Usa: ipvn7-cli wot vouch <subject_did> <score> [tag]")
				return
			}
			fmt.Println("\n--- Avales de Confianza Emitidos (Web of Trust) ---")
			fmt.Printf("%-24s %-8s %-16s %s\n", "SUBJECT DID", "SCORE", "TAG", "FECHA")
			fmt.Println(strings.Repeat("-", 70))
			for _, v := range vouches {
				shortDID := v.SubjectDID
				if len(shortDID) > 22 {
					shortDID = shortDID[:22] + "..."
				}
				fmt.Printf("%-24s %-8.2f %-16s %s\n", shortDID, v.Score, v.Tag, v.Timestamp.Format("2006-01-02 15:04"))
			}
			fmt.Println()

		case "vouch":
			if len(os.Args) < 5 {
				fmt.Println("Uso: ipvn7-cli wot vouch <subject_did> <score: 0.0 a 1.0> [tag] [comment]")
				return
			}
			subjectDID := os.Args[3]
			score, err := strconv.ParseFloat(os.Args[4], 64)
			if err != nil || score < 0.0 || score > 1.0 {
				fmt.Println("[ERROR] El score de confianza debe ser un número entre 0.0 y 1.0")
				return
			}
			tag := "endorsed-peer"
			if len(os.Args) >= 6 {
				tag = os.Args[5]
			}
			comment := ""
			if len(os.Args) >= 7 {
				comment = strings.Join(os.Args[6:], " ")
			}

			vouch := dht.TrustVouch{
				IssuerDID:  localDID,
				SubjectDID: subjectDID,
				Score:      score,
				Tag:        tag,
				Comment:    comment,
				Timestamp:  time.Now(),
			}
			vouch.Sign(priv)
			if err := wotEngine.AddVouch(vouch, true); err != nil {
				fmt.Printf("[ERROR] %v\n", err)
				return
			}
			_ = wotEngine.SaveToFile(wotPath)
			fmt.Printf("[OK] Aval firmado digitalmente (Ed25519) registrado para %s con score %.2f (%s)\n", subjectDID, score, tag)

		case "score":
			if len(os.Args) < 4 {
				fmt.Println("Uso: ipvn7-cli wot score <target_did>")
				return
			}
			targetDID := os.Args[3]
			score := wotEngine.CalculateTrust(localDID, targetDID, 3)
			fmt.Println("\n--- Cálculo de Confianza P2P (Web-of-Trust) ---")
			fmt.Printf("  Nodo Origen (Tú)  : %s\n", localDID)
			fmt.Printf("  Nodo Destino      : %s\n", targetDID)
			fmt.Printf("  Score Transitivo  : %.4f (Escala 0.0 a 1.0)\n\n", score)

		case "export-kuzu":
			stmts := wotEngine.ExportKuzuCypher()
			if len(stmts) == 0 {
				fmt.Println("// No hay relaciones de confianza para exportar")
				return
			}
			fmt.Println("// Sentencias Cypher para Kùzu Graph Engine:")
			for _, s := range stmts {
				fmt.Println(s)
			}
		}

	case "qos":
		qos := adapters.NewHierarchicalQoS()
		if len(os.Args) < 3 {
			fmt.Println("Uso: ipvn7-cli qos [check|challenge|demo]")
			return
		}
		sub := os.Args[2]
		switch sub {
		case "check":
			if len(os.Args) < 5 {
				fmt.Println("Uso: ipvn7-cli qos check <did> <bytes>")
				return
			}
			did := os.Args[3]
			bytes, _ := strconv.Atoi(os.Args[4])
			admitted, ch := qos.IngressCheck(did, adapters.PriorityInteractive, bytes)
			fmt.Println("\n--- Evaluación de Caudal QoS Token Bucket ---")
			fmt.Printf("  Target DID  : %s\n", did)
			fmt.Printf("  Packet Size : %d bytes\n", bytes)
			fmt.Printf("  Admitido    : %v\n", admitted)
			if !admitted && ch != nil {
				fmt.Printf("  [!] Tasa excedida. Desafío PoW emitido: ID=%s (Dificultad: %d bits)\n", ch.ChallengeID, ch.Difficulty)
			}
			fmt.Println()

		case "challenge":
			if len(os.Args) < 4 {
				fmt.Println("Uso: ipvn7-cli qos challenge <did>")
				return
			}
			did := os.Args[3]
			_, ch := qos.IngressCheck(did, adapters.PriorityInteractive, 1000*1024)
			if ch != nil {
				fmt.Println("\n--- Desafío Anti-DDoS Proof-of-Work (PoW) ---")
				fmt.Printf("  Challenge ID: %s\n", ch.ChallengeID)
				fmt.Printf("  Target DID  : %s\n", ch.TargetDID)
				fmt.Printf("  Dificultad  : %d bits de ceros requeridos\n", ch.Difficulty)
				fmt.Printf("  Expira en   : %s\n\n", ch.ExpiresAt.Format("15:04:05"))
			}

		case "demo":
			did := "did:ipv7:test_peer_anti_ddos"
			fmt.Println("\n--- Demostración Dinámica Anti-DDoS PoW ---")
			fmt.Printf("[1] Agotando bucket interactivo (250 KB) para %s...\n", did)
			qos.IngressCheck(did, adapters.PriorityInteractive, 250*1024)
			admitted, ch := qos.IngressCheck(did, adapters.PriorityInteractive, 1024)
			if !admitted && ch != nil {
				fmt.Printf("[2] Desafío emitido: ID=%s (Dificultad %d bits)\n", ch.ChallengeID, ch.Difficulty)
				start := time.Now()
				nonce := adapters.SolveChallenge(ch)
				elapsed := time.Since(start)
				fmt.Printf("[3] Desafío resuelto por CPU en %v! Nonce ganador: %d\n", elapsed, nonce)
				ok := qos.VerifyAndCredit(ch.ChallengeID, nonce)
				fmt.Printf("[4] Verificación criptográfica en nodo: %v (Tokens acreditados)\n", ok)
				admitted, _ := qos.IngressCheck(did, adapters.PriorityInteractive, 1024)
				fmt.Printf("[5] Nuevo paquete tras resolver PoW: Admitido=%v\n\n", admitted)
			}
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
