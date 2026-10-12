package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type RuleItem struct {
	Name     string   `yaml:"name"`
	Behavior string   `yaml:"behavior"` // domain or ipcidr
	Sources  []string `yaml:"sources"`
}

type Config struct {
	Rules []RuleItem `yaml:"rules"`
}

type ClashPayload struct {
	Payload []string `yaml:"payload"`
}

func fetchBytes(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 45 * time.Second}
		req, err := http.NewRequest("GET", source, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mihomo-Ruleset-Builder")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}

func cleanLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
		return ""
	}

	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimSpace(line)

	prefixes := []string{
		"IP-CIDR,", "IP-CIDR6,", "IP-CIDR6 ", "IP-CIDR ",
		"DOMAIN-SUFFIX,", "DOMAIN,", "DOMAIN-KEYWORD,",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToUpper(line), p) {
			line = line[len(p):]
			break
		}
	}

	if idx := strings.Index(line, ","); idx != -1 {
		line = line[:idx]
	}

	line = strings.Trim(line, `"' `)
	return strings.TrimSpace(line)
}

func parseTextOrYaml(raw []byte) []string {
	var cp ClashPayload
	if err := yaml.Unmarshal(raw, &cp); err == nil && len(cp.Payload) > 0 {
		var res []string
		for _, item := range cp.Payload {
			cleaned := cleanLine(item)
			if cleaned != "" {
				res = append(res, cleaned)
			}
		}
		return res
	}

	var res []string
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		cleaned := cleanLine(scanner.Text())
		if cleaned != "" {
			res = append(res, cleaned)
		}
	}
	return res
}

func parseMrs(mihomoPath, behavior, source, tempDir string) ([]string, error) {
	raw, err := fetchBytes(source)
	if err != nil {
		return nil, err
	}

	tempMrs := filepath.Join(tempDir, fmt.Sprintf("in-%d.mrs", time.Now().UnixNano()))
	tempTxt := filepath.Join(tempDir, fmt.Sprintf("out-%d.txt", time.Now().UnixNano()))
	defer os.Remove(tempMrs)
	defer os.Remove(tempTxt)

	if err := os.WriteFile(tempMrs, raw, 0644); err != nil {
		return nil, err
	}

	cmd := exec.Command(mihomoPath, "convert-ruleset", behavior, "mrs", tempMrs, tempTxt)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("mihomo unpack failed: %s (%v)", string(out), err)
	}

	txtBytes, err := os.ReadFile(tempTxt)
	if err != nil {
		return nil, err
	}

	return parseTextOrYaml(txtBytes), nil
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	outDir := flag.String("out", "output", "Output directory")
	mihomoBin := flag.String("mihomo", "./mihomo", "Path to mihomo binary")
	buildMrs := flag.Bool("build-mrs", false, "Compile output lists into .mrs format using mihomo")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Config read error: %v\n", err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("Config parse error: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Printf("Cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	tempDir := filepath.Join(*outDir, ".tmp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	for _, rule := range cfg.Rules {
		baseName := fmt.Sprintf("%s@%s", rule.Name, rule.Behavior)
		listPath := filepath.Join(*outDir, baseName+".list")
		mrsPath := filepath.Join(*outDir, baseName+".mrs")

		fmt.Printf("==> Processing: %s\n", baseName)

		mergedMap := make(map[string]struct{})

		for _, src := range rule.Sources {
			var items []string
			var err error

			if strings.HasSuffix(strings.ToLower(src), ".mrs") {
				items, err = parseMrs(*mihomoBin, rule.Behavior, src, tempDir)
			} else {
				var raw []byte
				raw, err = fetchBytes(src)
				if err == nil {
					items = parseTextOrYaml(raw)
				}
			}

			if err != nil {
				fmt.Printf("  [WARN] Failed %s: %v\n", src, err)
				continue
			}

			for _, item := range items {
				mergedMap[item] = struct{}{}
			}
		}

		if len(mergedMap) == 0 {
			fmt.Printf("  [SKIP] List is empty: %s\n", baseName)
			continue
		}

		uniqueList := make([]string, 0, len(mergedMap))
		for k := range mergedMap {
			uniqueList = append(uniqueList, k)
		}
		sort.Strings(uniqueList)

		content := strings.Join(uniqueList, "\n") + "\n"
		if err := os.WriteFile(listPath, []byte(content), 0644); err != nil {
			fmt.Printf("  [ERROR] Write error for %s: %v\n", listPath, err)
			continue
		}
		fmt.Printf("  [OK] Generated text: %s (%d records)\n", listPath, len(uniqueList))

		// Compilation into .mrs if the flag is set
		if *buildMrs {
			cmd := exec.Command(*mihomoBin, "convert-ruleset", rule.Behavior, "text", listPath, mrsPath)
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("  [ERROR] Failed to compile %s: %s\n", mrsPath, string(out))
			} else {
				fmt.Printf("  [OK] Compiled binary: %s\n", mrsPath)
			}
		}
	}
}
