package zfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"
)

func (z *zfs) listByType(t, filter string) ([]*Dataset, error) {
	args := []string{"list", "-rp", "-t", t, "-o", "all"}

	if filter != "" {
		args = append(args, filter)
	}
	out, err := z.doOutput(args...)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	var datasets []*Dataset

	name := ""
	var ds *Dataset
	for _, line := range out[1:] {
		if name != line[0] {
			name = line[0]
			ds = &Dataset{z: z, Name: name, props: make(map[string]string)}
			datasets = append(datasets, ds)
		}
		if err := ds.parseProps([][]string{out[0], line}); err != nil {
			return nil, err
		}
	}

	return datasets, nil
}

func (d *Dataset) parseProps(out [][]string) error {
	var err error

	if len(out) != 2 {
		return errors.New("output does not match what is expected on this platform")
	}
	for i, v := range out[0] {
		val := "-"
		if i < len(out[1]) {
			val = out[1][i]
		}
		d.props[strings.ToLower(v)] = val
	}

	if len(d.props) <= len(dsPropList) {
		return errors.New("output does not match what is expected on this platform")
	}
	setString(&d.Name, d.props["name"])
	setString(&d.Origin, d.props["origin"])

	if err = setUint(&d.Used, d.props["used"]); err != nil {
		// return err
		return fmt.Errorf("failed to parse used: %w", err)
	}
	if err = setUint(&d.Avail, d.props["avail"]); err != nil {
		return fmt.Errorf("failed to parse avail: %w", err)
	}

	setString(&d.Mountpoint, d.props["mountpoint"])
	setString(&d.Compression, d.props["compress"])
	setString(&d.Type, d.props["type"])

	if err = setUint(&d.Volsize, d.props["volsize"]); err != nil {
		return fmt.Errorf("failed to parse volsize: %w", err)
	}

	if d.props["volblock"] != "" && d.props["volblock"] != "-" {
		if err = setUint(&d.VolBlockSize, d.props["volblock"]); err != nil {
			return fmt.Errorf("failed to parse volblock: %w", err)
		}
	}

	if err = setUint(&d.Quota, d.props["quota"]); err != nil {
		return fmt.Errorf("failed to parse quota: %w", err)
	}
	if err = setUint(&d.Referenced, d.props["refer"]); err != nil {
		return fmt.Errorf("failed to parse refer: %w", err)
	}

	if runtime.GOOS == "solaris" {
		return nil
	}

	if err = setUint(&d.Written, d.props["written"]); err != nil {
		return fmt.Errorf("failed to parse written: %w", err)
	}
	if err = setUint(&d.Logicalused, d.props["lused"]); err != nil {
		return fmt.Errorf("failed to parse lused: %w", err)
	}
	if err = setUint(&d.Usedbydataset, d.props["usedds"]); err != nil {
		return fmt.Errorf("failed to parse usedds: %w", err)
	}
	return nil
}

func (z *zfs) run(in io.Reader, out io.Writer, cmd string, args ...string) ([][]string, error) {
	var stdout, stderr bytes.Buffer

	if z.sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}

	cmdOut := out

	if cmdOut == nil {
		cmdOut = &stdout
	}

	joinedArgs := strings.Join(args, " ")

	if err := z.exec.Run(in, cmdOut, &stderr, cmd, args...); err != nil {
		return nil, &Error{
			Err:    err,
			Debug:  strings.Join([]string{cmd, joinedArgs}, " "),
			Stderr: stderr.String(),
		}
	}

	if out != nil {
		return nil, nil
	}

	lines := strings.Split(stdout.String(), "\n")
	lines = lines[0 : len(lines)-1]

	output := make([][]string, len(lines))

	for i, l := range lines {
		output[i] = strings.Fields(l)
	}

	return output, nil
}

// https://docs.oracle.com/cd/E26505_01/html/E37384/gbcpt.html
func IsValidPoolName(name string) bool {
	// Must start with a letter and contain only alphanumeric characters, '_', '-', or '.'
	validNamePattern := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]*$`)
	if !validNamePattern.MatchString(name) {
		return false
	}

	if len(name) >= 2 && name[0] == 'c' && name[1] >= '0' && name[1] <= '9' {
		return false
	}

	reservedNames := map[string]bool{
		"log":    true,
		"mirror": true,
		"raidz":  true,
		"raidz1": true,
		"raidz2": true,
		"raidz3": true,
		"spare":  true,
	}

	lowerName := strings.ToLower(name)
	if reservedNames[lowerName] {
		return false
	}

	for reserved := range reservedNames {
		if strings.HasPrefix(lowerName, reserved) {
			return false
		}
	}

	if strings.Contains(name, "%") {
		return false
	}

	return true
}
