package ptp

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type TimeStatus struct {
	ClockID      string
	MasterOffset int64
	IngressTime  uint64
	GmPresent    bool
	GmIdentity   string
}

type Checker struct {
	// Remove syncThresholdNs - clients will decide sync status
}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) GetLocalTimeStatus() (*TimeStatus, error) {
	cmd := exec.Command("pmc", "-u", "-b", "0", "GET TIME_STATUS_NP")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute pmc: %v", err)
	}
	
	return c.parseTimeStatus(string(output))
}

func (c *Checker) GetRemoteTimeStatus(targetIP string) (*TimeStatus, error) {
	// Ping to ensure ARP entry exists
	pingCmd := exec.Command("ping", "-c", "1", targetIP)
	pingCmd.Run() // Ignore error as ping might fail but ARP entry could still exist
	
	// Get MAC address and interface from ARP table
	macAddr, iface, err := getMACAndInterfaceFromARP(targetIP)
	if err != nil {
		return nil, fmt.Errorf("failed to get MAC address: %v", err)
	}
	
	// Convert MAC to PTP clockIdentity
	clockID := macToClockID(macAddr)
	
	// Get remote time status
	cmd := exec.Command("pmc", "-b", "0", "GET TIME_STATUS_NP", "-i", iface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute pmc: %v", err)
	}
	
	// Filter output for specific clockID
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var capturing bool
	var result strings.Builder
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Check if this is the start of a response block
		if strings.Contains(line, ".fffe.") {
			if strings.HasPrefix(strings.TrimSpace(line), clockID) {
				capturing = true
			} else {
				capturing = false
			}
		}
		
		if capturing {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	
	if result.Len() == 0 {
		return nil, fmt.Errorf("no response from device")
	}
	
	return c.parseTimeStatus(result.String())
}

func (c *Checker) parseTimeStatus(output string) (*TimeStatus, error) {
	status := &TimeStatus{}
	
	// Parse clockID from response line
	clockIDRe := regexp.MustCompile(`([0-9a-f]{6}\.fffe\.[0-9a-f]{6})-\d+`)
	if match := clockIDRe.FindStringSubmatch(output); len(match) > 1 {
		status.ClockID = match[1]
	}
	
	// Parse master_offset
	masterOffsetRe := regexp.MustCompile(`master_offset\s+(-?\d+)`)
	if match := masterOffsetRe.FindStringSubmatch(output); len(match) > 1 {
		offset, err := strconv.ParseInt(match[1], 10, 64)
		if err == nil {
			status.MasterOffset = offset
		}
	}
	
	// Parse ingress_time
	ingressTimeRe := regexp.MustCompile(`ingress_time\s+(\d+)`)
	if match := ingressTimeRe.FindStringSubmatch(output); len(match) > 1 {
		time, err := strconv.ParseUint(match[1], 10, 64)
		if err == nil {
			status.IngressTime = time
		}
	}
	
	// Parse gmPresent
	gmPresentRe := regexp.MustCompile(`gmPresent\s+(true|false)`)
	if match := gmPresentRe.FindStringSubmatch(output); len(match) > 1 {
		status.GmPresent = match[1] == "true"
	}
	
	// Parse gmIdentity
	gmIdentityRe := regexp.MustCompile(`gmIdentity\s+([0-9a-f]{6}\.fffe\.[0-9a-f]{6})`)
	if match := gmIdentityRe.FindStringSubmatch(output); len(match) > 1 {
		status.GmIdentity = match[1]
	}
	
	// Return raw data only - let clients decide sync status
	
	return status, nil
}

func getMACAndInterfaceFromARP(ip string) (string, string, error) {
	cmd := exec.Command("arp", "-n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to execute arp: %v", err)
	}
	
	scanner := bufio.NewScanner(bytes.NewReader(output))
	// Match IP address at start, then skip HWtype column, then capture MAC address and interface
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(ip) + `\s+\S+\s+([0-9a-fA-F:]{17})\s+\S+\s+(\S+)`)
	
	for scanner.Scan() {
		line := scanner.Text()
		if match := re.FindStringSubmatch(line); len(match) > 2 {
			return match[1], match[2], nil
		}
	}
	
	return "", "", fmt.Errorf("MAC address for IP %s not found in ARP table", ip)
}

func macToClockID(mac string) string {
	// Remove colons from MAC address
	cleanMAC := strings.ReplaceAll(mac, ":", "")
	cleanMAC = strings.ToLower(cleanMAC)
	
	// Format: first 6 chars + ".fffe." + last 6 chars
	if len(cleanMAC) != 12 {
		return ""
	}
	
	return cleanMAC[:6] + ".fffe." + cleanMAC[6:]
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}