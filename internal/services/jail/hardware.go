package jail

import (
	"fmt"
	"sort"
	"strings"
	jailModels "sylve/internal/db/models/jail"
	"sylve/pkg/utils"

	cpuid "github.com/klauspost/cpuid/v2"
)

func (s *Service) UpdateMemory(ctId uint, memoryBytes int64) error {
	if memoryBytes < 0 {
		return fmt.Errorf("invalid memory value: %d", memoryBytes)
	}

	const MiB = int64(1024 * 1024)
	mb := (memoryBytes + MiB - 1) / MiB
	if mb < 1 {
		return fmt.Errorf("memory must be at least 1MB, got: %dMB", mb)
	}

	cfg, err := s.GetJailConfig(ctId)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg) == "" {
		return fmt.Errorf("jail config not found for CTID: %d", ctId)
	}

	ctIdHash := utils.HashIntToNLetters(int(ctId), 5)
	prefix := fmt.Sprintf(`exec.poststart += "rctl -a jail:%s:memoryuse:deny=`, ctIdHash)

	lines := strings.Split(cfg, "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			lines[i] = fmt.Sprintf(`	exec.poststart += "rctl -a jail:%s:memoryuse:deny=%dM";`, ctIdHash, mb)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf(`exec.poststart += "rctl -a jail:%s:memoryuse:deny=%dM";`, ctIdHash, mb))
	}

	newCfg := strings.Join(lines, "\n")
	if err := s.SaveJailConfig(ctId, newCfg); err != nil {
		return fmt.Errorf("failed to save jail config: %w", err)
	}

	var jail jailModels.Jail
	if err := s.DB.Find(&jail, "ct_id = ?", ctId).Error; err != nil {
		return fmt.Errorf("failed to find jail with CTID %d: %w", ctId, err)
	}

	jail.Memory = int(memoryBytes)

	if err := s.DB.Save(&jail).Error; err != nil {
		return fmt.Errorf("failed to update jail memory in database: %w", err)
	}

	_, err = utils.RunCommand("rctl", "-a", fmt.Sprintf("jail:%s:memoryuse:deny=%dM", ctIdHash, mb))

	if err != nil {
		return fmt.Errorf("failed to apply memory limit with rctl: %w", err)
	}

	return nil
}

func (s *Service) UpdateCPU(ctId uint, cores int64) error {
	if cores <= 0 {
		return fmt.Errorf("invalid cores value: %d (must be >= 1)", cores)
	}

	numLogical := int64(cpuid.CPU.LogicalCores)

	if cores > numLogical {
		return fmt.Errorf("requested cores (%d) exceed logical cores available (%d)", cores, numLogical)
	}

	cfg, err := s.GetJailConfig(ctId)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg) == "" {
		return fmt.Errorf("jail config not found for CTID: %d", ctId)
	}

	var currentJails []jailModels.Jail
	if err := s.DB.Find(&currentJails).Error; err != nil {
		return fmt.Errorf("failed_to_fetch_current_jails: %w", err)
	}

	coreUsage := map[int]int{}
	for _, j := range currentJails {
		if j.CTID == int(ctId) {
			continue
		}
		for _, c := range j.CPUSet {
			coreUsage[c]++
		}
	}

	type coreCount struct {
		Core  int
		Count int
	}

	all := make([]coreCount, 0, cpuid.CPU.LogicalCores)
	for i := 0; i < cpuid.CPU.LogicalCores; i++ {
		all = append(all, coreCount{Core: i, Count: coreUsage[i]})
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Count < all[j].Count })

	selected := make([]int, 0, cores)
	for i := 0; i < int(cores) && i < len(all); i++ {
		selected = append(selected, all[i].Core)
	}

	if len(selected) == 0 {
		return fmt.Errorf("no CPU cores selected")
	}

	coreListStr := strings.Trim(strings.Replace(fmt.Sprint(selected), " ", ",", -1), "[]")
	ctIdHash := utils.HashIntToNLetters(int(ctId), 5)
	wantLine := fmt.Sprintf(`	exec.created += "cpuset -l %s -j %s";`, coreListStr, ctIdHash)
	lines := strings.Split(cfg, "\n")
	found := false

	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, `exec.created += "cpuset -l`) && strings.HasSuffix(t, fmt.Sprintf(`-j %s";`, ctIdHash)) {
			lines[i] = wantLine
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, wantLine)
	}
	newCfg := strings.Join(lines, "\n")

	if err := s.SaveJailConfig(ctId, newCfg); err != nil {
		return fmt.Errorf("failed to save jail config: %w", err)
	}

	var jail jailModels.Jail
	if err := s.DB.Find(&jail, "ct_id = ?", ctId).Error; err != nil {
		return fmt.Errorf("failed to find jail with CTID %d: %w", ctId, err)
	}

	jail.Cores = int(cores)
	jail.CPUSet = selected

	if err := s.DB.Save(&jail).Error; err != nil {
		return fmt.Errorf("failed to update jail CPU in database: %w", err)
	}

	_, err = utils.RunCommand("cpuset", "-l", coreListStr, "-j", ctIdHash)
	if err != nil {
		return fmt.Errorf("failed to apply CPU set with cpuset: %w", err)
	}

	return nil
}
