package ptp

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
	// Get MAC address and interface from /proc/net/arp
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
	// Read /proc/net/arp directly
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return "", "", fmt.Errorf("failed to open /proc/net/arp: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	// Skip header line
	if scanner.Scan() {
		// Header: IP address       HW type     Flags       HW address            Mask     Device
	}
	
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		
		// Expected format: IP address, HW type, Flags, HW address, Mask, Device
		if len(fields) >= 6 {
			if fields[0] == ip {
				mac := fields[3]
				device := fields[5]
				
				// Check if MAC is valid (not 00:00:00:00:00:00)
				if mac != "00:00:00:00:00:00" {
					return mac, device, nil
				}
				return "", "", fmt.Errorf("incomplete ARP entry for IP %s (MAC is 00:00:00:00:00:00)", ip)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading /proc/net/arp: %v", err)
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